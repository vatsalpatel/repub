package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"repub/internal/domain"
	"repub/internal/repository/pkg"
	"repub/internal/repository/pubspec"
	"repub/internal/repository/storage"
	"slices"
	"strings"

	"github.com/goccy/go-json"
)

type PubService interface {
	GetPackage(ctx context.Context, name string) (*domain.PackageResponse, error)
	GetPackageDetail(ctx context.Context, name string) (*domain.PackageDetail, error)
	GetPackageVersion(ctx context.Context, name, version string) (*domain.VersionResponse, error)
	PublishPackage(ctx context.Context, req *domain.PublishRequest) (*domain.PublishResponse, error)
	ListPackages(ctx context.Context, page, size int) ([]*domain.Package, error)
	DownloadPackage(ctx context.Context, name, version string) ([]byte, error)
	GetAdvisories(ctx context.Context, name string) (*domain.AdvisoriesResponse, error)
}

type (
	PackageDependencies struct {
		Port    string
		Package pkg.Repository
		Storage storage.Repository
		Pubspec pubspec.Repository
	}
	packageService struct {
		PackageDependencies
	}
)

func NewPubService(deps PackageDependencies) PubService {
	return &packageService{
		PackageDependencies: deps,
	}
}

func (s *packageService) baseURL() string {
	return fmt.Sprintf("http://localhost:%s", s.Port)
}

func (s *packageService) GetPackage(ctx context.Context, name string) (*domain.PackageResponse, error) {
	pkg, err := s.Package.GetPackage(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}
	if pkg == nil {
		return nil, nil
	}

	versions, err := s.Package.GetPackageVersions(ctx, pkg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get package versions: %w", err)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("package has no versions")
	}

	// Convert to response format
	versionResponses := make([]domain.VersionResponse, len(versions))
	for i, v := range versions {
		resp, err := s.versionToResponseWithPackage(v, pkg.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to convert version response: %w", err)
		}
		versionResponses[i] = resp
	}

	latest, err := s.versionToResponseWithPackage(versions[0], pkg.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to convert latest version response: %w", err)
	}

	return &domain.PackageResponse{
		Name:     pkg.Name,
		Latest:   latest,
		Versions: versionResponses,
	}, nil
}

func (s *packageService) GetPackageDetail(ctx context.Context, name string) (*domain.PackageDetail, error) {
	pkg, err := s.Package.GetPackage(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}
	if pkg == nil {
		return nil, nil
	}

	versions, err := s.Package.GetPackageVersions(ctx, pkg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get package versions: %w", err)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("package has no versions")
	}

	return &domain.PackageDetail{
		Package:  pkg,
		Latest:   versions[0], // First is latest due to ORDER BY created_at DESC
		Versions: versions,
	}, nil
}

func (s *packageService) PublishPackage(ctx context.Context, req *domain.PublishRequest) (*domain.PublishResponse, error) {
	// 1. Extract and parse pubspec.yaml from archive
	pubspecContent, readme, changelog, err := s.extractFilesFromArchive(req.Archive)
	if err != nil {
		return nil, fmt.Errorf("failed to extract files from archive: %w", err)
	}

	// 2. Parse and validate pubspec
	pubspec, err := s.Pubspec.ParseYAML(ctx, pubspecContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pubspec.yaml: %w", err)
	}

	// 3. Get or create package
	pkg, err := s.Package.GetPackage(ctx, pubspec.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing package: %w", err)
	}

	if pkg == nil {
		// Create new package
		pkg, err = s.Package.CreatePackage(ctx, pubspec.Name, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create package: %w", err)
		}
	}

	// 4. Check if uploader is authorized (add them if first time)
	uploaders, err := s.Package.GetUploaders(ctx, pkg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get uploaders: %w", err)
	}

	// If no uploaders exist, add the current uploader
	if len(uploaders) == 0 {
		err = s.Package.AddUploader(ctx, pkg.ID, req.Uploader)
		if err != nil {
			return nil, fmt.Errorf("failed to add uploader: %w", err)
		}
	} else {
		// Check if uploader is authorized
		authorized := slices.Contains(uploaders, req.Uploader)
		if !authorized {
			return nil, fmt.Errorf("unauthorized to upload to package %s", pubspec.Name)
		}
	}

	// 5. Check if version already exists
	versions, err := s.Package.GetPackageVersions(ctx, pkg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get package versions: %w", err)
	}

	for _, v := range versions {
		if v.Version == pubspec.Version {
			return nil, fmt.Errorf("version %s already exists for package %s", pubspec.Version, pubspec.Name)
		}
	}

	// 6. Store archive file
	archivePath, err := s.Storage.Store(pubspec.Name, pubspec.Version, req.Archive)
	if err != nil {
		return nil, fmt.Errorf("failed to store archive: %w", err)
	}

	// 7. Calculate SHA256 hash
	sha256Hash := s.calculateSHA256(req.Archive)

	// 8. Create package version record
	version := &domain.PackageVersion{
		PackageID:     pkg.ID,
		Version:       pubspec.Version,
		Description:   &pubspec.Description,
		PubspecYaml:   pubspecContent,
		Readme:        readme,
		Changelog:     changelog,
		ArchivePath:   archivePath,
		ArchiveSha256: &sha256Hash,
		Uploader:      &req.Uploader,
	}

	createdVersion, err := s.Package.CreateVersion(ctx, version)
	if err != nil {
		// Clean up stored archive on failure
		_ = s.Storage.Delete(archivePath)
		return nil, fmt.Errorf("failed to create version record: %w", err)
	}

	return &domain.PublishResponse{
		URL: fmt.Sprintf("%s/packages/%s/versions/%s", s.baseURL(), pubspec.Name, createdVersion.Version),
		Fields: map[string]string{
			"package": pubspec.Name,
			"version": createdVersion.Version,
		},
	}, nil
}

