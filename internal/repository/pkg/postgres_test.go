package pkg

import (
	"context"
	"database/sql"
	"repub/internal/db/postgres"
	"repub/internal/domain"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// Mock queries for testing
type mockQueries struct {
	packages  map[string]*postgres.Package
	versions  map[int32][]*postgres.PackageVersion
	uploaders map[int32][]string
}

func newMockQueries() *mockQueries {
	return &mockQueries{
		packages:  make(map[string]*postgres.Package),
		versions:  make(map[int32][]*postgres.PackageVersion),
		uploaders: make(map[int32][]string),
	}
}

func (m *mockQueries) GetPackage(ctx context.Context, name string) (postgres.Package, error) {
	pkg, exists := m.packages[name]
	if !exists {
		return postgres.Package{}, sql.ErrNoRows
	}
	return *pkg, nil
}

func (m *mockQueries) CreatePackage(ctx context.Context, params postgres.CreatePackageParams) (postgres.Package, error) {
	pkg := postgres.Package{
		ID:        int32(len(m.packages) + 1),
		Name:      params.Name,
		Private:   params.Private,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.packages[params.Name] = &pkg
	return pkg, nil
}

func (m *mockQueries) ListPackages(ctx context.Context, params postgres.ListPackagesParams) ([]postgres.Package, error) {
	var result []postgres.Package
	for _, pkg := range m.packages {
		result = append(result, *pkg)
	}
	return result, nil
}

func (m *mockQueries) GetPackageVersions(ctx context.Context, packageID int32) ([]postgres.PackageVersion, error) {
	versions := m.versions[packageID]
	var result []postgres.PackageVersion
	for _, v := range versions {
		result = append(result, *v)
	}
	return result, nil
}

func (m *mockQueries) GetLatestPackageVersion(ctx context.Context, packageID int32) (postgres.PackageVersion, error) {
	versions := m.versions[packageID]
	if len(versions) == 0 {
		return postgres.PackageVersion{}, sql.ErrNoRows
	}
	return *versions[0], nil
}

func (m *mockQueries) CreatePackageVersion(ctx context.Context, params postgres.CreatePackageVersionParams) (postgres.PackageVersion, error) {
	version := postgres.PackageVersion{
		ID:            int32(len(m.versions) + 1),
		PackageID:     params.PackageID,
		Version:       params.Version,
		Description:   params.Description,
		PubspecYaml:   params.PubspecYaml,
		ArchivePath:   params.ArchivePath,
		ArchiveSha256: params.ArchiveSha256,
		Uploader:      params.Uploader,
		Retracted:     false,
		CreatedAt:     time.Now(),
	}

	m.versions[params.PackageID] = append([]*postgres.PackageVersion{&version}, m.versions[params.PackageID]...)
	return version, nil
}

func (m *mockQueries) GetPackageUploaders(ctx context.Context, packageID int32) ([]string, error) {
	return m.uploaders[packageID], nil
}

func (m *mockQueries) AddPackageUploader(ctx context.Context, params postgres.AddPackageUploaderParams) error {
	m.uploaders[params.PackageID] = append(m.uploaders[params.PackageID], params.Uploader)
	return nil
}

func TestPostgresPackageRepository_GetPackage(t *testing.T) {
	queries := newMockQueries()
	repo := NewPostgresPackageRepository(queries)

	// Test non-existent package
	pkg, err := repo.GetPackage(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Expected no error for non-existent package, got %v", err)
	}
	if pkg != nil {
		t.Error("Expected nil for non-existent package")
	}

	// Create test package
	_, err = queries.CreatePackage(context.Background(), postgres.CreatePackageParams{
		Name:    "testpkg",
		Private: false,
	})
	if err != nil {
		t.Fatalf("Failed to create test package: %v", err)
	}

	// Test existing package
	pkg, err = repo.GetPackage(context.Background(), "testpkg")
	if err != nil {
		t.Fatalf("GetPackage failed: %v", err)
	}
	if pkg == nil {
		t.Fatal("Expected package, got nil")
	}
	if pkg.Name != "testpkg" {
		t.Errorf("Expected name 'testpkg', got %s", pkg.Name)
	}
}

func TestPostgresPackageRepository_CreatePackage(t *testing.T) {
	queries := newMockQueries()
	repo := NewPostgresPackageRepository(queries)

	pkg, err := repo.CreatePackage(context.Background(), "newpkg", true)
	if err != nil {
		t.Fatalf("CreatePackage failed: %v", err)
	}

	if pkg.Name != "newpkg" {
		t.Errorf("Expected name 'newpkg', got %s", pkg.Name)
	}
	if pkg.Private != true {
		t.Errorf("Expected private true, got %t", pkg.Private)
	}
}

func TestPostgresPackageRepository_ListPackages(t *testing.T) {
	queries := newMockQueries()
	repo := NewPostgresPackageRepository(queries)

	// Create test packages
	_, err := queries.CreatePackage(context.Background(), postgres.CreatePackageParams{
		Name:    "pkg1",
		Private: false,
	})
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	_, err = queries.CreatePackage(context.Background(), postgres.CreatePackageParams{
		Name:    "pkg2",
		Private: true,
	})
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	packages, err := repo.ListPackages(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("ListPackages failed: %v", err)
	}

	if len(packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(packages))
	}
}

func TestPostgresPackageRepository_GetPackageVersions(t *testing.T) {
	queries := newMockQueries()
	repo := NewPostgresPackageRepository(queries)

	// Create test package
	pkg, err := queries.CreatePackage(context.Background(), postgres.CreatePackageParams{
		Name:    "testpkg",
		Private: false,
	})
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	// Create test version
	_, err = queries.CreatePackageVersion(context.Background(), postgres.CreatePackageVersionParams{
		PackageID:   pkg.ID,
		Version:     "1.0.0",
		PubspecYaml: "name: testpkg\nversion: 1.0.0",
		ArchivePath: "/storage/testpkg/1.0.0/archive.tar.gz",
	})
	if err != nil {
		t.Fatalf("Failed to create version: %v", err)
	}

	versions, err := repo.GetPackageVersions(context.Background(), pkg.ID)
	if err != nil {
		t.Fatalf("GetPackageVersions failed: %v", err)
	}

	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}

	if versions[0].Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", versions[0].Version)
	}
}

