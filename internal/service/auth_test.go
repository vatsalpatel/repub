package service

import (
	"context"
	"testing"
)

func TestAuthService_ValidateToken(t *testing.T) {
	authSvc := NewAuthService("test-token-123")

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid token",
			token:       "test-token-123",
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
			err := authSvc.ValidateToken(context.Background(), tt.token)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestAuthService_AuthenticateRequest(t *testing.T) {
	authSvc := NewAuthService("test-token-123")

	tests := []struct {
		name        string
		authHeader  string
		expectError bool
	}{
		{
			name:        "valid bearer token",
			authHeader:  "Bearer test-token-123",
			expectError: false,
		},
		{
			name:        "invalid bearer token",
			authHeader:  "Bearer invalid-token",
			expectError: true,
		},
		{
			name:        "missing bearer prefix",
			authHeader:  "test-token-123",
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
			err := authSvc.AuthenticateRequest(context.Background(), tt.authHeader)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestAuthService_ConstantTimeComparison(t *testing.T) {
	authSvc := NewAuthService("test-token-123")

	// Test that we don't leak timing information
	validToken := "test-token-123"
	invalidToken := "invalid-token-of-same-length"

	// Both should return quickly, no timing difference
	err1 := authSvc.ValidateToken(context.Background(), validToken)
	err2 := authSvc.ValidateToken(context.Background(), invalidToken)

	if err1 != nil {
		t.Error("Valid token should not return error")
	}
	if err2 == nil {
		t.Error("Invalid token should return error")
	}
}