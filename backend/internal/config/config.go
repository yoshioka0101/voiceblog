package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Port               string
	LogLevel           slog.Level
	DBDSN              string
	GoogleClientID     string
	GeminiAPIKey       string
	TokenEncryptionKey string
}

func Load() (Config, error) {
	port := getEnv("PORT", "8080")
	if err := validatePort(port); err != nil {
		return Config{}, err
	}

	logLevel, err := parseLogLevel(getEnv("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}

	dbDSN := getEnv("DB_DSN", "")
	if dbDSN == "" {
		return Config{}, fmt.Errorf("DB_DSN is required")
	}

	googleClientID := getEnv("GOOGLE_CLIENT_ID", "")
	if googleClientID == "" {
		return Config{}, fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}

	return Config{
		Port:               port,
		LogLevel:           logLevel,
		DBDSN:              dbDSN,
		GoogleClientID:     googleClientID,
		GeminiAPIKey:       getEnv("GEMINI_API_KEY", ""),
		TokenEncryptionKey: getEnv("TOKEN_ENCRYPTION_KEY", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseLogLevel(value string) (slog.Level, error) {
	level := slog.LevelInfo
	if err := level.UnmarshalText([]byte(value)); err != nil {
		return slog.LevelInfo, fmt.Errorf("invalid LOG_LEVEL: %w", err)
	}
	return level, nil
}

func validatePort(value string) error {
	_, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid PORT: %w", err)
	}
	return nil
}
