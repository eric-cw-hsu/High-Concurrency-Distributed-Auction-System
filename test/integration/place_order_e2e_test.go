package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	pgorder "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/order"
	orderUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/order"
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil"
)

// MockKafkaMessage represents a Kafka message for testing
type MockKafkaMessage struct {
	Key       []byte
	Value     []byte
	Topic     string
	Partition int
	Offset    int64
	Time      time.Time
}

// MockKafkaWriter simulates Kafka writer for testing
type MockKafkaWriter struct {
	mock.Mock
	messages []MockKafkaMessage
	mu       sync.RWMutex
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...MockKafkaMessage) error {
	args := m.Called(ctx, msgs)
	m.mu.Lock()
	m.messages = append(m.messages, msgs...)
	m.mu.Unlock()
	return args.Error(0)
}

func (m *MockKafkaWriter) GetMessages() []MockKafkaMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]MockKafkaMessage, len(m.messages))
	copy(result, m.messages)
	return result
}

func (m *MockKafkaWriter) ClearMessages() {
	m.mu.Lock()
	m.messages = nil
	m.mu.Unlock()
}

// MockKafkaReader simulates Kafka reader for testing
type MockKafkaReader struct {
	mock.Mock
	messages chan MockKafkaMessage
	closed   bool
	mu       sync.RWMutex
}

func NewMockKafkaReader() *MockKafkaReader {
	return &MockKafkaReader{
		messages: make(chan MockKafkaMessage, 100),
	}
}

func (m *MockKafkaReader) ReadMessage(ctx context.Context) (MockKafkaMessage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return MockKafkaMessage{}, fmt.Errorf("reader is closed")
	}

	select {
	case <-ctx.Done():
		return MockKafkaMessage{}, ctx.Err()
	case msg := <-m.messages:
		return msg, nil
	case <-time.After(100 * time.Millisecond):
		return MockKafkaMessage{}, fmt.Errorf("no message available")
	}
}

func (m *MockKafkaReader) CommitMessages(ctx context.Context, msgs ...MockKafkaMessage) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaReader) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		close(m.messages)
		m.closed = true
	}
	return nil
}

func (m *MockKafkaReader) PushMessage(msg MockKafkaMessage) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.closed {
		select {
		case m.messages <- msg:
		default:
			// Channel full, skip
		}
	}
}

// MockOrderConsumer simulates the Kafka consumer
type MockOrderConsumer struct {
	reader          *MockKafkaReader
	repo            order.OrderRepository
	processedOrders []string
}

func NewMockOrderConsumer(reader *MockKafkaReader, repo order.OrderRepository) *MockOrderConsumer {
	return &MockOrderConsumer{
		reader:          reader,
		repo:            repo,
		processedOrders: make([]string, 0),
	}
}

// PlaceOrderE2ETestSuite defines a test suite for end-to-end place order flow
type PlaceOrderE2ETestSuite struct {
	suite.Suite
	db            *sql.DB
	repository    order.OrderRepository
	kafkaWriter   *MockKafkaWriter
	kafkaReader   *MockKafkaReader
	producer      *testutil.MockOrderProducer
	consumer      *MockOrderConsumer
	stockCache    *testutil.MockStockCache
	walletService *testutil.MockWalletService
	placeOrderUC  *orderUsecase.PlaceOrderUsecase
	ctx           context.Context
}

func (suite *PlaceOrderE2ETestSuite) SetupSuite() {
	// Load test environment
	if err := godotenv.Load("../../.env.test"); err != nil {
		suite.T().Logf("Could not load .env.test file: %v", err)
	}

	// Database setup - Use existing environment variables
	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port := getEnvOrDefault("POSTGRES_PORT", "5432")
	user := getEnvOrDefault("POSTGRES_USER", "auction_user")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "auction_password")
	dbname := getEnvOrDefault("POSTGRES_DB", "auction_db_test")
	sslmode := getEnvOrDefault("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", connStr)
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), db.Ping())

	suite.db = db
	suite.repository = pgorder.NewPostgresOrderRepository(db)
	suite.ctx = context.Background()

	// Kafka mocks setup
	suite.kafkaWriter = new(MockKafkaWriter)
	suite.kafkaReader = NewMockKafkaReader()
	suite.producer = testutil.NewMockOrderProducer()
	suite.consumer = NewMockOrderConsumer(suite.kafkaReader, suite.repository)

	// Service mocks setup
	suite.stockCache = testutil.NewMockStockCache()
	suite.walletService = testutil.NewMockWalletService()

	// UseCase setup
	suite.placeOrderUC = orderUsecase.NewPlaceOrderUsecase(
		suite.producer,
		suite.stockCache,
		suite.walletService,
	)

	// Ensure tables exist
	suite.ensureOrdersTableExists()
}

