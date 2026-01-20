package config

import "fmt"

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

func loadRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvInt("REDIS_DB", 0),
		PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
	}
}

func (c *RedisConfig) Validate() error {
	if c.Addr == "" {
		return fmt.Errorf("redis address is required")
	}
	return nil
}
