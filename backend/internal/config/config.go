package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Port     string
	LogLevel slog.Level
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

	return Config{
		Port:     port,
		LogLevel: logLevel,
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
