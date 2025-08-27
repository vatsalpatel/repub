package handlers

import (
	"log/slog"
	"net/http"
	"repub/internal/service"
	"repub/web/templates"

	"github.com/go-chi/chi/v5"
)

// Web handlers for server-side rendered pages

func IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if err := templates.Index().Render(r.Context(), w); err != nil {
			slog.Error("Failed to render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func PackagesListHandler(pubSvc service.PubService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		packages, err := pubSvc.ListPackages(r.Context(), 1, 20)
		if err != nil {
			slog.Error("Error listing packages", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := templates.PackagesList(packages).Render(r.Context(), w); err != nil {
			slog.Error("Failed to render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func PackageDetailHandler(pubSvc service.PubService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		packageName := chi.URLParam(r, "package")

		detail, err := pubSvc.GetPackageDetail(r.Context(), packageName)
		if err != nil {
			slog.Error("Error getting package detail", "error", err, "package", packageName)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if detail == nil {
			http.Error(w, "Package not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := templates.PackageDetail(detail).Render(r.Context(), w); err != nil {
			slog.Error("Failed to render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func VersionDetailHandler(pubSvc service.PubService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		packageName := chi.URLParam(r, "package")
		version := chi.URLParam(r, "version")

		versionResp, err := pubSvc.GetPackageVersion(r.Context(), packageName, version)
		if err != nil {
			slog.Error("Error getting package version", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if versionResp == nil {
			http.Error(w, "Version not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := templates.VersionDetail(packageName, versionResp.Version).Render(r.Context(), w); err != nil {
			slog.Error("Failed to render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}