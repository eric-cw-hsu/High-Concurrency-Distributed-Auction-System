package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/stock"
	kafkaConsumer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/consumer"
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil"
)

// StockConsumerTestSuite tests the stock consumer functionality
type StockConsumerTestSuite struct {
	suite.Suite
	ctx context.Context

	// Test infrastructure
	redisHelper *testutil.RedisTestHelper
	producer    *kafka.Writer

	// Services under test
	stockCache stock.StockCache
	consumer   kafkaConsumer.EventConsumer // Will implement this
}

func (suite *StockConsumerTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Setup Redis for stock cache
	var err error
	suite.redisHelper, err = testutil.NewRedisTestHelper(suite.ctx)
	suite.Require().NoError(err)

	// Setup Kafka producer for test events
	suite.producer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "stock.events.test",
		Balancer: &kafka.LeastBytes{},
	})

	// Initialize stock cache
	suite.stockCache, err = suite.redisHelper.GetStockCache()
	suite.Require().NoError(err)

	// Initialize mock consumer for testing
	suite.consumer = testutil.NewMockStockEventConsumer(suite.stockCache)
}

func (suite *StockConsumerTestSuite) TearDownSuite() {
	if suite.producer != nil {
		suite.producer.Close()
	}
	if suite.redisHelper != nil {
		suite.redisHelper.Close()
	}
}

func (suite *StockConsumerTestSuite) SetupTest() {
	// Clean Redis before each test
	suite.redisHelper.CleanRedis()
}

// TestStockConsumer_HandleOrderPlacedEvent tests basic stock update on order placed
func (suite *StockConsumerTestSuite) TestStockConsumer_HandleOrderPlacedEvent() {
	t := suite.T()

	// Arrange: Set initial stock
	stockId := "stock-123"
	initialStock := 100
	err := suite.stockCache.SetStock(suite.ctx, stockId, initialStock)
	require.NoError(t, err)

	// Create OrderPlacedEvent
	event := order.OrderPlacedEvent{
		OrderId:    "order-456",
		StockId:    stockId,
		BuyerId:    "buyer-789",
		Quantity:   5,
		TotalPrice: 500.0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Timestamp:  time.Now(),
	}

	// Act: Process the event through consumer
	err = suite.consumer.HandleEvent(suite.ctx, &event)
	require.NoError(t, err)

	// Assert: Stock should be decremented
	finalStock, err := suite.stockCache.GetStock(suite.ctx, stockId)
	require.NoError(t, err)
	assert.Equal(t, initialStock-event.Quantity, finalStock)
}

// TestStockConsumer_HandleOrderCancelledEvent tests stock restoration on order cancellation
func (suite *StockConsumerTestSuite) TestStockConsumer_HandleOrderCancelledEvent() {
	t := suite.T()

	// Arrange: Set reduced stock (simulating after order)
	stockId := "stock-123"
	currentStock := 95
	cancelledQuantity := 5
	err := suite.stockCache.SetStock(suite.ctx, stockId, currentStock)
	require.NoError(t, err)

	// Create OrderCancelledEvent
	event := order.OrderCancelledEvent{
		OrderId:   "order-456",
		StockId:   stockId,
		BuyerId:   "buyer-789",
		Quantity:  cancelledQuantity,
		Timestamp: time.Now(),
	}

	// Act: Process the cancellation event
	err = suite.consumer.HandleEvent(suite.ctx, &event)
	require.NoError(t, err)

	// Assert: Stock should be restored
	finalStock, err := suite.stockCache.GetStock(suite.ctx, stockId)
	require.NoError(t, err)
	assert.Equal(t, currentStock+cancelledQuantity, finalStock)
}

// TestStockConsumer_CrashRecovery tests the critical crash recovery scenario
func (suite *StockConsumerTestSuite) TestStockConsumer_CrashRecovery() {
	t := suite.T()

	// Arrange: Initial state
	stockId := "stock-123"
	initialStock := 100
	err := suite.stockCache.SetStock(suite.ctx, stockId, initialStock)
	require.NoError(t, err)

	// Simulate multiple pending events in Kafka queue
	pendingEvents := []order.OrderPlacedEvent{
		{OrderId: "order-1", StockId: stockId, Quantity: 5, Timestamp: time.Now()},
		{OrderId: "order-2", StockId: stockId, Quantity: 3, Timestamp: time.Now()},
		{OrderId: "order-3", StockId: stockId, Quantity: 2, Timestamp: time.Now()},
	}

	// Send events to Kafka (simulating events that were sent before crash)
	for _, event := range pendingEvents {
		eventBytes, _ := json.Marshal(event)
		err := suite.producer.WriteMessages(suite.ctx, kafka.Message{
			Key:   []byte(event.OrderId),
			Value: eventBytes,
		})
		require.NoError(t, err)
	}

	// Act: Simulate consumer restart with recovery
	err = suite.consumer.StartWithRecovery(suite.ctx)
	require.NoError(t, err)

	// Wait for all events to be processed
	time.Sleep(2 * time.Second)

	// Assert: Stock should reflect all processed events
	finalStock, err := suite.stockCache.GetStock(suite.ctx, stockId)
	require.NoError(t, err)

	expectedFinalStock := initialStock - (5 + 3 + 2) // All quantities deducted
	assert.Equal(t, expectedFinalStock, finalStock)
}

