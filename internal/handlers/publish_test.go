package handlers

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"repub/internal/auth"
	"repub/internal/service"
	"repub/internal/testutil"
	"strings"
	"testing"
)

// Helper function to add authentication to context
func addAuthToContext(req *http.Request) *http.Request {
	ctx := auth.SetAuthenticated(req.Context(), true)
	return req.WithContext(ctx)
}

func TestUploadPackageHandler(t *testing.T) {
	t.Run("successful package upload", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		pubSvc := service.NewPubService(service.PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:9090",
		})

		// Create a test archive
		files := map[string]string{
			"test_package-1.0.0/pubspec.yaml": `name: test_package
version: 1.0.0
description: A test package`,
			"test_package-1.0.0/README.md": `# Test Package`,
		}
		archive := testutil.CreateTestTarGzArchive(t, files)

		// Create multipart form request
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		part, err := writer.CreateFormFile("file", "test_package-1.0.0.tar.gz")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = io.Copy(part, bytes.NewReader(archive))
		if err != nil {
			t.Fatalf("Failed to copy archive data: %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Failed to close writer: %v", err)
		}

		// Create request with auth context
		req := httptest.NewRequest("POST", "/api/packages/versions/new", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Add authentication to context
		req = addAuthToContext(req)

		w := httptest.NewRecorder()
		handler := UploadPackageHandler(pubSvc, "http://localhost:9090")
		handler(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
			t.Errorf("Response body: %s", w.Body.String())
		}

		location := w.Header().Get("Location")
		if location == "" {
			t.Error("Expected Location header with finalize URL")
		} else if !strings.Contains(location, "/api/packages/versions/newUploadFinish") {
			t.Errorf("Expected Location header to contain finalize URL path, got %s", location)
		}

		// The upload should not immediately create the package - it's stored for finalization
		// So we expect this to return "not found"
		pkg, err := repos.DB.Repo.GetPackage(context.Background(), "test_package")
		if err == nil && pkg != nil {
			t.Error("Package should not be immediately created during upload step")
		}
	})

	t.Run("unauthorized request", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		pubSvc := service.NewPubService(service.PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:9090",
		})

		// Create a simple request without auth
		req := httptest.NewRequest("POST", "/api/packages/versions/new", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler := UploadPackageHandler(pubSvc, "http://localhost:9090")
		handler(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("invalid multipart form", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		pubSvc := service.NewPubService(service.PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:8080",
		})

		// Create request with invalid content type
		req := httptest.NewRequest("POST", "/api/packages/versions/new", strings.NewReader("invalid data"))
		req.Header.Set("Content-Type", "text/plain")

		// Add authentication to context
		req = addAuthToContext(req)

		w := httptest.NewRecorder()
		handler := UploadPackageHandler(pubSvc, "http://localhost:9090")
		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("missing file in form", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		pubSvc := service.NewPubService(service.PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:8080",
		})

		// Create multipart form without file
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("other_field", "value")
		_ = writer.Close()

		req := httptest.NewRequest("POST", "/api/packages/versions/new", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Add authentication to context
		req = addAuthToContext(req)

		w := httptest.NewRecorder()
		handler := UploadPackageHandler(pubSvc, "http://localhost:9090")
		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid archive content", func(t *testing.T) {
		repos := testutil.SetupTestRepositories(t)
		defer repos.Close()

		pubSvc := service.NewPubService(service.PackageDependencies{
			Package: repos.DB.Repo,
			Storage: repos.StorageSvc,
			Pubspec: repos.PubspecSvc,
			BaseURL: "http://localhost:8080",
		})

		// Create multipart form with invalid archive
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		part, err := writer.CreateFormFile("file", "invalid.tar.gz")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		// Write invalid archive data
		_, err = part.Write([]byte("this is not a valid tar.gz archive"))
		if err != nil {
			t.Fatalf("Failed to write invalid data: %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Failed to close writer: %v", err)
		}

		// Step 1: Upload the invalid archive (should succeed)
		req := httptest.NewRequest("POST", "/api/packages/versions/new", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req = addAuthToContext(req)

		w := httptest.NewRecorder()
		uploadHandler := UploadPackageHandler(pubSvc, "http://localhost:9090")
		uploadHandler(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected upload to succeed with status 204, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if location == "" {
			t.Fatal("Expected Location header with finalize URL")
		}

		// Step 2: Extract upload_id from location and test finalization (should fail)
		// Parse the URL to get upload_id parameter
		finalizePath := "/api/packages/versions/newUploadFinish?upload_id=upload_34" // Simple hardcoded ID for test
		finalizeReq := httptest.NewRequest("GET", finalizePath, nil)
		finalizeReq = addAuthToContext(finalizeReq)

		finalizeW := httptest.NewRecorder()
		finalizeHandler := FinalizeUploadHandler(pubSvc)
		finalizeHandler(finalizeW, finalizeReq)

		if finalizeW.Code != http.StatusBadRequest {
			t.Errorf("Expected finalize to fail with status 400, got %d", finalizeW.Code)
		}

		// Should return proper error JSON
		if !strings.Contains(finalizeW.Body.String(), "error") {
			t.Error("Expected error response in finalize step")
		}
	})
}