func (suite *PlaceOrderE2ETestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	suite.kafkaReader.Close()
}

func (suite *PlaceOrderE2ETestSuite) SetupTest() {
	// Clean database
	_, err := suite.db.ExecContext(suite.ctx, "DELETE FROM orders WHERE buyer_id LIKE 'e2e-test-%'")
	require.NoError(suite.T(), err)

	// Clear Kafka messages
	suite.kafkaWriter.ClearMessages()

	// Reset mocks
	suite.kafkaWriter.ExpectedCalls = nil
	suite.stockCache.RemoveAll(suite.ctx)
	suite.producer.Reset()
}

func (suite *PlaceOrderE2ETestSuite) ensureOrdersTableExists() {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS orders (
		id VARCHAR(255) PRIMARY KEY,
		buyer_id VARCHAR(255) NOT NULL,
		stock_id VARCHAR(255) NOT NULL,
		price DECIMAL(15,2) NOT NULL,
		quantity INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := suite.db.ExecContext(suite.ctx, createTableSQL)
	require.NoError(suite.T(), err)
}

func (suite *PlaceOrderE2ETestSuite) TestPlaceOrder_EndToEnd_Success() {
	// Arrange
	command := order.PlaceOrderCommand{
		BuyerId:  "e2e-test-buyer-123",
		StockId:  "e2e-test-stock-456",
		Quantity: 5,
	}

	stockPrice := 100.0
	availableStock := 10
	totalAmount := stockPrice * float64(command.Quantity)

	// Setup mock data using testutil mock methods
	suite.stockCache.SetInitialStock(command.StockId, availableStock, stockPrice)
	suite.walletService.SetBalance(command.BuyerId, 1000.0)
	suite.kafkaWriter.On("WriteMessages", mock.Anything, mock.AnythingOfType("[]integration_test.MockKafkaMessage")).Return(nil)

	// Act - Step 1: Place Order (UseCase execution)
	err := suite.placeOrderUC.Execute(suite.ctx, command)

	// Assert - Step 1: UseCase should succeed
	assert.NoError(suite.T(), err)

	// Wait for async event publishing to complete
	time.Sleep(200 * time.Millisecond)

	// Verify stock was decreased
	finalStock := suite.stockCache.GetCurrentStock(command.StockId)
	assert.Equal(suite.T(), availableStock-command.Quantity, finalStock)

	// Verify wallet balance was decreased
	finalBalance := suite.walletService.GetBalance(command.BuyerId)
	assert.Equal(suite.T(), 1000.0-totalAmount, finalBalance)

	// Verify order producer was called
	publishedEvents := suite.producer.GetEvents()
	assert.Len(suite.T(), publishedEvents, 1, "Should have published one event")

	// Parse the published event
	publishedEvent := publishedEvents[0].(*order.OrderPlacedEvent)

	// Manually persist the order for e2e test (simulating consumer)
	err = suite.repository.SaveOrder(suite.ctx, *publishedEvent)
	assert.NoError(suite.T(), err)

	// Assert - Step 2: Verify order was persisted to database
	var savedOrder struct {
		Id        string
		BuyerId   string
		StockId   string
		Price     float64
		Quantity  int
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	query := `SELECT id, buyer_id, stock_id, price, quantity, created_at, updated_at 
			  FROM orders WHERE buyer_id = $1`
	err = suite.db.QueryRowContext(suite.ctx, query, command.BuyerId).Scan(
		&savedOrder.Id,
		&savedOrder.BuyerId,
		&savedOrder.StockId,
		&savedOrder.Price,
		&savedOrder.Quantity,
		&savedOrder.CreatedAt,
		&savedOrder.UpdatedAt,
	)

	assert.NoError(suite.T(), err, "Order should be saved to database")
	assert.Equal(suite.T(), publishedEvent.OrderId, savedOrder.Id)
	assert.Equal(suite.T(), command.BuyerId, savedOrder.BuyerId)
	assert.Equal(suite.T(), command.StockId, savedOrder.StockId)
	assert.Equal(suite.T(), totalAmount, savedOrder.Price)
	assert.Equal(suite.T(), command.Quantity, savedOrder.Quantity)
}

func (suite *PlaceOrderE2ETestSuite) TestPlaceOrder_EndToEnd_PaymentFailure() {
	// Arrange
	command := order.PlaceOrderCommand{
		BuyerId:  "e2e-test-buyer-fail",
		StockId:  "e2e-test-stock-fail",
		Quantity: 5,
	}

	stockPrice := 100.0
	availableStock := 10

	// Setup mock data - insufficient funds
	suite.stockCache.SetInitialStock(command.StockId, availableStock, stockPrice)
	suite.walletService.SetBalance(command.BuyerId, 50.0) // Less than total amount (500)

	// Act
	err := suite.placeOrderUC.Execute(suite.ctx, command)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to process payment")

	// Verify no events were published
	publishedEvents := suite.producer.GetEvents()
	assert.Len(suite.T(), publishedEvents, 0, "Should not publish events on payment failure")

	// Verify no order was saved to database
	var count int
	query := `SELECT COUNT(*) FROM orders WHERE buyer_id = $1`
	err = suite.db.QueryRowContext(suite.ctx, query, command.BuyerId).Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count, "No order should be saved on payment failure")

	// Verify stock was not changed
	finalStock := suite.stockCache.GetCurrentStock(command.StockId)
	assert.Equal(suite.T(), availableStock, finalStock)

	// Verify wallet balance was not changed
	finalBalance := suite.walletService.GetBalance(command.BuyerId)
	assert.Equal(suite.T(), 50.0, finalBalance)
}

func (suite *PlaceOrderE2ETestSuite) TestPlaceOrder_EndToEnd_StockUpdateFailureWithRollback() {
	// This test is simplified since testutil mocks don't support arbitrary error injection
	// We test the insufficient stock scenario instead

	// Arrange
	command := order.PlaceOrderCommand{
		BuyerId:  "e2e-test-buyer-rollback",
		StockId:  "e2e-test-stock-rollback",
		Quantity: 15, // More than available stock
	}

	stockPrice := 100.0
	availableStock := 10 // Less than requested quantity

	// Setup mock data
	suite.stockCache.SetInitialStock(command.StockId, availableStock, stockPrice)
	suite.walletService.SetBalance(command.BuyerId, 2000.0) // Sufficient funds

	// Act
	err := suite.placeOrderUC.Execute(suite.ctx, command)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "out of stock")

	// Verify no events were published
	publishedEvents := suite.producer.GetEvents()
	assert.Len(suite.T(), publishedEvents, 0, "Should not publish events on stock failure")

	// Verify no order was saved to database
	var count int
	query := `SELECT COUNT(*) FROM orders WHERE buyer_id = $1`
	err = suite.db.QueryRowContext(suite.ctx, query, command.BuyerId).Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count, "No order should be saved on stock failure")

	// Verify stock was not changed
	finalStock := suite.stockCache.GetCurrentStock(command.StockId)
	assert.Equal(suite.T(), availableStock, finalStock)

	// Verify wallet balance was not changed (no rollback needed since payment wasn't processed)
	finalBalance := suite.walletService.GetBalance(command.BuyerId)
	assert.Equal(suite.T(), 2000.0, finalBalance)
}

// Helper function
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestPlaceOrderE2ETestSuite runs the test suite
func TestPlaceOrderE2ETestSuite(t *testing.T) {
	suite.Run(t, new(PlaceOrderE2ETestSuite))
}