func TestPostgresPackageRepository_CreateVersion(t *testing.T) {
	queries := newMockQueries()
	repo := NewPostgresPackageRepository(queries)

	// Create test package
	pkg, err := queries.CreatePackage(context.Background(), postgres.CreatePackageParams{
		Name:    "testpkg",
		Private: false,
	})
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	desc := "Test description"
	sha := "abc123"
	uploader := "test@example.com"

	version := &domain.PackageVersion{
		PackageID:     pkg.ID,
		Version:       "1.0.0",
		Description:   &desc,
		PubspecYaml:   "name: testpkg\nversion: 1.0.0",
		ArchivePath:   "/storage/testpkg/1.0.0/archive.tar.gz",
		ArchiveSha256: &sha,
		Uploader:      &uploader,
	}

	created, err := repo.CreateVersion(context.Background(), version)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	if created.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", created.Version)
	}

	if created.Description == nil || *created.Description != desc {
		t.Errorf("Expected description '%s', got %v", desc, created.Description)
	}
}

func TestPostgresPackageRepository_GetLatestVersion(t *testing.T) {
	queries := newMockQueries()
	repo := NewPostgresPackageRepository(queries)

	// Test non-existent package
	version, err := repo.GetLatestVersion(context.Background(), 999)
	if err != nil {
		t.Fatalf("Expected no error for non-existent package, got %v", err)
	}
	if version != nil {
		t.Error("Expected nil for non-existent package")
	}

	// Create test package with version
	pkg, err := queries.CreatePackage(context.Background(), postgres.CreatePackageParams{
		Name:    "testpkg",
		Private: false,
	})
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	_, err = queries.CreatePackageVersion(context.Background(), postgres.CreatePackageVersionParams{
		PackageID:   pkg.ID,
		Version:     "1.0.0",
		PubspecYaml: "name: testpkg\nversion: 1.0.0",
		ArchivePath: "/storage/testpkg/1.0.0/archive.tar.gz",
	})
	if err != nil {
		t.Fatalf("Failed to create version: %v", err)
	}

	version, err = repo.GetLatestVersion(context.Background(), pkg.ID)
	if err != nil {
		t.Fatalf("GetLatestVersion failed: %v", err)
	}

	if version == nil {
		t.Fatal("Expected version, got nil")
	}

	if version.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", version.Version)
	}
}

func TestPostgresPackageRepository_Uploaders(t *testing.T) {
	queries := newMockQueries()
	repo := NewPostgresPackageRepository(queries)

	// Create test package
	pkg, err := queries.CreatePackage(context.Background(), postgres.CreatePackageParams{
		Name:    "testpkg",
		Private: false,
	})
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	// Add uploader
	err = repo.AddUploader(context.Background(), pkg.ID, "test@example.com")
	if err != nil {
		t.Fatalf("AddUploader failed: %v", err)
	}

	// Get uploaders
	uploaders, err := repo.GetUploaders(context.Background(), pkg.ID)
	if err != nil {
		t.Fatalf("GetUploaders failed: %v", err)
	}

	if len(uploaders) != 1 {
		t.Errorf("Expected 1 uploader, got %d", len(uploaders))
	}

	if uploaders[0] != "test@example.com" {
		t.Errorf("Expected uploader 'test@example.com', got %s", uploaders[0])
	}
}

func TestNullStringToPtr(t *testing.T) {
	// Test valid string
	validString := sql.NullString{String: "test", Valid: true}
	ptr := nullStringToPtr(validString)
	if ptr == nil {
		t.Error("Expected pointer for valid string")
	} else if *ptr != "test" {
		t.Errorf("Expected 'test', got %s", *ptr)
	}

	// Test null string
	nullString := sql.NullString{String: "", Valid: false}
	ptr = nullStringToPtr(nullString)
	if ptr != nil {
		t.Error("Expected nil for null string")
	}
}


