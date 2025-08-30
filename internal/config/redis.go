package config

type RedisConfig struct {
	Host     string
	Port     int
	DB       int
	Password string
}

func LoadRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnvAsInt("REDIS_PORT", 6379),
		DB:       getEnvAsInt("REDIS_DB", 0),
		Password: getEnv("REDIS_PASSWORD", ""),
	}
}
