package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	gcs "cloud.google.com/go/storage"
)

const (
	gcsTestBucket    = "test-bucket"
	gcsTestContainer = "fake-gcs-test"
	gcsTestPort      = "4443"
)

var gcsTestClient *gcs.Client

func TestMain(m *testing.M) {
	if err := startGCSEmulator(); err != nil {
		fmt.Println("GCS emulator not available:", err)
	}

	code := m.Run()

	stopGCSEmulator()
	os.Exit(code)
}

func startGCSEmulator() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found: %w", err)
	}

	_ = exec.Command("docker", "rm", "-f", gcsTestContainer).Run()

	emulatorURL := fmt.Sprintf("http://localhost:%s", gcsTestPort)
	cmd := exec.Command("docker", "run", "-d",
		"--name", gcsTestContainer,
		"-p", gcsTestPort+":4443",
		"fsouza/fake-gcs-server",
		"-scheme", "http",
		"-port", "4443",
		"-external-url", emulatorURL,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start container: %v\n%s", err, out)
	}

	if err := waitForEmulator(); err != nil {
		stopGCSEmulator()
		return err
	}

	os.Setenv("STORAGE_EMULATOR_HOST", "localhost:"+gcsTestPort)

	client, err := gcs.NewClient(context.Background(), gcs.WithJSONReads())
	if err != nil {
		stopGCSEmulator()
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Bucket(gcsTestBucket).Create(context.Background(), "test-project", nil); err != nil {
		stopGCSEmulator()
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	gcsTestClient = client
	return nil
}

func waitForEmulator() error {
	baseURL := fmt.Sprintf("http://localhost:%s", gcsTestPort)
	for range 30 {
		resp, err := http.Get(baseURL + "/storage/v1/b")
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("emulator failed to become ready")
}

func stopGCSEmulator() {
	_ = exec.Command("docker", "rm", "-f", gcsTestContainer).Run()
	os.Unsetenv("STORAGE_EMULATOR_HOST")
}

func skipIfNoEmulator(t *testing.T) {
	t.Helper()
	if gcsTestClient == nil {
		t.Skip("GCS emulator not available")
	}
}

func newTestGCSRepo(t *testing.T) Repository {
	t.Helper()
	skipIfNoEmulator(t)
	return newGCSRepositoryWithClient(gcsTestClient, gcsTestBucket)
}

func TestGCSRepository_Store(t *testing.T) {
	repo := newTestGCSRepo(t)

	data := []byte("test package data")
	path, err := repo.Store("testpkg", "1.0.0", data)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	expected := "testpkg/1.0.0/testpkg-1.0.0.tar.gz"
	if path != expected {
		t.Errorf("expected path %s, got %s", expected, path)
	}

	if !repo.Exists(path) {
		t.Error("file should exist after storing")
	}
}

func TestGCSRepository_Get(t *testing.T) {
	repo := newTestGCSRepo(t)

	data := []byte("test get data")
	path, err := repo.Store("getpkg", "1.0.0", data)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	retrieved, err := repo.Get(path)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !bytes.Equal(retrieved, data) {
		t.Errorf("expected %s, got %s", data, retrieved)
	}
}

func TestGCSRepository_GetReader(t *testing.T) {
	repo := newTestGCSRepo(t)

	data := []byte("test reader data")
	path, err := repo.Store("readerpkg", "1.0.0", data)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	reader, err := repo.GetReader(path)
	if err != nil {
		t.Fatalf("GetReader failed: %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if !bytes.Equal(got, data) {
		t.Errorf("expected %s, got %s", data, got)
	}
}

func TestGCSRepository_Delete(t *testing.T) {
	repo := newTestGCSRepo(t)

	data := []byte("delete me")
	path, err := repo.Store("delpkg", "1.0.0", data)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if !repo.Exists(path) {
		t.Error("file should exist before deletion")
	}

	if err := repo.Delete(path); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if repo.Exists(path) {
		t.Error("file should not exist after deletion")
	}
}

func TestGCSRepository_Exists_NonExistent(t *testing.T) {
	repo := newTestGCSRepo(t)

	if repo.Exists("nonexistent/path") {
		t.Error("non-existent file should not exist")
	}
}

func TestGCSRepository_LegacyPathStripping(t *testing.T) {
	repo := newTestGCSRepo(t)

	data := []byte("legacy data")
	path, err := repo.Store("legacypkg", "2.0.0", data)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	legacyPath := "/app/storage/" + path

	retrieved, err := repo.Get(legacyPath)
	if err != nil {
		t.Fatalf("Get with legacy path failed: %v", err)
	}
	if !bytes.Equal(retrieved, data) {
		t.Errorf("expected %s, got %s", data, retrieved)
	}

	if !repo.Exists(legacyPath) {
		t.Error("Exists should work with legacy path")
	}

	reader, err := repo.GetReader(legacyPath)
	if err != nil {
		t.Fatalf("GetReader with legacy path failed: %v", err)
	}
	reader.Close()

	if err := repo.Delete(legacyPath); err != nil {
		t.Fatalf("Delete with legacy path failed: %v", err)
	}

	if repo.Exists(path) {
		t.Error("file should not exist after deletion via legacy path")
	}
}

func TestGCSRepository_ErrorCases(t *testing.T) {
	repo := newTestGCSRepo(t)

	_, err := repo.Get("nonexistent/object")
	if err == nil {
		t.Error("Get non-existent should return error")
	}

	_, err = repo.GetReader("nonexistent/object")
	if err == nil {
		t.Error("GetReader non-existent should return error")
	}
}
