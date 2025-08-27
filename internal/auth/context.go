package auth

import (
	"context"
)

// contextKey is used for context keys to avoid collisions
type contextKey string

// AuthContextKey is the key used to store authentication status in request context
const AuthContextKey contextKey = "authenticated"

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	auth, ok := ctx.Value(AuthContextKey).(bool)
	if !ok {
		return false
	}
	return auth
}

// SetAuthenticated marks the request as authenticated in the context
func SetAuthenticated(ctx context.Context, authenticated bool) context.Context {
	return context.WithValue(ctx, AuthContextKey, authenticated)
}