package redis

import (
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient = redis.Client

func NewRedisClient(cfg config.RedisConfig) *RedisClient {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DB:       cfg.DB,
		Password: cfg.Password,
	})
}
