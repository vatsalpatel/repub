package service

import (
	"context"
	"repub/internal/domain"
	"repub/internal/testutil"
	"strings"
	"testing"
)

func TestPubService_GetPackage(t *testing.T) {
	repos := testutil.SetupTestRepositories(t)
	defer repos.Close()

	svc := NewPubService(PackageDependencies{
		Package: repos.DB.Repo,
		Storage: repos.StorageSvc,
		Pubspec: repos.PubspecSvc,
		BaseURL: "http://localhost:8080",
	})

	ctx := context.Background()

	// Create test package
	pkg, err := repos.DB.CreateTestPackage(ctx, "testpkg", false)
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	// Create test version
	desc := "Test package"
	_, err = repos.DB.CreateTestPackageVersion(ctx, pkg.ID, testutil.CreateVersionRequest{
		Version:     "1.0.0",
		Description: &desc,
		PubspecYaml: "name: testpkg\nversion: 1.0.0",
		ArchivePath: "/storage/testpkg/1.0.0/testpkg-1.0.0.tar.gz",
	})
	if err != nil {
		t.Fatalf("Failed to create version: %v", err)
	}

	// Test GetPackage
	result, err := svc.GetPackage(ctx, "testpkg")
	if err != nil {
		t.Fatalf("GetPackage failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected package, got nil")
	}

	if result.Name != "testpkg" {
		t.Errorf("Expected name 'testpkg', got %s", result.Name)
	}

	if result.Latest.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", result.Latest.Version)
	}

	if len(result.Versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(result.Versions))
	}
}

func TestPubService_GetPackage_NotFound(t *testing.T) {
	repos := testutil.SetupTestRepositories(t)
	defer repos.Close()

	svc := NewPubService(PackageDependencies{
		Package: repos.DB.Repo,
		Storage: repos.StorageSvc,
		Pubspec: repos.PubspecSvc,
		BaseURL: "http://localhost:8080",
	})

	result, err := svc.GetPackage(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("GetPackage failed: %v", err)
	}

	if result != nil {
		t.Error("Expected nil for non-existent package")
	}
}

