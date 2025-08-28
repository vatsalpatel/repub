package service_test

import (
	"context"
	"repub/internal/config"
	"repub/internal/service"
	"testing"
)

func TestAuthService_ValidateReadToken(t *testing.T) {
	readTokens := []config.Token{
		{Name: "READER", Value: "read-token-123"},
	}
	writeTokens := []config.Token{
		{Name: "WRITER", Value: "write-token-456"},
	}
	authSvc := service.NewAuthService(readTokens, writeTokens)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid read token",
			token:       "read-token-123",
			expectError: false,
		},
		{
			name:        "valid write token (can also read)",
			token:       "write-token-456",
			expectError: false,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authSvc.ValidateReadToken(context.Background(), tt.token)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestAuthService_ValidateWriteToken(t *testing.T) {
	readTokens := []config.Token{
		{Name: "READER", Value: "read-token-123"},
	}
	writeTokens := []config.Token{
		{Name: "WRITER", Value: "write-token-456"},
	}
	authSvc := service.NewAuthService(readTokens, writeTokens)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid write token",
			token:       "write-token-456",
			expectError: false,
		},
		{
			name:        "read token cannot write",
			token:       "read-token-123",
			expectError: true,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authSvc.ValidateWriteToken(context.Background(), tt.token)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestAuthService_AuthenticateReadRequest(t *testing.T) {
	readTokens := []config.Token{
		{Name: "READER", Value: "read-token-123"},
	}
	writeTokens := []config.Token{
		{Name: "WRITER", Value: "write-token-456"},
	}
	authSvc := service.NewAuthService(readTokens, writeTokens)

	tests := []struct {
		name        string
		authHeader  string
		expectError bool
	}{
		{
			name:        "valid read token",
			authHeader:  "Bearer read-token-123",
			expectError: false,
		},
		{
			name:        "valid write token (can read)",
			authHeader:  "Bearer write-token-456",
			expectError: false,
		},
		{
			name:        "invalid bearer token",
			authHeader:  "Bearer invalid-token",
			expectError: true,
		},
		{
			name:        "missing bearer prefix",
			authHeader:  "read-token-123",
			expectError: true,
		},
		{
			name:        "empty auth header",
			authHeader:  "",
			expectError: true,
		},
		{
			name:        "bearer with no token",
			authHeader:  "Bearer ",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authSvc.AuthenticateReadRequest(context.Background(), tt.authHeader)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestAuthService_AuthenticateWriteRequest(t *testing.T) {
	readTokens := []config.Token{
		{Name: "READER", Value: "read-token-123"},
	}
	writeTokens := []config.Token{
		{Name: "WRITER", Value: "write-token-456"},
	}
	authSvc := service.NewAuthService(readTokens, writeTokens)

	tests := []struct {
		name        string
		authHeader  string
		expectError bool
	}{
		{
			name:        "valid write token",
			authHeader:  "Bearer write-token-456",
			expectError: false,
		},
		{
			name:        "read token cannot write",
			authHeader:  "Bearer read-token-123",
			expectError: true,
		},
		{
			name:        "invalid bearer token",
			authHeader:  "Bearer invalid-token",
			expectError: true,
		},
		{
			name:        "empty auth header",
			authHeader:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authSvc.AuthenticateWriteRequest(context.Background(), tt.authHeader)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

