package stockcache

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisStockCache struct {
	rdb    *redis.Client
	script *redis.Script
	prefix string
}

//go:embed decrease_stock.lua
var scriptFS embed.FS

func NewRedisStockCache(rdb *redis.Client) (*RedisStockCache, error) {
	luaContent, err := scriptFS.ReadFile("decrease_stock.lua")
	if err != nil {
		return nil, fmt.Errorf("failed to read Lua script: %w", err)
	}

	return &RedisStockCache{
		rdb:    rdb,
		script: redis.NewScript(string(luaContent)),
		prefix: "stock:",
	}, nil
}

// DecreaseStock decreases the stock of an item by the specified quantity.
// It returns the timestamp of the operation if successful, or an error if the item is not found or out of stock.
func (sc *RedisStockCache) DecreaseStock(ctx context.Context, stockId string, quantity int) (int64, error) {
	key := sc.prefix + stockId

	raw, err := sc.script.Run(ctx, sc.rdb, []string{key}, quantity).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrease stock: %w", err)
	}

	result, ok := raw.([]interface{})
	if !ok || len(result) != 2 {
		return 0, fmt.Errorf("unexpected result format from Lua script: %v", raw)
	}

	status, okStatus := result[0].(int64)
	timestamp, okTimestamp := result[1].(int64)
	if !okStatus || !okTimestamp {
		return 0, fmt.Errorf("unexpected result types from Lua script: %v", result)
	}

	switch status {
	case 0:
		return timestamp, nil // Stock decreased successfully
	case -1:
		return 0, fmt.Errorf("item not found: %s", stockId)
	case -2:
		return 0, fmt.Errorf("out of stock for item: %s", stockId)
	default:
		return 0, fmt.Errorf("unexpected result from Lua script: %d", result)
	}
}

func (sc *RedisStockCache) RestoreStock(ctx context.Context, stockId string, quantity int) error {
	key := sc.prefix + stockId

	raw, err := sc.script.Run(ctx, sc.rdb, []string{key}, -1*quantity).Result()
	if err != nil {
		return fmt.Errorf("failed to restore stock: %w", err)
	}

	result, ok := raw.([]interface{})
	if !ok || len(result) != 2 {
		return fmt.Errorf("unexpected result format from Lua script: %v", raw)
	}

	status, okStatus := result[0].(int64)
	_, okTimestamp := result[1].(int64)
	if !okStatus || !okTimestamp {
		return fmt.Errorf("unexpected result types from Lua script: %v", result)
	}

	switch status {
	case 0:
		return nil // Stock decreased successfully
	case -1:
		return fmt.Errorf("item not found: %s", stockId)
	case -2:
		return fmt.Errorf("out of stock for item: %s", stockId)
	default:
		return fmt.Errorf("unexpected result from Lua script: %d", result)
	}
}

func (sc *RedisStockCache) GetStock(ctx context.Context, stockId string) (int, error) {
	key := sc.prefix + stockId

	stock, err := sc.rdb.Get(ctx, key).Int()
	if errors.Is(err, redis.Nil) {
		return 0, fmt.Errorf("item not found: %s", stockId)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get stock: %w", err)
	}

	return stock, nil
}

func (sc *RedisStockCache) SetStock(ctx context.Context, stockId string, quantity int) error {
	key := sc.prefix + stockId

	if err := sc.rdb.Set(ctx, key, quantity, 0).Err(); err != nil {
		return fmt.Errorf("failed to set stock: %w", err)
	}

	return nil
}

func (sc *RedisStockCache) SetPrice(ctx context.Context, stockId string, price float64) error {
	key := sc.prefix + "price:" + stockId

	if err := sc.rdb.Set(ctx, key, price, 0).Err(); err != nil {
		return fmt.Errorf("failed to set price: %w", err)
	}

	return nil
}

func (sc *RedisStockCache) GetPrice(ctx context.Context, stockId string) (float64, error) {
	key := sc.prefix + "price:" + stockId

	price, err := sc.rdb.Get(ctx, key).Float64()
	if errors.Is(err, redis.Nil) {
		return 0, fmt.Errorf("price not found for item: %s", stockId)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}

	return price, nil
}

func (sc *RedisStockCache) RemoveAll(ctx context.Context) error {
	keys, err := sc.rdb.Keys(ctx, sc.prefix+"*").Result()
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	if len(keys) == 0 {
		return nil // No keys to delete
	}

	if err := sc.rdb.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}