func TestPubService_ListPackages(t *testing.T) {
	repos := testutil.SetupTestRepositories(t)
	defer repos.Close()

	svc := NewPubService(PackageDependencies{
		Package: repos.DB.Repo,
		Storage: repos.StorageSvc,
		Pubspec: repos.PubspecSvc,
		BaseURL: "http://localhost:8080",
	})

	ctx := context.Background()

	// Create test packages
	_, err := repos.DB.CreateTestPackage(ctx, "pkg1", false)
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	_, err = repos.DB.CreateTestPackage(ctx, "pkg2", true)
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	// Test ListPackages
	result, err := svc.ListPackages(ctx, 1, 10)
	if err != nil {
		t.Fatalf("ListPackages failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(result))
	}
}

func TestPubService_PublishPackage(t *testing.T) {
	t.Run("successful first package publish", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		svc := NewPubService(PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:8080",
		})

		// Create a test archive with pubspec.yaml, README.md, and CHANGELOG.md
		files := map[string]string{
			"test_package-1.0.0/pubspec.yaml": `name: test_package
version: 1.0.0
description: A test package
homepage: https://example.com
repository: https://github.com/example/test_package`,
			"test_package-1.0.0/README.md": `# Test Package

This is a test package for testing purposes.`,
			"test_package-1.0.0/CHANGELOG.md": `## 1.0.0

- Initial release`,
		}
		archive := testutil.CreateTestTarGzArchive(t, files)

		req := &domain.PublishRequest{
			Archive:  archive,
			Uploader: "test@example.com",
		}

		result, err := svc.PublishPackage(context.Background(), req)
		if err != nil {
			t.Fatalf("PublishPackage failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected publish response, got nil")
		}

		if result.URL == "" {
			t.Error("Expected non-empty URL in response")
		}

		// Verify the package was created
		ctx := context.Background()
		pkg, err := repos.DB.Repo.GetPackage(ctx, "test_package")
		if err != nil {
			t.Fatalf("Failed to get created package: %v", err)
		}
		if pkg == nil {
			t.Fatal("Package was not created")
		}

		// Verify the version was created
		versions, err := repos.DB.Repo.GetPackageVersions(ctx, pkg.ID)
		if err != nil {
			t.Fatalf("Failed to get package versions: %v", err)
		}
		if len(versions) != 1 {
			t.Errorf("Expected 1 version, got %d", len(versions))
		}
		if versions[0].Version != "1.0.0" {
			t.Errorf("Expected version 1.0.0, got %s", versions[0].Version)
		}

		// Verify README and CHANGELOG were stored
		if versions[0].Readme == nil {
			t.Error("Expected README to be stored")
		} else if !strings.Contains(*versions[0].Readme, "test package for testing") {
			t.Error("README content not correctly stored")
		}

		if versions[0].Changelog == nil {
			t.Error("Expected CHANGELOG to be stored")
		} else if !strings.Contains(*versions[0].Changelog, "Initial release") {
			t.Error("CHANGELOG content not correctly stored")
		}

		// Verify uploader was added
		uploaders, err := repos.DB.Repo.GetUploaders(ctx, pkg.ID)
		if err != nil {
			t.Fatalf("Failed to get uploaders: %v", err)
		}
		if len(uploaders) != 1 || uploaders[0] != "test@example.com" {
			t.Errorf("Expected uploader test@example.com, got %v", uploaders)
		}
	})

	t.Run("publish additional version by authorized uploader", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		svc := NewPubService(PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:8080",
		})

		ctx := context.Background()

		// First, publish version 1.0.0
		files1 := map[string]string{
			"test_package-1.0.0/pubspec.yaml": `name: test_package
version: 1.0.0
description: A test package`,
		}
		archive1 := testutil.CreateTestTarGzArchive(t, files1)

		req1 := &domain.PublishRequest{
			Archive:  archive1,
			Uploader: "test@example.com",
		}

		_, err := svc.PublishPackage(ctx, req1)
		if err != nil {
			t.Fatalf("First publish failed: %v", err)
		}

		// Then publish version 1.1.0 by the same uploader
		files2 := map[string]string{
			"test_package-1.1.0/pubspec.yaml": `name: test_package
version: 1.1.0
description: A test package with updates`,
		}
		archive2 := testutil.CreateTestTarGzArchive(t, files2)

		req2 := &domain.PublishRequest{
			Archive:  archive2,
			Uploader: "test@example.com",
		}

		result, err := svc.PublishPackage(ctx, req2)
		if err != nil {
			t.Fatalf("Second publish failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected publish response, got nil")
		}

		// Verify both versions exist
		pkg, err := repos.DB.Repo.GetPackage(ctx, "test_package")
		if err != nil {
			t.Fatalf("Failed to get package: %v", err)
		}

		versions, err := repos.DB.Repo.GetPackageVersions(ctx, pkg.ID)
		if err != nil {
			t.Fatalf("Failed to get package versions: %v", err)
		}
		if len(versions) != 2 {
			t.Errorf("Expected 2 versions, got %d", len(versions))
		}
	})

	t.Run("reject unauthorized uploader", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		svc := NewPubService(PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:8080",
		})

		ctx := context.Background()

		// First, publish version 1.0.0 by original uploader
		files1 := map[string]string{
			"test_package-1.0.0/pubspec.yaml": `name: test_package
version: 1.0.0
description: A test package`,
		}
		archive1 := testutil.CreateTestTarGzArchive(t, files1)

		req1 := &domain.PublishRequest{
			Archive:  archive1,
			Uploader: "original@example.com",
		}

		_, err := svc.PublishPackage(ctx, req1)
		if err != nil {
			t.Fatalf("First publish failed: %v", err)
		}

		// Try to publish version 1.1.0 by different uploader (should fail)
		files2 := map[string]string{
			"test_package-1.1.0/pubspec.yaml": `name: test_package
version: 1.1.0
description: Malicious update`,
		}
		archive2 := testutil.CreateTestTarGzArchive(t, files2)

		req2 := &domain.PublishRequest{
			Archive:  archive2,
			Uploader: "malicious@example.com",
		}

		_, err = svc.PublishPackage(ctx, req2)
		if err == nil {
			t.Error("Expected unauthorized uploader to be rejected")
		}
		if !strings.Contains(err.Error(), "unauthorized") {
			t.Errorf("Expected unauthorized error, got: %v", err)
		}
	})

	t.Run("reject duplicate version", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		svc := NewPubService(PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:8080",
		})

		ctx := context.Background()

		// Publish version 1.0.0
		files := map[string]string{
			"test_package-1.0.0/pubspec.yaml": `name: test_package
version: 1.0.0
description: A test package`,
		}
		archive := testutil.CreateTestTarGzArchive(t, files)

		req := &domain.PublishRequest{
			Archive:  archive,
			Uploader: "test@example.com",
		}

		_, err := svc.PublishPackage(ctx, req)
		if err != nil {
			t.Fatalf("First publish failed: %v", err)
		}

		// Try to publish the same version again
		_, err = svc.PublishPackage(ctx, req)
		if err == nil {
			t.Error("Expected duplicate version to be rejected")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected version exists error, got: %v", err)
		}
	})

	t.Run("reject invalid pubspec", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		svc := NewPubService(PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:8080",
		})

		// Create archive with invalid pubspec.yaml
		files := map[string]string{
			"test_package-1.0.0/pubspec.yaml": `invalid yaml: [
			missing closing bracket`,
		}
		archive := testutil.CreateTestTarGzArchive(t, files)

		req := &domain.PublishRequest{
			Archive:  archive,
			Uploader: "test@example.com",
		}

		_, err := svc.PublishPackage(context.Background(), req)
		if err == nil {
			t.Error("Expected invalid pubspec to be rejected")
		}
		if !strings.Contains(err.Error(), "parse pubspec.yaml") {
			t.Errorf("Expected pubspec parse error, got: %v", err)
		}
	})

	t.Run("reject archive without pubspec", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		svc := NewPubService(PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:8080",
		})

		// Create archive without pubspec.yaml
		files := map[string]string{
			"test_package-1.0.0/README.md": `# Test Package`,
		}
		archive := testutil.CreateTestTarGzArchive(t, files)

		req := &domain.PublishRequest{
			Archive:  archive,
			Uploader: "test@example.com",
		}

		_, err := svc.PublishPackage(context.Background(), req)
		if err == nil {
			t.Error("Expected archive without pubspec to be rejected")
		}
		if !strings.Contains(err.Error(), "pubspec.yaml not found") {
			t.Errorf("Expected missing pubspec error, got: %v", err)
		}
	})
}

func TestStringValue(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "nil pointer",
			input:    nil,
			expected: "",
		},
		{
			name:     "non-nil pointer",
			input:    stringPtr("test"),
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringValue(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPubService_ErrorCases(t *testing.T) {
	t.Run("GetPackage with no versions", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		svc := NewPubService(PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:8080",
		})

		ctx := context.Background()

		// Create package without versions
		_, err := repos.DB.CreateTestPackage(ctx, "testpkg", false)
		if err != nil {
			t.Fatalf("Failed to create package: %v", err)
		}

		// Should return error when no versions exist
		_, err = svc.GetPackage(ctx, "testpkg")
		if err == nil {
			t.Error("Expected error for package with no versions, got nil")
		}
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
