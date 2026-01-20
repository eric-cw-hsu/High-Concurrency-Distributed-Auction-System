package config

import (
	"fmt"
	"time"
)

type DatabaseConfig struct {
	Driver   string
	DSN      string
	MaxOpen  int
	MaxIdle  int
	Lifetime time.Duration
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:   getEnv("DB_DRIVER", "postgres"),
		DSN:      getEnv("DB_DSN", "postgres://user:pass@localhost:5432/order_db?sslmode=disable"),
		MaxOpen:  getEnvInt("DB_MAX_OPEN", 25),
		MaxIdle:  getEnvInt("DB_MAX_IDLE", 10),
		Lifetime: getEnvDuration("DB_LIFETIME", 5*time.Minute),
	}
}

func (c *DatabaseConfig) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}
	return nil
}
