package config

type JWTConfig struct {
	SecretKey string
	Issuer    string
	ExpiresIn int // in seconds
}

func LoadJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey: getEnvMust("JWT_SECRET_KEY"),
		Issuer:    getEnv("JWT_ISSUER", "auction-service"),
		ExpiresIn: getEnvAsInt("JWT_EXPIRES_IN", 3600), // default to 1 hour
	}
}
