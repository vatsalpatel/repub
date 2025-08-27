package main

import (
	"context"
	"log/slog"
	"net/http"
	"repub/internal/auth"
	"repub/internal/service"
)


// AuthMiddleware creates middleware that validates bearer tokens
func AuthMiddleware(authSvc service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			
			err := authSvc.AuthenticateRequest(r.Context(), authHeader)
			if err != nil {
				slog.Debug("Authentication failed", "error", err, "path", r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add authentication status to context
			ctx := auth.SetAuthenticated(r.Context(), true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuthMiddleware creates middleware that requires authentication
func RequireAuthMiddleware(authSvc service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			
			err := authSvc.AuthenticateRequest(r.Context(), authHeader)
			if err != nil {
				slog.Debug("Authentication required but failed", "error", err, "path", r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add authentication status to context
			ctx := auth.SetAuthenticated(r.Context(), true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth wraps a handler to require authentication (for compatibility)
func RequireAuth(authSvc service.AuthService, handler http.HandlerFunc) http.HandlerFunc {
	middleware := RequireAuthMiddleware(authSvc)
	return middleware(handler).ServeHTTP
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	return auth.IsAuthenticated(ctx)
}

// OptionalAuth middleware allows both authenticated and unauthenticated requests
func OptionalAuth(authSvc service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			
			if authHeader != "" {
				err := authSvc.AuthenticateRequest(r.Context(), authHeader)
				if err == nil {
					// Add authentication status to context if authentication succeeds
					ctx := auth.SetAuthenticated(r.Context(), true)
					r = r.WithContext(ctx)
				}
				// If authentication fails, continue without authentication (don't error)
			}

			next.ServeHTTP(w, r)
		})
	}
}