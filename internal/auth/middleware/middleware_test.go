package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"repub/internal/auth"
	"repub/internal/auth/middleware"
	"repub/internal/config"
	"repub/internal/service"
	"testing"
)

func TestRequireReadAuthMiddleware(t *testing.T) {
	readTokens := []config.Token{
		{Name: "READER", Value: "read-token"},
	}
	writeTokens := []config.Token{
		{Name: "WRITER", Value: "write-token"},
	}
	authSvc := service.NewAuthService(readTokens, writeTokens)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if middleware.IsAuthenticated(r.Context()) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("not authenticated"))
		}
	})

	middleware := middleware.RequireAuthMiddleware(authSvc, false)
	handler := middleware(testHandler)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid read token",
			authHeader:     "Bearer read-token",
			expectedStatus: http.StatusOK,
			expectedBody:   "authenticated",
		},
		{
			name:           "valid write token",
			authHeader:     "Bearer write-token",
			expectedStatus: http.StatusOK,
			expectedBody:   "authenticated",
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "no auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "malformed auth header",
			authHeader:     "Basic token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := w.Body.String()
			if body != tt.expectedBody+"\n" && body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestRequireWriteAuthMiddleware(t *testing.T) {
	readTokens := []config.Token{
		{Name: "READER", Value: "read-token"},
	}
	writeTokens := []config.Token{
		{Name: "WRITER", Value: "write-token"},
	}
	authSvc := service.NewAuthService(readTokens, writeTokens)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if middleware.IsAuthenticated(r.Context()) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("not authenticated"))
		}
	})

	middleware := middleware.RequireAuthMiddleware(authSvc, true)
	handler := middleware(testHandler)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid write token",
			authHeader:     "Bearer write-token",
			expectedStatus: http.StatusOK,
			expectedBody:   "authenticated",
		},
		{
			name:           "read token cannot write",
			authHeader:     "Bearer read-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "no auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := w.Body.String()
			if body != tt.expectedBody+"\n" && body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestRequireAuth(t *testing.T) {
	readTokens := []config.Token{
		{Name: "READER", Value: "read-token"},
	}
	writeTokens := []config.Token{
		{Name: "WRITER", Value: "write-token"},
	}
	authSvc := service.NewAuthService(readTokens, writeTokens)

	handler := middleware.RequireAuth(authSvc, func(w http.ResponseWriter, r *http.Request) {
		if !middleware.IsAuthenticated(r.Context()) {
			t.Error("Expected authentication in context")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Test with valid read token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer read-token")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test with invalid token
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	w = httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestOptionalAuth(t *testing.T) {
	readTokens := []config.Token{
		{Name: "READER", Value: "read-token"},
	}
	writeTokens := []config.Token{
		{Name: "WRITER", Value: "write-token"},
	}
	authSvc := service.NewAuthService(readTokens, writeTokens)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if middleware.IsAuthenticated(r.Context()) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("anonymous"))
		}
	})

	middleware := middleware.OptionalAuth(authSvc)
	handler := middleware(testHandler)

	tests := []struct {
		name         string
		authHeader   string
		expectedBody string
	}{
		{
			name:         "valid read token",
			authHeader:   "Bearer read-token",
			expectedBody: "authenticated",
		},
		{
			name:         "valid write token",
			authHeader:   "Bearer write-token",
			expectedBody: "authenticated",
		},
		{
			name:         "invalid token",
			authHeader:   "Bearer invalid-token",
			expectedBody: "anonymous",
		},
		{
			name:         "no auth header",
			authHeader:   "",
			expectedBody: "anonymous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			body := w.Body.String()
			if body != tt.expectedBody+"\n" && body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestIsAuthenticated(t *testing.T) {
	// Test with authenticated context
	ctx := auth.SetAuthenticated(context.Background(), true)

	authenticated := middleware.IsAuthenticated(ctx)
	if !authenticated {
		t.Error("Expected authenticated context to return true")
	}

	// Test with unauthenticated context
	unauthCtx := auth.SetAuthenticated(context.Background(), false)
	authenticated = middleware.IsAuthenticated(unauthCtx)
	if authenticated {
		t.Error("Expected unauthenticated context to return false")
	}

	// Test with no authentication in context
	emptyCtx := context.Background()
	authenticated = middleware.IsAuthenticated(emptyCtx)
	if authenticated {
		t.Error("Expected empty context to return false")
	}
}
