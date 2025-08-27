package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"repub/internal/auth"
	"repub/internal/domain"
	"repub/internal/service"
	"strings"
	"sync"
)

// In-memory storage for pending uploads (development implementation)
var (
	pendingUploads = make(map[string]*domain.PublishRequest)
	uploadMutex    = sync.RWMutex{}
)

// UploadPackageHandler handles package upload (step 2 of the workflow)
func UploadPackageHandler(pubSvc service.PubService, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAuthenticated(r.Context()) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse multipart form (dart pub client sends the archive as a file)
		err := r.ParseMultipartForm(32 << 20) // 32MB max memory
		if err != nil {
			slog.Error("Failed to parse multipart form", "error", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Get the uploaded file
		file, _, err := r.FormFile("file")
		if err != nil {
			slog.Error("Failed to get uploaded file", "error", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		defer func() { _ = file.Close() }()

		// Read the archive data
		archiveData, err := io.ReadAll(file)
		if err != nil {
			slog.Error("Failed to read archive data", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Create publish request and store it temporarily
		publishReq := &domain.PublishRequest{
			Archive:  archiveData,
			Uploader: "authenticated-user",
		}

		// Generate a unique finalize token
		finalizeToken := fmt.Sprintf("upload_%d", len(archiveData)) // Simple token generation
		
		// Store the upload for finalization
		uploadMutex.Lock()
		pendingUploads[finalizeToken] = publishReq
		uploadMutex.Unlock()

		// Return 204 with finalize URL as per pub spec
		finalizeURL := fmt.Sprintf("%s/api/packages/versions/newUploadFinish?upload_id=%s", 
			strings.TrimSuffix(baseURL, "/"), finalizeToken)
		
		w.Header().Set("Location", finalizeURL)
		w.WriteHeader(http.StatusNoContent)
		slog.Info("Package upload received, awaiting finalization", "finalize_url", finalizeURL)
	}
}

// FinalizeUploadHandler handles the finalization of package upload (step 3 of the workflow)
func FinalizeUploadHandler(pubSvc service.PubService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAuthenticated(r.Context()) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get upload ID from query parameters
		uploadID := r.URL.Query().Get("upload_id")
		if uploadID == "" {
			http.Error(w, "Missing upload_id parameter", http.StatusBadRequest)
			return
		}

		// Retrieve the pending upload
		uploadMutex.Lock()
		publishReq, exists := pendingUploads[uploadID]
		if exists {
			delete(pendingUploads, uploadID) // Remove from pending
		}
		uploadMutex.Unlock()

		if !exists {
			response := map[string]interface{}{
				"error": map[string]string{
					"code":    "UPLOAD_NOT_FOUND",
					"message": "Upload not found or already processed",
				},
			}
			w.Header().Set("Content-Type", "application/vnd.pub.v2+json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(response)
			return
		}

		// Now actually publish the package
		_, err := pubSvc.PublishPackage(r.Context(), publishReq)
		if err != nil {
			slog.Error("Failed to publish package", "error", err)
			response := map[string]interface{}{
				"error": map[string]string{
					"code":    "PUBLISH_FAILED",
					"message": err.Error(),
				},
			}
			w.Header().Set("Content-Type", "application/vnd.pub.v2+json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(response)
			return
		}

		// Return success response as per pub spec
		response := map[string]interface{}{
			"success": map[string]string{
				"message": "Package published successfully",
			},
		}

		w.Header().Set("Content-Type", "application/vnd.pub.v2+json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("Failed to encode success response", "error", err)
		}
		slog.Info("Package published successfully")
	}
}