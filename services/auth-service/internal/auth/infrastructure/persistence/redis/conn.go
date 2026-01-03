package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func MustConnect(addr, password string, db int) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	return redisClient
}
