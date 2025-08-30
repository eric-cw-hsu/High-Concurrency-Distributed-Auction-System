package stockcache_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/redis/stockcache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func setupRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
		rdb.Close()
	})
	return rdb
}

func TestDecreaseStockItemNotFound(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)

	repo, err := stockcache.NewRedisStockCache(rdb)
	require.NoError(t, err)

	const itemID = "item-not-found"

	timestamp, err := repo.DecreaseStock(ctx, itemID, 1)
	require.Error(t, err)
	require.EqualError(t, err, "item not found: item-not-found")
	require.Equal(t, int64(0), timestamp)

	stock, err := repo.GetStock(ctx, itemID)
	require.Error(t, err)
	require.EqualError(t, err, "item not found: item-not-found")
	require.Equal(t, 0, stock)
}

func TestDecreaseStock(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)

	repo, err := stockcache.NewRedisStockCache(rdb)
	require.NoError(t, err)

	const itemID = "item-123"
	const key = "stock:" + itemID

	err = rdb.Set(ctx, key, 5, 0).Err()
	require.NoError(t, err)

	timestamp, err := repo.DecreaseStock(ctx, itemID, 3)
	require.NoError(t, err)

	stock, err := repo.GetStock(ctx, itemID)
	require.NoError(t, err)
	require.Equal(t, 2, stock)
	require.Greater(t, timestamp, int64(0), "timestamp should be greater than 0")

	timestamp, err = repo.DecreaseStock(ctx, itemID, 3)
	require.Error(t, err)
	require.EqualError(t, err, "out of stock for item: item-123")
	require.Equal(t, int64(0), timestamp)

	stock, err = repo.GetStock(ctx, itemID)
	require.NoError(t, err)
	require.Equal(t, 2, stock)
}

func TestDecreaseStockTimestampInorder(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)

	repo, err := stockcache.NewRedisStockCache(rdb)
	require.NoError(t, err)

	const itemID = "item-timestamp"
	const key = "stock:" + itemID

	err = rdb.Set(ctx, key, 5, 0).Err()
	require.NoError(t, err)

	timestamps := make([]int64, 0, 5)
	for i := 0; i < 5; i++ {
		timestamp, err := repo.DecreaseStock(ctx, itemID, 1)
		require.NoError(t, err)
		timestamps = append(timestamps, timestamp)
	}

	for i := 1; i < len(timestamps); i++ {
		require.GreaterOrEqual(t, timestamps[i], timestamps[i-1], "timestamps should be in non-decreasing order")
	}

	stock, err := repo.GetStock(ctx, itemID)
	require.NoError(t, err)
	require.Equal(t, 0, stock)
}

func TestConcurrentDecreaseStock(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)

	repo, err := stockcache.NewRedisStockCache(rdb)
	require.NoError(t, err)

	const itemID = "item-concurrent"
	const key = "stock:" + itemID

	err = rdb.Set(ctx, key, 1, 0).Err()
	require.NoError(t, err)

	const concurrentUsers = 10000
	results := make(chan bool, concurrentUsers)

	for i := 0; i < concurrentUsers; i++ {
		go func() {
			_, err := repo.DecreaseStock(ctx, itemID, 1)
			if err != nil {
				results <- false
				return
			}
			results <- true
		}()
	}

	successCount := 0
	for i := 0; i < concurrentUsers; i++ {
		if <-results {
			successCount++
		}
	}

	require.Equal(t, 1, successCount)

	stock, err := repo.GetStock(ctx, itemID)
	require.NoError(t, err)
	require.Equal(t, 0, stock)
}

func TestConcurrentDecreaseUntilDepleted(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)

	repo, err := stockcache.NewRedisStockCache(rdb)
	require.NoError(t, err)

	const itemID = "item-flash-sale"
	const key = "stock:" + itemID
	const initialStock = 100
	const concurrentUsers = 10000

	err = rdb.Set(ctx, key, initialStock, 0).Err()
	require.NoError(t, err)

	var mu sync.Mutex
	totalSuccess := 0

	var wg sync.WaitGroup
	wg.Add(concurrentUsers)

	for i := 0; i < concurrentUsers; i++ {
		go func() {
			defer wg.Done()

			qty := rand.Intn(3) + 1 // Randomize quantity between 1 and 3
			_, err := repo.DecreaseStock(ctx, itemID, qty)
			if err == nil {
				mu.Lock()
				totalSuccess += qty
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	stock, err := repo.GetStock(ctx, itemID)
	require.NoError(t, err)

	require.LessOrEqual(t, totalSuccess, initialStock, "total success should not exceed initial stock")
	require.Equal(t, initialStock-totalSuccess, stock, "remaining stock should equal initial stock minus total success")
}
