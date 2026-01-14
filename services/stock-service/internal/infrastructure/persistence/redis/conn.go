package redis

import (
	"context"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// MustConnect connects to Redis or panics
func MustConnect(cfg config.RedisConfig) *redis.Client {
	zap.L().Info("connecting to redis",
		zap.String("addr", cfg.GetAddr()),
		zap.Int("db", cfg.DB),
	)

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		zap.L().Fatal("failed to connect to redis",
			zap.String("addr", cfg.GetAddr()),
			zap.Error(err),
		)
	}

	zap.L().Info("redis connected successfully")

	return client
}
