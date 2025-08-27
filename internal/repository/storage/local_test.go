package storage

import (
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"
)

// testFS wraps fstest.MapFS to implement our FileSystem interface
type testFS struct {
	fstest.MapFS
}

// normalize path to work with fstest.MapFS (remove leading slash)
func (t *testFS) normalizePath(name string) string {
	if len(name) > 0 && name[0] == '/' {
		return name[1:]
	}
	return name
}

func (t *testFS) Open(name string) (fs.File, error) {
	return t.MapFS.Open(t.normalizePath(name))
}

func (t *testFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	t.MapFS[t.normalizePath(name)] = &fstest.MapFile{Data: data, Mode: perm}
	return nil
}

func (t *testFS) MkdirAll(path string, perm fs.FileMode) error {
	return nil // no-op for in-memory FS
}

func (t *testFS) Remove(name string) error {
	normalizedName := t.normalizePath(name)
	if _, exists := t.MapFS[normalizedName]; !exists {
		return fs.ErrNotExist
	}
	delete(t.MapFS, normalizedName)
	return nil
}

func (t *testFS) Stat(name string) (fs.FileInfo, error) {
	normalizedName := t.normalizePath(name)
	if file, exists := t.MapFS[normalizedName]; exists {
		return &testFileInfo{
			name: filepath.Base(name),
			size: int64(len(file.Data)),
			mode: file.Mode,
		}, nil
	}
	return nil, fs.ErrNotExist
}

type testFileInfo struct {
	name string
	size int64
	mode fs.FileMode
}

func (fi *testFileInfo) Name() string       { return fi.name }
func (fi *testFileInfo) Size() int64        { return fi.size }
func (fi *testFileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *testFileInfo) ModTime() time.Time { return time.Now() }
func (fi *testFileInfo) IsDir() bool        { return false }
func (fi *testFileInfo) Sys() interface{}   { return nil }

func TestLocalRepository_Store(t *testing.T) {
	fs := &testFS{fstest.MapFS{}}
	repo := NewLocalRepositoryWithFS(fs, "/storage")
	
	data := []byte("test package data")
	path, err := repo.Store("testpkg", "1.0.0", data)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	
	expected := "/storage/testpkg/1.0.0/testpkg-1.0.0.tar.gz"
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
	
	// Verify file was stored
	if !repo.Exists(path) {
		t.Error("File should exist after storing")
	}
}

func TestLocalRepository_Get(t *testing.T) {
	fs := &testFS{fstest.MapFS{}}
	repo := NewLocalRepositoryWithFS(fs, "/storage")
	
	data := []byte("test package data")
	path, err := repo.Store("testpkg", "1.0.0", data)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	
	retrieved, err := repo.Get(path)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	
	if string(retrieved) != string(data) {
		t.Errorf("Expected %s, got %s", string(data), string(retrieved))
	}
}

func TestLocalRepository_GetReader(t *testing.T) {
	fs := &testFS{fstest.MapFS{}}
	repo := NewLocalRepositoryWithFS(fs, "/storage")
	
	data := []byte("test package data")
	path, err := repo.Store("testpkg", "1.0.0", data)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	
	reader, err := repo.GetReader(path)
	if err != nil {
		t.Fatalf("GetReader failed: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("Failed to close reader: %v", err)
		}
	}()
	
	buf := make([]byte, len(data))
	n, err := reader.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	
	if n != len(data) || string(buf) != string(data) {
		t.Errorf("Expected %s, got %s", string(data), string(buf))
	}
}

func TestLocalRepository_Delete(t *testing.T) {
	fs := &testFS{fstest.MapFS{}}
	repo := NewLocalRepositoryWithFS(fs, "/storage")
	
	data := []byte("test package data")
	path, err := repo.Store("testpkg", "1.0.0", data)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	
	if !repo.Exists(path) {
		t.Error("File should exist before deletion")
	}
	
	err = repo.Delete(path)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	
	if repo.Exists(path) {
		t.Error("File should not exist after deletion")
	}
}

func TestLocalRepository_Exists(t *testing.T) {
	fs := &testFS{fstest.MapFS{}}
	repo := NewLocalRepositoryWithFS(fs, "/storage")
	
	// Test non-existent file
	if repo.Exists("/nonexistent") {
		t.Error("Non-existent file should not exist")
	}
	
	// Test existing file
	data := []byte("test")
	path, err := repo.Store("testpkg", "1.0.0", data)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	
	if !repo.Exists(path) {
		t.Error("Stored file should exist")
	}
}

func TestLocalRepository_ErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() Repository
		testFunc func(repo Repository) error
	}{
		{
			name: "Get non-existent file",
			setup: func() Repository {
				return NewLocalRepositoryWithFS(&testFS{fstest.MapFS{}}, "/storage")
			},
			testFunc: func(repo Repository) error {
				_, err := repo.Get("/nonexistent")
				return err
			},
		},
		{
			name: "GetReader non-existent file",
			setup: func() Repository {
				return NewLocalRepositoryWithFS(&testFS{fstest.MapFS{}}, "/storage")
			},
			testFunc: func(repo Repository) error {
				_, err := repo.GetReader("/nonexistent")
				return err
			},
		},
		{
			name: "Delete non-existent file",
			setup: func() Repository {
				return NewLocalRepositoryWithFS(&testFS{fstest.MapFS{}}, "/storage")
			},
			testFunc: func(repo Repository) error {
				return repo.Delete("/nonexistent")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setup()
			err := tt.testFunc(repo)
			if err == nil {
				t.Error("Expected error for operation on non-existent file")
			}
		})
	}
}

func TestNewLocalRepository_Coverage(t *testing.T) {
	// Test the constructor that uses osFileSystem
	repo := NewLocalRepository("/tmp")
	if repo == nil {
		t.Error("Expected repository, got nil")
	}
}

