package integration_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	walletRepo "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/wallet"
	redisStock "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/redis/stockcache"
	orderUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/order"
	walletUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil"
)

// IntegrationTestSuite is the main integration test suite
type IntegrationTestSuite struct {
	suite.Suite

	// Test infrastructure
	dbHelper    *testutil.DatabaseTestHelper
	redisHelper *testutil.RedisTestHelper
	userHelper  *testutil.UserTestHelper

	// Service dependencies
	walletRepo     *walletRepo.PostgresWalletRepository
	walletService  walletUsecase.WalletService
	addFundUsecase *walletUsecase.AddFundUsecase
	stockCache     *redisStock.RedisStockCache
	producer       *testutil.MockOrderProducer
	placeOrderUC   *orderUsecase.PlaceOrderUsecase

	ctx context.Context
}

// SetupSuite sets up the test suite
func (suite *IntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Setup database
	var err error
	suite.dbHelper, err = testutil.NewDatabaseTestHelper(suite.ctx)
	suite.Require().NoError(err, "Failed to setup database helper")

	// Run migrations
	suite.T().Log("Running database migrations...")
	err = suite.dbHelper.RunMigrations()
	suite.Require().NoError(err, "Failed to run database migrations")

	// Setup Redis
	suite.redisHelper, err = testutil.NewRedisTestHelper(suite.ctx)
	suite.Require().NoError(err, "Failed to setup Redis helper")

	// Initialize services
	suite.walletRepo = walletRepo.NewPostgresWalletRepository(suite.dbHelper.DB)
	mockEventPub := testutil.NewMockWalletEventPublisher()
	suite.walletService = walletUsecase.NewWalletService(suite.walletRepo, mockEventPub)
	suite.addFundUsecase = walletUsecase.NewAddFundUsecase(suite.walletRepo, mockEventPub)

	suite.stockCache, err = redisStock.NewRedisStockCache(suite.redisHelper.Client)
	suite.Require().NoError(err, "Failed to create stock cache")

	suite.producer = testutil.NewMockOrderProducer()
	suite.placeOrderUC = orderUsecase.NewPlaceOrderUsecase(
		suite.producer,
		suite.stockCache,
		suite.walletService,
	)

	// Setup user helper
	suite.userHelper = testutil.NewUserTestHelper(suite.ctx, suite.dbHelper.DB, suite.walletService, suite.addFundUsecase)

	suite.T().Log("Integration test suite setup completed")
}

// TearDownSuite cleans up after the test suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.redisHelper != nil {
		suite.redisHelper.Close()
	}
	if suite.dbHelper != nil {
		suite.dbHelper.Close()
	}
}

// SetupTest prepares for each test
func (suite *IntegrationTestSuite) SetupTest() {
	// Clean database and Redis before each test
	suite.dbHelper.CleanDatabase()
	suite.redisHelper.CleanRedis()
	suite.producer.Reset()
}

// TestFullIntegrationAuctionRush tests auction performance with configurable thresholds
func (suite *IntegrationTestSuite) TestFullIntegrationAuctionRush() {
	config := testutil.GetDefaultTestConfiguration()
	thresholds := testutil.GetDefaultPerformanceThresholds()

	result := suite.runAuctionPerformanceTest(config)
	suite.logPerformanceResults(result, "Default Configuration")

	// Check performance thresholds
	if !result.MeetsThresholds(thresholds) {
		failures := result.GetFailedThresholds(thresholds)
		suite.T().Logf("Performance thresholds not met:")
		for _, failure := range failures {
			suite.T().Logf("  - %s", failure)
		}
		suite.Fail("Performance test failed to meet thresholds")
	}

	// Business logic assertions
	assert.Equal(suite.T(), result.StockConsumed, result.SuccessfulOrders,
		"Stock consumed should equal successful orders")
	assert.GreaterOrEqual(suite.T(), int(result.FinalStock), 0,
		"Stock should never go negative")
	assert.Equal(suite.T(), result.SuccessfulOrders, suite.producer.GetPublishCount(),
		"Kafka messages should equal successful orders")
}

// TestHighPerformanceAuctionRush tests with aggressive performance thresholds
func (suite *IntegrationTestSuite) TestHighPerformanceAuctionRush() {
	config := testutil.GetDefaultTestConfiguration()
	thresholds := testutil.GetHighPerformanceThresholds()

	result := suite.runAuctionPerformanceTest(config)
	suite.logPerformanceResults(result, "High Performance Configuration")

	// Check high performance thresholds (may fail on slower systems)
	if !result.MeetsThresholds(thresholds) {
		failures := result.GetFailedThresholds(thresholds)
		suite.T().Logf("High performance thresholds not met (this may be acceptable):")
		for _, failure := range failures {
			suite.T().Logf("  - %s", failure)
		}
		// Don't fail test for high performance - just log
		suite.T().Log("High performance test completed (check logs for threshold results)")
	} else {
		suite.T().Log("🏆 Excellent! System meets high performance thresholds")
	}
}

