package config

import "time"

type PostgresConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int
	ConnLifetime time.Duration
}

// LoadPostgresConfig reads from env and returns a PostgresConfig
func LoadPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:         getEnv("POSTGRES_HOST", "localhost"),
		Port:         getEnvAsInt("POSTGRES_PORT", 5432),
		User:         getEnv("POSTGRES_USER", "postgres"),
		Password:     getEnv("POSTGRES_PASSWORD", "password"),
		DBName:       getEnv("POSTGRES_DB", "auction"),
		MaxOpenConns: getEnvAsInt("POSTGRES_MAX_OPEN_CONNS", 100),
		MaxIdleConns: getEnvAsInt("POSTGRES_MAX_IDLE_CONNS", 50),
		ConnLifetime: getEnvAsDuration("POSTGRES_CONN_MAX_LIFETIME", 30*time.Minute),
	}
}
