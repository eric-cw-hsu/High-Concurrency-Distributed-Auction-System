package config

import "eric-cw-hsu.github.io/scalable-auction-system/internal/shared"

// APIConfig represents API service configuration
type APIConfig struct {
	Port        string
	JWTSecret   string
	DatabaseURL string
	RedisURL    string
	KafkaBroker string
	Environment string
}

// LoadAPIConfig loads API configuration from environment variables
func LoadAPIConfig() *APIConfig {
	return &APIConfig{
		Port:        shared.GetEnv("PORT", "8080"),
		JWTSecret:   shared.GetEnv("JWT_SECRET", "your-secret-key"),
		DatabaseURL: shared.GetEnv("DATABASE_URL", "postgres://localhost/auction_db"),
		RedisURL:    shared.GetEnv("REDIS_URL", "redis://localhost:6379"),
		KafkaBroker: shared.GetEnv("KAFKA_BROKER", "localhost:9092"),
		Environment: shared.GetEnv("ENVIRONMENT", "development"),
	}
}