// TestStressAuctionRush tests system under stress conditions
func (suite *IntegrationTestSuite) TestStressAuctionRush() {
	config := testutil.GetStressTestConfiguration()
	// Use more relaxed thresholds for stress test
	thresholds := testutil.PerformanceThresholds{
		MinSuccessRate:    0.90, // 90% under stress
		MaxAverageLatency: 100 * time.Millisecond,
		MinThroughput:     100.0, // Lower throughput expectation
		MaxTestDuration:   10 * time.Second,
	}

	result := suite.runAuctionPerformanceTest(config)
	suite.logPerformanceResults(result, "Stress Test Configuration")

	// Check stress test thresholds
	if !result.MeetsThresholds(thresholds) {
		failures := result.GetFailedThresholds(thresholds)
		suite.T().Logf("Stress test thresholds not met:")
		for _, failure := range failures {
			suite.T().Logf("  - %s", failure)
		}
		suite.Fail("Stress test failed to meet minimum thresholds")
	}
}

// runAuctionPerformanceTest executes a performance test with the given configuration
func (suite *IntegrationTestSuite) runAuctionPerformanceTest(config testutil.TestConfiguration) testutil.PerformanceResult {
	// Setup test users with wallets
	testUsers, err := suite.userHelper.CreateTestUsersWithWallets(5, config.WalletBalance)
	suite.Require().NoError(err, "Failed to create test users")

	// Setup stock
	productId := "test-product-1"
	err = suite.stockCache.SetStock(suite.ctx, productId, int(config.InitialStock))
	suite.Require().NoError(err, "Failed to set initial stock")

	// Set stock price
	err = suite.stockCache.SetPrice(suite.ctx, productId, config.OrderPrice)
	suite.Require().NoError(err, "Failed to set stock price")

	// Test execution
	var wg sync.WaitGroup
	var successCount, failureCount int64
	totalOrders := int64(config.Concurrency * config.OrdersPerThread)

	startTime := time.Now()

	// Launch concurrent order placing
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(goroutineId int) {
			defer wg.Done()
			userIndex := goroutineId % len(testUsers)
			userId := testUsers[userIndex]

			for j := 0; j < config.OrdersPerThread; j++ {
				orderCmd := order.PlaceOrderCommand{
					StockId:  productId,
					BuyerId:  userId,
					Quantity: 1,
				}

				err := suite.placeOrderUC.Execute(suite.ctx, orderCmd)
				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					suite.T().Logf("Order failed (goroutine %d, order %d): %v", goroutineId, j, err)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Get final stock
	finalStock, err := suite.stockCache.GetStock(suite.ctx, productId)
	suite.Require().NoError(err, "Failed to get final stock")

	// Calculate metrics
	successfulOrders := atomic.LoadInt64(&successCount)
	failedOrders := atomic.LoadInt64(&failureCount)
	successRate := float64(successfulOrders) / float64(totalOrders)
	avgLatency := duration / time.Duration(totalOrders)
	throughput := float64(successfulOrders) / duration.Seconds()

	return testutil.PerformanceResult{
		Duration:         duration,
		TotalOrders:      totalOrders,
		SuccessfulOrders: successfulOrders,
		FailedOrders:     failedOrders,
		SuccessRate:      successRate,
		AverageLatency:   avgLatency,
		Throughput:       throughput,
		InitialStock:     config.InitialStock,
		FinalStock:       int64(finalStock),
		StockConsumed:    config.InitialStock - int64(finalStock),
	}
}

// logPerformanceResults logs detailed performance results
func (suite *IntegrationTestSuite) logPerformanceResults(result testutil.PerformanceResult, configName string) {
	suite.T().Logf("")
	suite.T().Logf("=== INTEGRATION AUCTION PERFORMANCE: %s ===", configName)
	suite.T().Logf("Setup: PostgreSQL + Redis + Mock Kafka")
	suite.T().Logf("Total Duration: %v", result.Duration)
	suite.T().Logf("Total Orders: %d", result.TotalOrders)
	suite.T().Logf("Successful: %d", result.SuccessfulOrders)
	suite.T().Logf("Failed: %d", result.FailedOrders)
	suite.T().Logf("Success Rate: %.2f%%", result.SuccessRate*100)
	suite.T().Logf("Average Latency: %v", result.AverageLatency)
	suite.T().Logf("Throughput: %.2f orders/sec", result.Throughput)
	suite.T().Logf("Initial Stock: %d", result.InitialStock)
	suite.T().Logf("Final Stock: %d", result.FinalStock)
	suite.T().Logf("Consumed: %d", result.StockConsumed)
	suite.T().Logf("")
}

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
