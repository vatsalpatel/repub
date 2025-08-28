package service

import (
	"context"
	"fmt"
	"repub/internal/config"
	"strings"
)

type AuthService interface {
	ValidateReadToken(ctx context.Context, token string) error
	ValidateWriteToken(ctx context.Context, token string) error
	AuthenticateReadRequest(ctx context.Context, authHeader string) error
	AuthenticateWriteRequest(ctx context.Context, authHeader string) error
}

type authService struct {
	readTokens  map[string]struct{}
	writeTokens map[string]struct{}
}

func NewAuthService(readTokens, writeTokens []config.Token) AuthService {
	readMap := make(map[string]struct{})
	for _, token := range readTokens {
		readMap[token.Value] = struct{}{}
	}

	writeMap := make(map[string]struct{})
	for _, token := range writeTokens {
		writeMap[token.Value] = struct{}{}
	}

	return &authService{
		readTokens:  readMap,
		writeTokens: writeMap,
	}
}

func (s *authService) ValidateReadToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// Check both read and write tokens (write tokens can read too)
	if _, exists := s.readTokens[token]; exists {
		return nil
	}
	if _, exists := s.writeTokens[token]; exists {
		return nil
	}

	return fmt.Errorf("invalid token")
}

func (s *authService) ValidateWriteToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// Only write tokens can write
	if _, exists := s.writeTokens[token]; exists {
		return nil
	}

	return fmt.Errorf("invalid token")
}

func (s *authService) AuthenticateReadRequest(ctx context.Context, authHeader string) error {
	if authHeader == "" {
		return fmt.Errorf("authorization header is required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return fmt.Errorf("authorization header must start with 'Bearer '")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	return s.ValidateReadToken(ctx, token)
}

func (s *authService) AuthenticateWriteRequest(ctx context.Context, authHeader string) error {
	if authHeader == "" {
		return fmt.Errorf("authorization header is required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return fmt.Errorf("authorization header must start with 'Bearer '")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	return s.ValidateWriteToken(ctx, token)
}

