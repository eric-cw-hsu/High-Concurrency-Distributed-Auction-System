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

	Kafka    KafkaConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Outbox   OutboxConfig
	Service  ServiceConfig
	Logger   LoggerConfig
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		ServiceName: getEnv("SERVICE_NAME", "stock-service"),
		Env:         getEnv("ENV", "local"),
		Server:      loadServerConfig(),
		Database:    loadDatabaseConfig(),
		Redis:       loadRedisConfig(),
		Outbox:      loadOutboxConfig(),
		Service:     loadServiceConfig(),
		Logger:      loadLoggerConfig(),
		Kafka:       loadKafkaConfig(),
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
	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("redis config: %w", err)
	}
	if err := c.Service.Validate(); err != nil {
		return fmt.Errorf("service config: %w", err)
	}
	if err := c.Kafka.Validate(); err != nil {
		return fmt.Errorf("kafka config: %w", err)
	}
	return nil
}

// Helper functions

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

// getEnvFloat gets environment variable as float64 with default value
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
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