func (s *packageService) ListPackages(ctx context.Context, page, size int) ([]*domain.Package, error) {
	offset := int32((page - 1) * size)
	limit := int32(size)

	packages, err := s.Package.ListPackages(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return packages, nil
}

func (s *packageService) versionToResponseWithPackage(v *domain.PackageVersion, packageName string) (domain.VersionResponse, error) {
	archiveURL := fmt.Sprintf("%s/packages/%s/versions/%s/download", s.baseURL(), packageName, v.Version)

	// Parse pubspec YAML to JSON
	parsed, err := s.Pubspec.ParseYAML(context.Background(), v.PubspecYaml)
	if err != nil {
		return domain.VersionResponse{}, err
	}

	jsonBytes, err := json.Marshal(parsed)
	if err != nil {
		return domain.VersionResponse{}, err
	}

	var pubspecJSON map[string]any
	if err := json.Unmarshal(jsonBytes, &pubspecJSON); err != nil {
		return domain.VersionResponse{}, err
	}

	return domain.VersionResponse{
		Version:       v.Version,
		Retracted:     v.Retracted,
		ArchiveURL:    archiveURL,
		ArchiveSha256: stringValue(v.ArchiveSha256),
		Pubspec:       pubspecJSON,
	}, nil
}

func (s *packageService) GetPackageVersion(ctx context.Context, name, version string) (*domain.VersionResponse, error) {
	pkg, err := s.Package.GetPackage(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}
	if pkg == nil {
		return nil, nil
	}

	versions, err := s.Package.GetPackageVersions(ctx, pkg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get package versions: %w", err)
	}

	for _, v := range versions {
		if v.Version == version {
			response, err := s.versionToResponseWithPackage(v, name)
			if err != nil {
				return nil, fmt.Errorf("failed to convert version response: %w", err)
			}
			return &response, nil
		}
	}

	return nil, nil // Version not found
}

func (s *packageService) DownloadPackage(ctx context.Context, name, version string) ([]byte, error) {
	pkg, err := s.Package.GetPackage(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}
	if pkg == nil {
		return nil, fmt.Errorf("package not found")
	}

	versions, err := s.Package.GetPackageVersions(ctx, pkg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get package versions: %w", err)
	}

	for _, v := range versions {
		if v.Version == version {
			// Get the archive from storage
			data, err := s.Storage.Get(v.ArchivePath)
			if err != nil {
				return nil, fmt.Errorf("failed to get archive: %w", err)
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("version not found")
}

func (s *packageService) GetAdvisories(ctx context.Context, name string) (*domain.AdvisoriesResponse, error) {
	// For now, return empty advisories
	// In a real implementation, this would query a security advisory database
	return &domain.AdvisoriesResponse{
		Advisories:        []domain.Advisory{},
		AdvisoriesUpdated: "2024-01-01T00:00:00Z",
	}, nil
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *packageService) extractFilesFromArchive(archiveData []byte) (pubspecContent string, readme *string, changelog *string, err error) {
	// Create a gzip reader
	gzReader, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gzReader.Close() }()

	// Create a tar reader
	tarReader := tar.NewReader(gzReader)

	var foundPubspec bool
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Get the file name relative to the package root
		fileName := strings.TrimPrefix(header.Name, "./")

		// Remove package name prefix if present (e.g., "package-1.0.0/pubspec.yaml" -> "pubspec.yaml")
		parts := strings.Split(fileName, "/")
		if len(parts) > 1 {
			fileName = strings.Join(parts[1:], "/")
		}

		switch strings.ToLower(fileName) {
		case "pubspec.yaml":
			// Only process root-level pubspec.yaml (no subdirectories)
			if !foundPubspec && !strings.Contains(fileName, "/") {
				content, err := io.ReadAll(tarReader)
				if err != nil {
					return "", nil, nil, fmt.Errorf("failed to read pubspec.yaml: %w", err)
				}
				pubspecContent = string(content)
				foundPubspec = true
			}

		case "readme.md":
			content, err := io.ReadAll(tarReader)
			if err != nil {
				return "", nil, nil, fmt.Errorf("failed to read README.md: %w", err)
			}
			readmeContent := string(content)
			readme = &readmeContent

		case "changelog.md":
			content, err := io.ReadAll(tarReader)
			if err != nil {
				return "", nil, nil, fmt.Errorf("failed to read CHANGELOG.md: %w", err)
			}
			changelogContent := string(content)
			changelog = &changelogContent
		}
	}

	if !foundPubspec {
		return "", nil, nil, fmt.Errorf("pubspec.yaml not found in archive")
	}

	return pubspecContent, readme, changelog, nil
}

func (s *packageService) calculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
