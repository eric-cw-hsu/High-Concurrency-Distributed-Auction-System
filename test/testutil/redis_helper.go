package testutil

import (
	"context"
	"fmt"
	"strconv"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/stock"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/redis/stockcache"
	"github.com/redis/go-redis/v9"
)

// RedisTestHelper provides common Redis testing utilities
type RedisTestHelper struct {
	Client *redis.Client
	ctx    context.Context
}

// NewRedisTestHelper creates a new Redis test helper
func NewRedisTestHelper(ctx context.Context) (*RedisTestHelper, error) {
	// Load test environment if not already loaded
	loadTestEnv() // Ignore error if already loaded

	// Setup Redis configuration
	redisHost := getEnvOrDefault("REDIS_HOST", "localhost")
	redisPort, _ := strconv.Atoi(getEnvOrDefault("REDIS_PORT", "6379"))
	redisDB, _ := strconv.Atoi(getEnvOrDefault("REDIS_DB", "0"))

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", redisHost, redisPort),
		DB:   redisDB,
	})

	// Test connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisTestHelper{
		Client: client,
		ctx:    ctx,
	}, nil
}

// CleanRedis flushes all keys from the test Redis database
func (h *RedisTestHelper) CleanRedis() error {
	return h.Client.FlushDB(h.ctx).Err()
}

// Close closes the Redis connection
func (h *RedisTestHelper) Close() error {
	if h.Client != nil {
		return h.Client.Close()
	}
	return nil
}

// GetStockCache returns a stock cache instance for testing
func (h *RedisTestHelper) GetStockCache() (stock.StockCache, error) {
	return stockcache.NewRedisStockCache(h.Client)
}
