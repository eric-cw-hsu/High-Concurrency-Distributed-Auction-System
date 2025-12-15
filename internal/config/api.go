package config

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
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/auction_db"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}