// TestStockConsumer_EventOrdering tests that events are processed in order
func (suite *StockConsumerTestSuite) TestStockConsumer_EventOrdering() {
	t := suite.T()

	// Arrange: Set initial stock
	stockId := "stock-123"
	initialStock := 10
	err := suite.stockCache.SetStock(suite.ctx, stockId, initialStock)
	require.NoError(t, err)

	// Create time-ordered events
	now := time.Now()
	events := []order.OrderPlacedEvent{
		{OrderId: "order-1", StockId: stockId, Quantity: 3, Timestamp: now.Add(1 * time.Second)},
		{OrderId: "order-2", StockId: stockId, Quantity: 2, Timestamp: now.Add(2 * time.Second)},
		{OrderId: "order-3", StockId: stockId, Quantity: 4, Timestamp: now.Add(3 * time.Second)},
	}

	// Act: Process events in different order (simulate out-of-order delivery)
	processOrder := []int{1, 0, 2} // Process order-2, order-1, order-3
	for _, idx := range processOrder {
		err := suite.consumer.HandleEvent(suite.ctx, &events[idx])
		require.NoError(t, err)
	}

	// Assert: Final stock should be correct regardless of processing order
	finalStock, err := suite.stockCache.GetStock(suite.ctx, stockId)
	require.NoError(t, err)

	expectedFinalStock := initialStock - (3 + 2 + 4)
	assert.Equal(t, expectedFinalStock, finalStock)
}

// TestStockConsumer_Idempotency tests that duplicate events don't cause issues
func (suite *StockConsumerTestSuite) TestStockConsumer_Idempotency() {
	t := suite.T()

	// Arrange: Set initial stock
	stockId := "stock-123"
	initialStock := 100
	err := suite.stockCache.SetStock(suite.ctx, stockId, initialStock)
	require.NoError(t, err)

	// Create an event
	event := order.OrderPlacedEvent{
		OrderId:    "order-456",
		StockId:    stockId,
		BuyerId:    "buyer-789",
		Quantity:   5,
		TotalPrice: 500.0,
		Timestamp:  time.Now(),
	}

	// Act: Process the same event multiple times
	err = suite.consumer.HandleEvent(suite.ctx, &event)
	require.NoError(t, err)

	err = suite.consumer.HandleEvent(suite.ctx, &event) // Duplicate
	require.NoError(t, err)

	err = suite.consumer.HandleEvent(suite.ctx, &event) // Another duplicate
	require.NoError(t, err)

	// Assert: Stock should be decremented only once
	finalStock, err := suite.stockCache.GetStock(suite.ctx, stockId)
	require.NoError(t, err)
	assert.Equal(t, initialStock-event.Quantity, finalStock)
}

// TestStockConsumer_ConcurrentEvents tests concurrent event processing
func (suite *StockConsumerTestSuite) TestStockConsumer_ConcurrentEvents() {
	t := suite.T()

	// Arrange: Set initial stock
	stockId := "stock-123"
	initialStock := 1000
	err := suite.stockCache.SetStock(suite.ctx, stockId, initialStock)
	require.NoError(t, err)

	// Create multiple events for concurrent processing
	numEvents := 100
	quantity := 5

	// Act: Process events concurrently
	errChan := make(chan error, numEvents)
	for i := 0; i < numEvents; i++ {
		go func(orderNum int) {
			event := order.OrderPlacedEvent{
				OrderId:   fmt.Sprintf("order-%d", orderNum),
				StockId:   stockId,
				Quantity:  quantity,
				Timestamp: time.Now(),
			}
			errChan <- suite.consumer.HandleEvent(suite.ctx, &event)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numEvents; i++ {
		err := <-errChan
		require.NoError(t, err)
	}

	// Assert: Final stock should be correct
	finalStock, err := suite.stockCache.GetStock(suite.ctx, stockId)
	require.NoError(t, err)

	expectedFinalStock := initialStock - (numEvents * quantity)
	assert.Equal(t, expectedFinalStock, finalStock)
}

func TestStockConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(StockConsumerTestSuite))
}
