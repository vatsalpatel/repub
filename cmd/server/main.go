package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	authmiddleware "repub/internal/auth/middleware"
	"repub/internal/config"
	"repub/internal/handlers"
	"repub/internal/repository/pkg"
	"repub/internal/repository/pkg/postgres"
	"repub/internal/repository/pubspec"
	"repub/internal/repository/storage"
	"repub/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Load()

	// Connect to database
	dbConn, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()

	if err := dbConn.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Initialize layers
	queries := postgres.New(dbConn)
	storageRepo := storage.NewLocalRepository(cfg.StoragePath)
	pubspecRepo := pubspec.NewParserRepository()

	// Repository layer
	packageRepo := pkg.NewPostgresPackageRepository(queries)

	// Service layer
	pubSvc := service.NewPubService(service.PackageDependencies{
		Storage: storageRepo,
		Package: packageRepo,
		Pubspec: pubspecRepo,
		BaseURL: cfg.BaseURL,
	})
	authSvc := service.NewAuthService(cfg.ReadTokens, cfg.WriteTokens)

	// Setup router
	r := setupRouter(pubSvc, authSvc)

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func setupRouter(pubSvc service.PubService, authSvc service.AuthService) *chi.Mux {
	cfg := config.Load() // Get config for base URL
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(authmiddleware.OptionalAuth(authSvc))

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/packages", func(r chi.Router) {
			// Read-only routes (require read tokens)
			r.Group(func(r chi.Router) {
				r.Use(authmiddleware.RequireAuthMiddleware(authSvc, false)) // false = read access sufficient
				r.Get("/{package}", handlers.GetPackageHandler(pubSvc))
				r.Get("/{package}/versions/{version}", handlers.GetPackageVersionHandler(pubSvc))
				r.Get("/{package}/advisories", handlers.GetAdvisoriesHandler(pubSvc))
			})

			// Write routes (require write tokens)
			r.Group(func(r chi.Router) {
				r.Use(authmiddleware.RequireAuthMiddleware(authSvc, true)) // true = write required
				r.Get("/versions/new", handlers.NewPackageVersionHandler(pubSvc))
				r.Post("/versions/new", handlers.UploadPackageHandler(pubSvc, cfg.BaseURL))
				r.Get("/versions/newUploadFinish", handlers.FinalizeUploadHandler(pubSvc))
			})
		})
	})

	// Package download routes
	r.Get("/packages/{package}/versions/{version}/download", handlers.DownloadPackageHandler(pubSvc))

	// Web routes (SSR with templ)
	r.Get("/", handlers.IndexHandler())
	r.Get("/packages", handlers.PackagesListHandler(pubSvc))
	r.Get("/packages/{package}", handlers.PackageDetailHandler(pubSvc))
	r.Get("/packages/{package}/versions/{version}", handlers.VersionDetailHandler(pubSvc))

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	return r
}
