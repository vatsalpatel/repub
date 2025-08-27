package service

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strings"
)

type AuthService interface {
	ValidateToken(ctx context.Context, token string) error
	AuthenticateRequest(ctx context.Context, authHeader string) error
}

type authService struct {
	validToken string
}

func NewAuthService(token string) AuthService {
	return &authService{
		validToken: token,
	}
}

func (s *authService) ValidateToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// Use constant time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(token), []byte(s.validToken)) != 1 {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func (s *authService) AuthenticateRequest(ctx context.Context, authHeader string) error {
	if authHeader == "" {
		return fmt.Errorf("authorization header is required")
	}

	// Check for Bearer token format
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return fmt.Errorf("authorization header must start with 'Bearer '")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	return s.ValidateToken(ctx, token)
}