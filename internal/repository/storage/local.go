package storage

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type localRepository struct {
	fs       FileSystem
	basePath string
}

type osFileSystem struct{}

func (osfs *osFileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (osfs *osFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (osfs *osFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (osfs *osFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (osfs *osFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func NewLocalRepository(basePath string) Repository {
	return NewLocalRepositoryWithFS(&osFileSystem{}, basePath)
}

func NewLocalRepositoryWithFS(filesystem FileSystem, basePath string) Repository {
	return &localRepository{
		fs:       filesystem,
		basePath: basePath,
	}
}

func (r *localRepository) Store(packageName, version string, data []byte) (string, error) {
	dir := filepath.Join(r.basePath, packageName, version)
	if err := r.fs.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	filename := fmt.Sprintf("%s-%s.tar.gz", packageName, version)
	path := filepath.Join(dir, filename)

	if err := r.fs.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return path, nil
}

func (r *localRepository) Get(path string) ([]byte, error) {
	file, err := r.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't fail the operation - we could log this with slog
			// For now, we silently ignore the error to not disrupt the read operation
			_ = closeErr
		}
	}()
	return io.ReadAll(file)
}

func (r *localRepository) GetReader(path string) (io.ReadCloser, error) {
	return r.fs.Open(path)
}

func (r *localRepository) Exists(path string) bool {
	_, err := r.fs.Stat(path)
	return err == nil
}

func (r *localRepository) Delete(path string) error {
	return r.fs.Remove(path)
}
