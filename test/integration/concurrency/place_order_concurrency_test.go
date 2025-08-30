package integration_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	orderUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/order"
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil/mocks"
)

// PlaceOrderConcurrencyTestSuite tests concurrent behavior of PlaceOrderUsecase
type PlaceOrderConcurrencyTestSuite struct {
	suite.Suite
	stockCache    *mocks.MockStockCache
	walletService *mocks.MockWalletService
	producer      *mocks.MockOrderProducer
	placeOrderUC  *orderUsecase.PlaceOrderUsecase
	ctx           context.Context
}

func (suite *PlaceOrderConcurrencyTestSuite) SetupSuite() {
	suite.stockCache = mocks.NewMockStockCache()
	suite.walletService = mocks.NewMockWalletService()
	suite.producer = mocks.NewMockOrderProducer()
	suite.ctx = context.Background()

	suite.placeOrderUC = orderUsecase.NewPlaceOrderUsecase(
		suite.producer,
		suite.stockCache,
		suite.walletService,
	)
}

func (suite *PlaceOrderConcurrencyTestSuite) SetupTest() {
	// Reset mock counters
	suite.producer.Reset()
}

func (suite *PlaceOrderConcurrencyTestSuite) TestConcurrentOrders_SameStock_InventoryRaceCondition() {
	// This test verifies whether inventory deduction is correct when multiple users compete for the same product simultaneously
	t := suite.T()

	// Arrange
	stockId := "concurrent-test-stock-1"
	initialStock := 10
	price := 100.0
	orderQuantity := 1
	numConcurrentUsers := 50 // 50 users competing for 10 products

	suite.stockCache.SetInitialStock(stockId, initialStock, price)

	// Set sufficient balance for each user
	for i := 0; i < numConcurrentUsers; i++ {
		userId := fmt.Sprintf("concurrent-user-%d", i)
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
				BuyerId:  fmt.Sprintf("concurrent-user-%d", userIndex),
				StockId:  stockId,
				Quantity: orderQuantity,
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

	t.Logf("Concurrency Test Results:")
	t.Logf("  Initial Stock: %d", initialStock)
	t.Logf("  Concurrent Users: %d", numConcurrentUsers)
	t.Logf("  Successful Orders: %d", finalSuccessCount)
	t.Logf("  Failed Orders: %d", len(errors))
	t.Logf("  Execution Time: %v", executionTime)
	t.Logf("  Throughput: %.2f orders/sec", throughput)

	// Core business logic verification
	assert.LessOrEqual(t, finalSuccessCount, int64(initialStock),
		"Successful orders should not exceed initial stock")

	// Verify that exactly the initial stock was consumed
	finalStock := suite.stockCache.GetCurrentStock(stockId)
	expectedFinalStock := initialStock - int(finalSuccessCount)
	assert.Equal(t, expectedFinalStock, finalStock,
		"Final stock should equal initial stock minus successful orders")

	// Verify event publish count matches successful orders
	publishCount := suite.producer.GetPublishCount()
	assert.Equal(t, finalSuccessCount, publishCount,
		"Published events should equal successful orders")

	// Performance verification - should process quickly with mocks
	assert.Greater(t, throughput, 100.0, "Mock throughput should be high")
}

// TestPlaceOrderConcurrencyTestSuite runs the concurrency test suite
func TestPlaceOrderConcurrencyTestSuite(t *testing.T) {
	suite.Run(t, new(PlaceOrderConcurrencyTestSuite))
}
