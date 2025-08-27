package storage

import (
	"io"
	"io/fs"
)

type Repository interface {
	Store(packageName, version string, data []byte) (string, error)
	Get(path string) ([]byte, error)
	GetReader(path string) (io.ReadCloser, error)
	Exists(path string) bool
	Delete(path string) error
}

type FileSystem interface {
	fs.FS
	WriteFile(name string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	Remove(name string) error
	Stat(name string) (fs.FileInfo, error)
}
