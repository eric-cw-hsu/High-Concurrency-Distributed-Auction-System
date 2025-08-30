package config

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared"
	"github.com/sirupsen/logrus"
)

// LoggerConfig represents logger service configuration
type LoggerConfig struct {
	KafkaBroker    string
	LogStorageType string
	LogFilePath    string
	Topics         []string
	LogLevel       string
}

// Logger holds a logger instance for the logger service
type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new configured logger instance for the logger service
func NewLogger(level string) *Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Set log level
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return &Logger{logger}
}

// LoadConfig loads logger configuration from environment variables
func LoadConfig() *LoggerConfig {
	// Load topics from environment variable
	topicsEnv := shared.GetEnv("KAFKA_TOPICS", "audit-logs,order-events,wallet-events,stock-events")
	topics := shared.ParseTopics(topicsEnv)

	return &LoggerConfig{
		KafkaBroker:    shared.GetEnv("KAFKA_BROKER", "localhost:9092"),
		LogStorageType: shared.GetEnv("LOG_STORAGE_TYPE", "file"),
		LogFilePath:    shared.GetEnv("LOG_FILE_PATH", "logs/audit.jsonl"),
		Topics:         topics,
		LogLevel:       shared.GetEnv("LOG_LEVEL", "info"),
	}
}

// GetFileDir returns the directory part of the log file path
func (c *LoggerConfig) GetFileDir() string {
	return shared.GetFileDir(c.LogFilePath)
}
