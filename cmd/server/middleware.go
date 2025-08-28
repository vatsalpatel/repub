package main

import (
	"context"
	"log/slog"
	"net/http"
	"repub/internal/auth"
	"repub/internal/service"
)


// RequireAuthMiddleware creates middleware that requires authentication
// writeRequired: if true, requires write tokens; if false, accepts read or write tokens
func RequireAuthMiddleware(authSvc service.AuthService, writeRequired bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			
			var err error
			if writeRequired {
				err = authSvc.AuthenticateWriteRequest(r.Context(), authHeader)
			} else {
				err = authSvc.AuthenticateReadRequest(r.Context(), authHeader)
			}
			
			if err != nil {
				authType := "read"
				if writeRequired {
					authType = "write"
				}
				slog.Debug("Authentication failed", "type", authType, "error", err, "path", r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add authentication status to context
			ctx := auth.SetAuthenticated(r.Context(), true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth wraps a handler to require read authentication (for compatibility)
func RequireAuth(authSvc service.AuthService, handler http.HandlerFunc) http.HandlerFunc {
	middleware := RequireAuthMiddleware(authSvc, false) // false = read access sufficient
	return middleware(handler).ServeHTTP
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	return auth.IsAuthenticated(ctx)
}

// OptionalAuth middleware allows both authenticated and unauthenticated requests
// Accepts both read and write tokens
func OptionalAuth(authSvc service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			
			if authHeader != "" {
				err := authSvc.AuthenticateReadRequest(r.Context(), authHeader)
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