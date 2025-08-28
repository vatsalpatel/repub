package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	readTokenPrefix  = "READ_TOKEN_"
	writeTokenPrefix = "WRITE_TOKEN_"
)

type Config struct {
	DatabaseURL string
	StoragePath string
	Port        string
	BaseURL     string
	LogLevel    slog.Level
	ReadTokens  []Token
	WriteTokens []Token
}

type Token struct {
	Name  string
	Value string
}

func Load() *Config {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	readTokens := parseTokensFromEnv(readTokenPrefix)
	writeTokens := parseTokensFromEnv(writeTokenPrefix)

	if len(readTokens) == 0 && len(writeTokens) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: At least one READ_TOKEN_* or WRITE_TOKEN_* environment variable is required")
		os.Exit(1)
	}

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/repub?sslmode=disable"),
		StoragePath: getEnv("STORAGE_PATH", "/tmp/storage"),
		Port:        getEnv("PORT", "9090"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:9090"),
		LogLevel:    parseLogLevel(getEnv("LOG_LEVEL", "info")),
		ReadTokens:  readTokens,
		WriteTokens: writeTokens,
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

func parseTokensFromEnv(prefix string) []Token {
	var tokens []Token

	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				envName := parts[0]
				envValue := parts[1]
				name := strings.TrimPrefix(envName, prefix)
				tokens = append(tokens, Token{
					Name:  name,
					Value: envValue,
				})
			}
		}
	}

	return tokens
}
