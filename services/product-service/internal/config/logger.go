package config

import "log/slog"

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level       string // "debug", "info", "warn", "error"
	Format      string // "json", "text"
	AddSource   bool
	Environment string // "development", "staging", "production"
}

func loadLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:       getEnv("LOG_LEVEL", "info"),
		Format:      getEnv("LOG_FORMAT", "json"),
		AddSource:   getEnv("LOG_ADD_SOURCE", "true") == "true",
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func (c LoggerConfig) GetSlogLevel() slog.Level {
	switch c.Level {
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
