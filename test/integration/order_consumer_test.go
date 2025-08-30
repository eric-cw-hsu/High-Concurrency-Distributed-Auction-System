package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	pgorder "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/order"
	kafkaconsumer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/consumer"
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil"
)

// OrderConsumerTestSuite tests the Kafka order consumer functionality
type OrderConsumerTestSuite struct {
	suite.Suite
	dbHelper    *testutil.DatabaseTestHelper
	kafkaHelper *testutil.KafkaTestHelper
	orderRepo   order.OrderRepository
	ctx         context.Context
	cancelFunc  context.CancelFunc
}

// SetupSuite sets up the test suite
func (suite *OrderConsumerTestSuite) SetupSuite() {
	suite.ctx, suite.cancelFunc = context.WithCancel(context.Background())

	// Setup database
	var err error
	suite.dbHelper, err = testutil.NewDatabaseTestHelper(suite.ctx)
	suite.Require().NoError(err, "Failed to setup database helper")

	// Run migrations
	err = suite.dbHelper.RunMigrations()
	suite.Require().NoError(err, "Failed to run migrations")

	// Setup Kafka helper
	suite.kafkaHelper, err = testutil.NewKafkaTestHelper(suite.ctx, []string{"localhost:9092"})
	suite.Require().NoError(err, "Failed to setup Kafka helper")

	// Setup order repository
	suite.orderRepo = pgorder.NewPostgresOrderRepository(suite.dbHelper.DB)
}

// TearDownSuite cleans up the test suite
func (suite *OrderConsumerTestSuite) TearDownSuite() {
	if suite.kafkaHelper != nil {
		suite.kafkaHelper.Close()
	}
	if suite.dbHelper != nil {
		suite.dbHelper.Close()
	}
	if suite.cancelFunc != nil {
		suite.cancelFunc()
	}
}

// SetupTest cleans the database before each test
func (suite *OrderConsumerTestSuite) SetupTest() {
	err := suite.dbHelper.CleanDatabase()
	suite.Require().NoError(err, "Failed to clean database")
}

// TestKafkaConsumer_PersistOrder tests that the Kafka consumer can successfully
// consume order events and persist them to the database
func (suite *OrderConsumerTestSuite) TestKafkaConsumer_PersistOrder() {
	testCtx, cancel := context.WithTimeout(suite.ctx, 15*time.Second)
	defer cancel()

	// Create a test topic using the Kafka helper
	topicName, err := suite.kafkaHelper.CreateOrderPlacedTestTopic()
	suite.Require().NoError(err, "Failed to create test topic")

	// Start the consumer in a goroutine
	var consumer kafkaconsumer.EventConsumer
	consumerReady := make(chan bool)
	consumerDone := make(chan error)

	go func() {
		reader := kafkaInfra.NewReader(
			suite.kafkaHelper.GetBrokers(),
			topicName,
			"order-consumer-group-test",
		)
		defer reader.Close()

		consumer = kafkaconsumer.NewOrderConsumer(reader, suite.orderRepo)

		// Signal that consumer is ready
		close(consumerReady)

		// Start consuming
		err := consumer.Start(testCtx)
		consumerDone <- err
	}()

	// Wait for consumer to be ready
	<-consumerReady
	time.Sleep(1 * time.Second) // Give consumer time to connect

	// Create test order event
	orderEvent := order.OrderPlacedEvent{
		OrderId:    uuid.NewString(),
		StockId:    "test-stock-" + uuid.New().String()[:8],
		BuyerId:    "test-buyer-" + uuid.New().String()[:8],
		Quantity:   2,
		TotalPrice: 100.0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Timestamp:  time.Now(),
	}

	// Publish the order event to Kafka
	err = suite.publishOrderEvent(testCtx, topicName, orderEvent)
	suite.Require().NoError(err, "Failed to publish order event")

	// Wait for the message to be consumed and processed
	time.Sleep(3 * time.Second)

	// Stop the consumer
	if consumer != nil {
		consumer.Stop()
	}
	cancel()

	// Wait for consumer to finish
	select {
	case err := <-consumerDone:
		// Consumer finished (expected due to context cancellation)
		suite.T().Logf("Consumer finished: %v", err)
	case <-time.After(5 * time.Second):
		suite.T().Log("Consumer didn't finish within timeout")
	}

	// Verify that the order was persisted in the database
	suite.verifyOrderInDatabase(orderEvent.OrderId, orderEvent)
} // publishOrderEvent publishes an order event to the specified Kafka topic
func (suite *OrderConsumerTestSuite) publishOrderEvent(ctx context.Context, topic string, event order.OrderPlacedEvent) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.OrderId),
		Value: payload,
	})
}

// verifyOrderInDatabase verifies that the order exists in the database with correct data
func (suite *OrderConsumerTestSuite) verifyOrderInDatabase(orderID string, expectedOrder order.OrderPlacedEvent) {
	// Check if order exists
	var count int
	row := suite.dbHelper.DB.QueryRow(`SELECT COUNT(*) FROM orders WHERE order_id = $1`, orderID)
	err := row.Scan(&count)
	suite.Require().NoError(err, "Failed to query order count")
	suite.Assert().Equal(1, count, "Order should exist in database")

	// Verify order details
	var dbOrder struct {
		OrderID    string
		StockID    string
		BuyerID    string
		Quantity   int
		TotalPrice float64
	}

	query := `
		SELECT order_id, stock_id, buyer_id, quantity, total_price 
		FROM orders 
		WHERE order_id = $1
	`
	row = suite.dbHelper.DB.QueryRow(query, orderID)
	err = row.Scan(&dbOrder.OrderID, &dbOrder.StockID, &dbOrder.BuyerID, &dbOrder.Quantity, &dbOrder.TotalPrice)
	suite.Require().NoError(err, "Failed to query order details")

	// Assert order fields
	assert.Equal(suite.T(), expectedOrder.OrderId, dbOrder.OrderID)
	assert.Equal(suite.T(), expectedOrder.StockId, dbOrder.StockID)
	assert.Equal(suite.T(), expectedOrder.BuyerId, dbOrder.BuyerID)
	assert.Equal(suite.T(), expectedOrder.Quantity, dbOrder.Quantity)
	assert.Equal(suite.T(), expectedOrder.TotalPrice, dbOrder.TotalPrice)
}

// TestOrderConsumerTestSuite runs the test suite
func TestOrderConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(OrderConsumerTestSuite))
}
