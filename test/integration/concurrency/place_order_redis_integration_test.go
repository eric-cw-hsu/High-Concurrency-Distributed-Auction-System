package integration_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	redisInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/redis"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/redis/stockcache"
	orderUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/order"
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil/mocks"
)

// RedisIntegrationTestSuite tests concurrent behavior with real Redis
type RedisIntegrationTestSuite struct {
	suite.Suite
	redisClient   *redis.Client
	stockCache    *stockcache.RedisStockCache
	walletService *mocks.MockWalletService
	producer      *mocks.MockOrderProducer
	placeOrderUC  *orderUsecase.PlaceOrderUsecase
	ctx           context.Context
}

func (suite *RedisIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Initialize Redis client
	redisCfg := config.LoadRedisConfig()
	suite.redisClient = redisInfra.NewRedisClient(redisCfg)

	// Test Redis connection
	_, err := suite.redisClient.Ping(suite.ctx).Result()
	if err != nil {
		suite.T().Skipf("Redis is not available: %v", err)
		return
	}

	// Initialize real Redis stock cache
	suite.stockCache, err = stockcache.NewRedisStockCache(suite.redisClient)
	if err != nil {
		suite.T().Fatalf("Failed to create Redis stock cache: %v", err)
	}

	// Use mocks for other services
	suite.walletService = mocks.NewMockWalletService()
	suite.producer = mocks.NewMockOrderProducer()

	suite.placeOrderUC = orderUsecase.NewPlaceOrderUsecase(
		suite.producer,
		suite.stockCache,
		suite.walletService,
	)
}

func (suite *RedisIntegrationTestSuite) SetupTest() {
	suite.producer.Reset()
}

func (suite *RedisIntegrationTestSuite) TearDownSuite() {
	if suite.redisClient != nil {
		suite.redisClient.Close()
	}
}

func (suite *RedisIntegrationTestSuite) TestRedisIntegration_ConcurrentOrders() {
	if suite.redisClient == nil {
		suite.T().Skip("Redis not available, skipping integration test")
	}

	t := suite.T()

	// Arrange
	stockId := "redis-test-stock-1"
	initialStock := 25
	price := 100.0
	numConcurrentUsers := 100

	// Initialize stock in Redis
	err := suite.stockCache.SetStock(suite.ctx, stockId, initialStock)
	require.NoError(t, err)

	err = suite.stockCache.SetPrice(suite.ctx, stockId, price)
	require.NoError(t, err)

	// Set sufficient balance for all users
	for i := 0; i < numConcurrentUsers; i++ {
		userId := fmt.Sprintf("redis-user-%d", i)
		suite.walletService.SetBalance(userId, 1000.0)
	}

	// Act - Execute concurrent orders
	var wg sync.WaitGroup
	results := make(chan error, numConcurrentUsers)
	successCount := int64(0)

	startTime := time.Now()

	for i := 0; i < numConcurrentUsers; i++ {
		wg.Add(1)
		go func(userIndex int) {
			defer wg.Done()

			command := order.PlaceOrderCommand{
				BuyerId:  fmt.Sprintf("redis-user-%d", userIndex),
				StockId:  stockId,
				Quantity: 1,
			}

			err := suite.placeOrderUC.Execute(suite.ctx, command)
			results <- err

			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	close(results)

	executionTime := time.Since(startTime)

	// Assert
	var errors []error
	for err := range results {
		if err != nil {
			errors = append(errors, err)
		}
	}

	finalSuccessCount := atomic.LoadInt64(&successCount)
	throughput := float64(finalSuccessCount) / executionTime.Seconds()

	t.Logf("Redis Integration Test Results:")
	t.Logf("  Initial Stock: %d", initialStock)
	t.Logf("  Concurrent Users: %d", numConcurrentUsers)
	t.Logf("  Successful Orders: %d", finalSuccessCount)
	t.Logf("  Failed Orders: %d", len(errors))
	t.Logf("  Execution Time: %v", executionTime)
	t.Logf("  Throughput: %.2f orders/sec", throughput)

	// Verify Redis state consistency
	finalStock, err := suite.stockCache.GetStock(suite.ctx, stockId)
	require.NoError(t, err)
	expectedFinalStock := initialStock - int(finalSuccessCount)
	assert.Equal(t, expectedFinalStock, finalStock, "Redis stock should be consistent")

	// Core verification
	assert.LessOrEqual(t, finalSuccessCount, int64(initialStock),
		"Successful orders should not exceed initial stock")

	// Performance assertions
	assert.Greater(t, throughput, 50.0, "Redis throughput should be reasonable")

	// Verify event publish count
	publishCount := suite.producer.GetPublishCount()
	assert.Equal(t, finalSuccessCount, publishCount,
		"Published events should equal successful orders")
}

// TestRedisIntegrationTestSuite runs the Redis integration test suite
func TestRedisIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration tests in short mode")
	}

	suite.Run(t, new(RedisIntegrationTestSuite))
}
