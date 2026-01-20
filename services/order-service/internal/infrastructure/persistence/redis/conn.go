package redis

import (
	"context"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func MustConnect(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Check connectivity
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		zap.L().Fatal("failed to connect to redis", zap.Error(err))
	}

	return client
}
