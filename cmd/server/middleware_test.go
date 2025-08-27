package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"repub/internal/auth"
	"repub/internal/service"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	authSvc := service.NewAuthService("test-token")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsAuthenticated(r.Context()) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("not authenticated"))
		}
	})

	middleware := AuthMiddleware(authSvc)
	handler := middleware(testHandler)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer test-token",
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

func TestRequireAuth(t *testing.T) {
	authSvc := service.NewAuthService("test-token")

	handler := RequireAuth(authSvc, func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r.Context()) {
			t.Error("Expected authentication in context")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Test with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
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
	authSvc := service.NewAuthService("test-token")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsAuthenticated(r.Context()) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("anonymous"))
		}
	})

	middleware := OptionalAuth(authSvc)
	handler := middleware(testHandler)

	tests := []struct {
		name         string
		authHeader   string
		expectedBody string
	}{
		{
			name:         "valid token",
			authHeader:   "Bearer test-token",
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

	authenticated := IsAuthenticated(ctx)
	if !authenticated {
		t.Error("Expected authenticated context to return true")
	}

	// Test with unauthenticated context
	unauthCtx := auth.SetAuthenticated(context.Background(), false)
	authenticated = IsAuthenticated(unauthCtx)
	if authenticated {
		t.Error("Expected unauthenticated context to return false")
	}

	// Test with no authentication in context
	emptyCtx := context.Background()
	authenticated = IsAuthenticated(emptyCtx)
	if authenticated {
		t.Error("Expected empty context to return false")
	}
}

