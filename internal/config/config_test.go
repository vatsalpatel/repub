package config

import (
	"log/slog"
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set required AUTH_TOKEN for test
	if err := os.Setenv("AUTH_TOKEN", "test-token"); err != nil {
		t.Fatalf("Failed to set AUTH_TOKEN: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("AUTH_TOKEN"); err != nil {
			t.Errorf("Failed to unset AUTH_TOKEN: %v", err)
		}
	}()
	
	// Test default values
	cfg := Load()

	if cfg.DatabaseURL != "postgres://localhost/repub?sslmode=disable" {
		t.Errorf("Expected default database URL, got %s", cfg.DatabaseURL)
	}

	if cfg.StoragePath != "./storage" {
		t.Errorf("Expected default storage path, got %s", cfg.StoragePath)
	}

	if cfg.Port != "9090" {
		t.Errorf("Expected default port 9090, got %s", cfg.Port)
	}

	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("Expected default log level info, got %v", cfg.LogLevel)
	}
	
	if cfg.AuthToken != "test-token" {
		t.Errorf("Expected AUTH_TOKEN test-token, got %s", cfg.AuthToken)
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
	if err := os.Setenv("PORT", "9090"); err != nil {
		t.Fatalf("Failed to set PORT: %v", err)
	}
	if err := os.Setenv("BASE_URL", "https://example.com"); err != nil {
		t.Fatalf("Failed to set BASE_URL: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL: %v", err)
	}
	if err := os.Setenv("AUTH_TOKEN", "env-test-token"); err != nil {
		t.Fatalf("Failed to set AUTH_TOKEN: %v", err)
	}

	defer func() {
		if err := os.Unsetenv("DATABASE_URL"); err != nil {
			t.Errorf("Failed to unset DATABASE_URL: %v", err)
		}
		if err := os.Unsetenv("STORAGE_PATH"); err != nil {
			t.Errorf("Failed to unset STORAGE_PATH: %v", err)
		}
		if err := os.Unsetenv("PORT"); err != nil {
			t.Errorf("Failed to unset PORT: %v", err)
		}
		if err := os.Unsetenv("BASE_URL"); err != nil {
			t.Errorf("Failed to unset BASE_URL: %v", err)
		}
		if err := os.Unsetenv("LOG_LEVEL"); err != nil {
			t.Errorf("Failed to unset LOG_LEVEL: %v", err)
		}
		if err := os.Unsetenv("AUTH_TOKEN"); err != nil {
			t.Errorf("Failed to unset AUTH_TOKEN: %v", err)
		}
	}()

	cfg := Load()

	if cfg.DatabaseURL != "postgres://test:test@localhost/test" {
		t.Errorf("Expected env database URL, got %s", cfg.DatabaseURL)
	}

	if cfg.StoragePath != "/tmp/storage" {
		t.Errorf("Expected env storage path, got %s", cfg.StoragePath)
	}

	if cfg.Port != "9090" {
		t.Errorf("Expected env port 9090, got %s", cfg.Port)
	}

	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("Expected debug log level, got %v", cfg.LogLevel)
	}
	
	if cfg.AuthToken != "env-test-token" {
		t.Errorf("Expected AUTH_TOKEN env-test-token, got %s", cfg.AuthToken)
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
