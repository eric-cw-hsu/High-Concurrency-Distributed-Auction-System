package config

import (
	"log/slog"
	"strings"
)

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level       string
	Format      string
	AddSource   bool
	Environment string // "development", "staging", "production"
}

// loadLoggerConfig loads logger configuration
func loadLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:       getEnv("LOG_LEVEL", "info"),
		Format:      getEnv("LOG_FORMAT", "json"),
		AddSource:   getEnv("LOG_ADD_SOURCE", "true") == "true",
		Environment: getEnv("ENV", "development"),
	}
}

func (c LoggerConfig) GetSlogLevel() slog.Level {
	level := strings.ToLower(c.Level)
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
