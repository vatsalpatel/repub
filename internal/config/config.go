package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	StoragePath string
	Port        string
	BaseURL     string
	LogLevel    slog.Level
	AuthToken   string
}

func Load() *Config {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		fmt.Fprintln(os.Stderr, "ERROR: AUTH_TOKEN environment variable is required")
		os.Exit(1)
	}

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/repub?sslmode=disable"),
		StoragePath: getEnv("STORAGE_PATH", "/tmp/storage"),
		Port:        getEnv("PORT", "9090"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:9090"),
		LogLevel:    parseLogLevel(getEnv("LOG_LEVEL", "info")),
		AuthToken:   authToken,
	}
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
