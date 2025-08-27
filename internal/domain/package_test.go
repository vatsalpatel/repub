package domain

import (
	"testing"
	"time"
)

func TestPackage(t *testing.T) {
	now := time.Now()
	pkg := Package{
		ID:        1,
		Name:      "testpkg",
		Private:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	if pkg.ID != 1 {
		t.Errorf("Expected ID 1, got %d", pkg.ID)
	}
	
	if pkg.Name != "testpkg" {
		t.Errorf("Expected name 'testpkg', got %s", pkg.Name)
	}
	
	if pkg.Private != false {
		t.Errorf("Expected private false, got %t", pkg.Private)
	}
}

func TestPackageVersion(t *testing.T) {
	now := time.Now()
	desc := "Test package"
	sha := "abc123"
	uploader := "test@example.com"
	
	version := PackageVersion{
		ID:            1,
		PackageID:     1,
		Version:       "1.0.0",
		Description:   &desc,
		PubspecYaml:   "name: testpkg\nversion: 1.0.0",
		ArchivePath:   "/storage/testpkg/1.0.0/archive.tar.gz",
		ArchiveSha256: &sha,
		Uploader:      &uploader,
		Retracted:     false,
		CreatedAt:     now,
	}
	
	if version.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", version.Version)
	}
	
	if version.Description == nil || *version.Description != desc {
		t.Errorf("Expected description '%s', got %v", desc, version.Description)
	}
	
	if version.ArchiveSha256 == nil || *version.ArchiveSha256 != sha {
		t.Errorf("Expected sha '%s', got %v", sha, version.ArchiveSha256)
	}
	
	if version.Uploader == nil || *version.Uploader != uploader {
		t.Errorf("Expected uploader '%s', got %v", uploader, version.Uploader)
	}
}

func TestPackageResponse(t *testing.T) {
	response := PackageResponse{
		Name:           "testpkg",
		IsDiscontinued: false,
		Latest: VersionResponse{
			Version:    "1.0.0",
			ArchiveURL: "http://example.com/archive.tar.gz",
			Pubspec: map[string]any{
				"name":    "testpkg",
				"version": "1.0.0",
			},
		},
		Versions: []VersionResponse{
			{
				Version:    "1.0.0",
				ArchiveURL: "http://example.com/archive.tar.gz",
				Pubspec: map[string]any{
					"name":    "testpkg",
					"version": "1.0.0",
				},
			},
		},
	}
	
	if response.Name != "testpkg" {
		t.Errorf("Expected name 'testpkg', got %s", response.Name)
	}
	
	if response.Latest.Version != "1.0.0" {
		t.Errorf("Expected latest version '1.0.0', got %s", response.Latest.Version)
	}
	
	if len(response.Versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(response.Versions))
	}
}

func TestPublishRequest(t *testing.T) {
	req := PublishRequest{
		Archive:  []byte("test archive data"),
		Uploader: "test@example.com",
	}
	
	if string(req.Archive) != "test archive data" {
		t.Errorf("Expected archive data, got %s", string(req.Archive))
	}
	
	if req.Uploader != "test@example.com" {
		t.Errorf("Expected uploader 'test@example.com', got %s", req.Uploader)
	}
}

func TestPublishResponse(t *testing.T) {
	response := PublishResponse{
		URL: "http://example.com/upload",
		Fields: map[string]string{
			"key": "value",
		},
	}
	
	if response.URL != "http://example.com/upload" {
		t.Errorf("Expected URL 'http://example.com/upload', got %s", response.URL)
	}
	
	if response.Fields["key"] != "value" {
		t.Errorf("Expected field value 'value', got %s", response.Fields["key"])
	}
}