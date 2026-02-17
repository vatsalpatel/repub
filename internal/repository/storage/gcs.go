package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	gcs "cloud.google.com/go/storage"
)

const legacyPathPrefix = "/app/storage/"

type gcsRepository struct {
	client *gcs.Client
	bucket string
}

func NewGCSRepository(bucket string) (Repository, error) {
	client, err := gcs.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	return newGCSRepositoryWithClient(client, bucket), nil
}

func newGCSRepositoryWithClient(client *gcs.Client, bucket string) Repository {
	return &gcsRepository{client: client, bucket: bucket}
}

func (r *gcsRepository) objectKey(path string) string {
	return strings.TrimPrefix(path, legacyPathPrefix)
}

func (r *gcsRepository) Store(packageName, version string, data []byte) (string, error) {
	key := fmt.Sprintf("%s/%s/%s-%s.tar.gz", packageName, version, packageName, version)
	w := r.client.Bucket(r.bucket).Object(key).NewWriter(context.Background())
	if _, err := w.Write(data); err != nil {
		return "", fmt.Errorf("failed to write to GCS: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close GCS writer: %w", err)
	}
	return key, nil
}

func (r *gcsRepository) Get(path string) ([]byte, error) {
	key := r.objectKey(path)
	rc, err := r.client.Bucket(r.bucket).Object(key).NewReader(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to read from GCS: %w", err)
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func (r *gcsRepository) GetReader(path string) (io.ReadCloser, error) {
	key := r.objectKey(path)
	rc, err := r.client.Bucket(r.bucket).Object(key).NewReader(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get reader from GCS: %w", err)
	}
	return rc, nil
}

func (r *gcsRepository) Exists(path string) bool {
	key := r.objectKey(path)
	_, err := r.client.Bucket(r.bucket).Object(key).Attrs(context.Background())
	return err == nil
}

func (r *gcsRepository) Delete(path string) error {
	key := r.objectKey(path)
	return r.client.Bucket(r.bucket).Object(key).Delete(context.Background())
}

