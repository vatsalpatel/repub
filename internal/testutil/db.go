package testutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"testing"

	"repub/internal/db/sqlite"
	"repub/internal/domain"
	"repub/internal/repository/pkg"
	"repub/internal/repository/pubspec"
	"repub/internal/repository/storage"

	_ "modernc.org/sqlite"
)

// TestDatabase provides a SQLite database for testing
type TestDatabase struct {
	DB      *sql.DB
	Queries *sqlite.Queries
	Repo    pkg.Repository
	cleanup func()
}

// SetupTestDatabase creates an in-memory SQLite database for testing
func SetupTestDatabase(t *testing.T) *TestDatabase {
	t.Helper()

	// Create in-memory SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}

	// Apply schema using helper
	schema, err := schemaFS.ReadFile("schema_sqlite.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Create queries and repository
	queries := sqlite.New(db)
	repo := newSQLitePackageRepository(queries)

	return &TestDatabase{
		DB:      db,
		Queries: queries,
		Repo:    repo,
		cleanup: func() { _ = db.Close() },
	}
}

// Close closes the test database
func (tdb *TestDatabase) Close() {
	if tdb.cleanup != nil {
		tdb.cleanup()
	}
}

// TestRepositories provides repository implementations for testing
type TestRepositories struct {
	DB         *TestDatabase
	StorageSvc storage.Repository
	PubspecSvc pubspec.Repository
	cleanup    func()
}

// SetupTestRepositories creates a complete test environment with all repositories
func SetupTestRepositories(t *testing.T) *TestRepositories {
	t.Helper()

	// Setup database
	db := SetupTestDatabase(t)

	// Create temporary directory for storage
	tmpDir := t.TempDir()
	storageRepo := storage.NewLocalRepository(tmpDir)
	pubspecRepo := pubspec.NewParserRepository()

	return &TestRepositories{
		DB:         db,
		StorageSvc: storageRepo,
		PubspecSvc: pubspecRepo,
		cleanup: func() {
			db.Close()
		},
	}
}

// Close closes all test repositories
func (tr *TestRepositories) Close() {
	if tr.cleanup != nil {
		tr.cleanup()
	}
}

// CreateTestPackage creates a test package in the database
func (tdb *TestDatabase) CreateTestPackage(ctx context.Context, name string, private bool) (*domain.Package, error) {
	return tdb.Repo.CreatePackage(ctx, name, private)
}

// CreateTestPackageWithMetadata creates a test package with full metadata
func (tdb *TestDatabase) CreateTestPackageWithMetadata(ctx context.Context, req CreatePackageRequest) (*domain.Package, error) {
	pkg, err := tdb.Repo.CreatePackage(ctx, req.Name, req.Private)
	if err != nil {
		return nil, err
	}

	// Update with metadata if provided
	if req.Description != nil || req.Homepage != nil || req.Repository != nil ||
		req.Documentation != nil {

		// Create a version of the sqlite repository that supports metadata updates
		sqliteRepo := tdb.Repo.(*sqlitePackageRepository)

		updateParams := sqlite.UpdatePackageMetadataParams{
			ID: int64(pkg.ID),
		}

		if req.Description != nil {
			updateParams.Description = sql.NullString{String: *req.Description, Valid: true}
		}
		if req.Homepage != nil {
			updateParams.Homepage = sql.NullString{String: *req.Homepage, Valid: true}
		}
		if req.Repository != nil {
			updateParams.Repository = sql.NullString{String: *req.Repository, Valid: true}
		}
		if req.Documentation != nil {
			updateParams.Documentation = sql.NullString{String: *req.Documentation, Valid: true}
		}

		err = sqliteRepo.queries.UpdatePackageMetadata(ctx, updateParams)
		if err != nil {
			return nil, err
		}

		// Refetch the updated package
		return tdb.Repo.GetPackage(ctx, req.Name)
	}

	return pkg, nil
}

// CreatePackageRequest contains parameters for creating a test package
type CreatePackageRequest struct {
	Name          string
	Private       bool
	Description   *string
	Homepage      *string
	Repository    *string
	Documentation *string
}

// CreateTestPackageVersion creates a test package version
func (tdb *TestDatabase) CreateTestPackageVersion(ctx context.Context, packageID int32, req CreateVersionRequest) (*domain.PackageVersion, error) {
	version := &domain.PackageVersion{
		PackageID:     packageID,
		Version:       req.Version,
		Description:   req.Description,
		PubspecYaml:   req.PubspecYaml,
		Readme:        req.Readme,
		Changelog:     req.Changelog,
		ArchivePath:   req.ArchivePath,
		ArchiveSha256: req.ArchiveSha256,
		Uploader:      req.Uploader,
	}

	return tdb.Repo.CreateVersion(ctx, version)
}

// CreateVersionRequest contains parameters for creating a test package version
type CreateVersionRequest struct {
	Version       string
	Description   *string
	PubspecYaml   string
	Readme        *string
	Changelog     *string
	ArchivePath   string
	ArchiveSha256 *string
	Uploader      *string
}

// CreateTestArchive creates a test archive file and returns the path
func (tr *TestRepositories) CreateTestArchive(t *testing.T, name, version string, content []byte) string {
	t.Helper()

	archivePath, err := tr.StorageSvc.Store(name, version, content)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}

	return archivePath
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// Must panics if err is not nil, useful for test setup
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

// CreateTestTarGzArchive creates a .tar.gz archive with the specified files
func CreateTestTarGzArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	for filename, content := range files {
		// Add file to tar
		header := &tar.Header{
			Name: filename,
			Mode: 0644,
			Size: int64(len(content)),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("Failed to write tar header: %v", err)
		}

		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write tar content: %v", err)
		}
	}

	// Close writers
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("Failed to close tar writer: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %v", err)
	}

	return buf.Bytes()
}
