package config

import (
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set required tokens for test
	if err := os.Setenv("READ_TOKEN_ALICE", "read-token-123"); err != nil {
		t.Fatalf("Failed to set READ_TOKEN_ALICE: %v", err)
	}
	if err := os.Setenv("WRITE_TOKEN_BOB", "write-token-456"); err != nil {
		t.Fatalf("Failed to set WRITE_TOKEN_BOB: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("READ_TOKEN_ALICE"); err != nil {
			t.Errorf("Failed to unset READ token: %v", err)
		}
		if err := os.Unsetenv("WRITE_TOKEN_BOB"); err != nil {
			t.Errorf("Failed to unset write token: %v", err)
		}
	}()

	cfg := Load()

	if cfg.DatabaseURL != "postgres://localhost/repub?sslmode=disable" {
		t.Errorf("Expected default database URL, got %s", cfg.DatabaseURL)
	}

	if cfg.StoragePath != "/tmp/storage" {
		t.Errorf("Expected default storage path, got %s", cfg.StoragePath)
	}

	if cfg.Port != "9090" {
		t.Errorf("Expected default port 9090, got %s", cfg.Port)
	}

	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("Expected default log level info, got %v", cfg.LogLevel)
	}

	if len(cfg.ReadTokens) != 1 || cfg.ReadTokens[0].Name != "ALICE" || cfg.ReadTokens[0].Value != "read-token-123" {
		t.Errorf("Expected ReadTokens to contain ALICE token, got %v", cfg.ReadTokens)
	}

	if len(cfg.WriteTokens) != 1 || cfg.WriteTokens[0].Name != "BOB" || cfg.WriteTokens[0].Value != "write-token-456" {
		t.Errorf("Expected WriteTokens to contain BOB token, got %v", cfg.WriteTokens)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	if err := os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test"); err != nil {
		t.Fatalf("Failed to set DATABASE_URL: %v", err)
	}
	if err := os.Setenv("STORAGE_PATH", "/tmp/storage"); err != nil {
		t.Fatalf("Failed to set STORAGE_PATH: %v", err)
	}
	if err := os.Setenv("PORT", "8080"); err != nil {
		t.Fatalf("Failed to set PORT: %v", err)
	}
	if err := os.Setenv("BASE_URL", "https://example.com"); err != nil {
		t.Fatalf("Failed to set BASE_URL: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL: %v", err)
	}
	if err := os.Setenv("READ_TOKEN_ALICE", "alice-read-123"); err != nil {
		t.Fatalf("Failed to set READ_TOKEN_ALICE: %v", err)
	}
	if err := os.Setenv("WRITE_TOKEN_BOB", "bob-write-456"); err != nil {
		t.Fatalf("Failed to set WRITE_TOKEN_BOB: %v", err)
	}
	if err := os.Setenv("WRITE_TOKEN_CHARLIE", "charlie-write-789"); err != nil {
		t.Fatalf("Failed to set WRITE_TOKEN_CHARLIE: %v", err)
	}

	defer func() {
		envVars := []string{"DATABASE_URL", "STORAGE_PATH", "PORT", "BASE_URL", "LOG_LEVEL", "READ_TOKEN_ALICE", "WRITE_TOKEN_BOB", "WRITE_TOKEN_CHARLIE"}
		for _, env := range envVars {
			if err := os.Unsetenv(env); err != nil {
				t.Errorf("Failed to unset %s: %v", env, err)
			}
		}
	}()

	cfg := Load()

	if cfg.DatabaseURL != "postgres://test:test@localhost/test" {
		t.Errorf("Expected env database URL, got %s", cfg.DatabaseURL)
	}

	if cfg.StoragePath != "/tmp/storage" {
		t.Errorf("Expected env storage path, got %s", cfg.StoragePath)
	}

	if cfg.Port != "8080" {
		t.Errorf("Expected env port 8080, got %s", cfg.Port)
	}

	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("Expected debug log level, got %v", cfg.LogLevel)
	}

	if len(cfg.ReadTokens) != 1 || cfg.ReadTokens[0].Name != "ALICE" || cfg.ReadTokens[0].Value != "alice-read-123" {
		t.Errorf("Expected ReadTokens to contain ALICE token, got %v", cfg.ReadTokens)
	}

	if len(cfg.WriteTokens) != 2 {
		t.Errorf("Expected 2 WriteTokens, got %d", len(cfg.WriteTokens))
	}
}

func TestParseTokensFromEnv(t *testing.T) {
	// Clean up any existing tokens
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "TEST_TOKEN_") {
			key := strings.SplitN(env, "=", 2)[0]
			os.Unsetenv(key)
		}
	}

	// Set test tokens
	testEnvs := map[string]string{
		"TEST_TOKEN_ALICE":   "alice-token-123",
		"TEST_TOKEN_BOB":     "bob-token-456",
		"TEST_TOKEN_CHARLIE": "charlie-token-789",
		"OTHER_VAR":          "should-not-match",
	}

	for key, value := range testEnvs {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}

	defer func() {
		for key := range testEnvs {
			if err := os.Unsetenv(key); err != nil {
				t.Errorf("Failed to unset %s: %v", key, err)
			}
		}
	}()

	tokens := parseTokensFromEnv("TEST_TOKEN_")

	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(tokens))
	}

	expectedTokens := map[string]string{
		"ALICE":   "alice-token-123",
		"BOB":     "bob-token-456", 
		"CHARLIE": "charlie-token-789",
	}

	tokenMap := make(map[string]string)
	for _, token := range tokens {
		tokenMap[token.Name] = token.Value
	}

	for expectedName, expectedValue := range expectedTokens {
		if value, exists := tokenMap[expectedName]; !exists {
			t.Errorf("Expected token %s not found", expectedName)
		} else if value != expectedValue {
			t.Errorf("Expected token %s to have value %s, got %s", expectedName, expectedValue, value)
		}
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"invalid", slog.LevelInfo}, // Default to info for invalid levels
		{"", slog.LevelInfo},        // Default to info for empty string
	}

	for _, test := range tests {
		result := parseLogLevel(test.input)
		if result != test.expected {
			t.Errorf("parseLogLevel(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}
