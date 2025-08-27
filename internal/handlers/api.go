package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"repub/internal/service"

	"github.com/go-chi/chi/v5"
)

// API handlers for Dart pub protocol endpoints

func GetPackageHandler(pubSvc service.PubService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		packageName := chi.URLParam(r, "package")

		pkg, err := pubSvc.GetPackage(r.Context(), packageName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if pkg == nil {
			http.Error(w, "Package not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.pub.v2+json")
		if err := json.NewEncoder(w).Encode(pkg); err != nil {
			slog.Error("Failed to encode package response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func GetPackageVersionHandler(pubSvc service.PubService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		packageName := chi.URLParam(r, "package")
		version := chi.URLParam(r, "version")

		versionResp, err := pubSvc.GetPackageVersion(r.Context(), packageName, version)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if versionResp == nil {
			http.Error(w, "Version not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.pub.v2+json")
		if err := json.NewEncoder(w).Encode(versionResp); err != nil {
			slog.Error("Failed to encode version response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func GetAdvisoriesHandler(pubSvc service.PubService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		packageName := chi.URLParam(r, "package")

		advisories, err := pubSvc.GetAdvisories(r.Context(), packageName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.pub.v2+json")
		if err := json.NewEncoder(w).Encode(advisories); err != nil {
			slog.Error("Failed to encode advisories response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func DownloadPackageHandler(pubSvc service.PubService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		packageName := chi.URLParam(r, "package")
		version := chi.URLParam(r, "version")

		data, err := pubSvc.DownloadPackage(r.Context(), packageName, version)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+packageName+"-"+version+".tar.gz\"")
		
		if _, err := w.Write(data); err != nil {
			slog.Error("Failed to write download response", "error", err)
		}
	}
}

// NewPackageVersionHandler returns the initial upload form for pub protocol
func NewPackageVersionHandler(pubSvc service.PubService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// According to pub protocol, this endpoint should return upload URL and fields
		// Build the absolute URL from the request
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)
		
		response := map[string]interface{}{
			"url": baseURL + "/api/packages/versions/new",
			"fields": map[string]string{},
		}

		w.Header().Set("Content-Type", "application/vnd.pub.v2+json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("Failed to encode new version response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}