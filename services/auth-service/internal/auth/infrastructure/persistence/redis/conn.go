package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func MustConnect(addr, password string, db int) *redis.Client {
	zap.L().Info("connecting to redis",
		zap.String("addr", addr),
		zap.Int("db", db),
	)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		zap.L().Fatal("failed to connect to redis",
			zap.String("addr", addr),
			zap.Error(err),
		)
	}

	zap.L().Info("redis connected successfully")

	return client
}
