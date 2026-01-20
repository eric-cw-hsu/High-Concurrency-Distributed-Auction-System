package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Database           DatabaseConfig
	Redis              RedisConfig
	Kafka              KafkaConfig
	GRPC               GRPCConfig
	Logger             LoggerConfig
	Outbox             OutboxConfig
	OrderTimeoutWorker OrderTimeoutWorkerConfig
}

func Load() (*Config, error) {
	cfg := &Config{
		Database:           loadDatabaseConfig(),
		Redis:              loadRedisConfig(),
		Kafka:              loadKafkaConfig(),
		GRPC:               loadGRPCConfig(),
		Logger:             loadLoggerConfig(),
		Outbox:             loadOutboxConfig(),
		OrderTimeoutWorker: loadOrderTimeoutWorkerConfig(),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	validators := []interface{ Validate() error }{
		&c.Database,
		&c.Redis,
		&c.Kafka,
		&c.GRPC,
		&c.Outbox,
		&c.OrderTimeoutWorker,
	}

	for _, v := range validators {
		if err := v.Validate(); err != nil {
			return err
		}
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
