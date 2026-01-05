package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	ServiceName string
	Env         string

	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	Outbox   OutboxConfig
	Logger   LoggerConfig
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		ServiceName: getEnv("SERVICE_NAME", "product-service"),
		Env:         getEnv("ENV", "local"),
		Server:      loadServerConfig(),
		Database:    loadDatabaseConfig(),
		Kafka:       loadKafkaConfig(),
		Outbox:      loadOutboxConfig(),
		Logger:      loadLoggerConfig(),
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config: %w", err)
	}
	if err := c.Kafka.Validate(); err != nil {
		return fmt.Errorf("kafka config: %w", err)
	}
	return nil
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as int with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvDuration gets environment variable as duration with default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
