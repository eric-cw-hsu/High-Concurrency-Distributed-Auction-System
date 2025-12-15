package order

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil"
)

// OrderRepositoryTestSuite defines a test suite for order repository integration tests
type OrderRepositoryTestSuite struct {
	suite.Suite
	dbHelper      *testutil.DatabaseTestHelper
	userHelper    *testutil.UserTestHelper
	productHelper *testutil.ProductTestHelper
	stockHelper   *testutil.StockTestHelper
	repository    order.OrderRepository
	ctx           context.Context

	// Test data
	testUsers    []string
	testProducts []string
	testStocks   []string
}

// SetupSuite runs once before all tests in the suite
func (suite *OrderRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Create database test helper
	dbHelper, err := testutil.NewDatabaseTestHelper(suite.ctx)
	require.NoError(suite.T(), err, "Failed to create database test helper")

	// Run migrations
	err = dbHelper.RunMigrations()
	require.NoError(suite.T(), err, "Failed to run migrations")

	suite.dbHelper = dbHelper
	suite.repository = NewPostgresOrderRepository(dbHelper.DB)

	// Initialize test helpers
	suite.userHelper = testutil.NewUserTestHelper(suite.ctx, dbHelper.DB, nil, nil)
	suite.productHelper = testutil.NewProductTestHelper(dbHelper.DB)
	suite.stockHelper = testutil.NewStockTestHelper(dbHelper.DB)
}

// TearDownSuite runs once after all tests in the suite
func (suite *OrderRepositoryTestSuite) TearDownSuite() {
	if suite.dbHelper != nil {
		suite.dbHelper.Close()
	}
}

// SetupTest runs before each test
func (suite *OrderRepositoryTestSuite) SetupTest() {
	// Clean up any existing test data using database helper
	err := suite.dbHelper.CleanDatabase()
	require.NoError(suite.T(), err)

	// Create test users (required for foreign key constraints)
	suite.testUsers, err = suite.userHelper.CreateTestUsersOnly(5)
	require.NoError(suite.T(), err, "Failed to create test users")

	// Create test products
	suite.testProducts, err = suite.productHelper.CreateTestProducts(suite.ctx, 3)
	require.NoError(suite.T(), err, "Failed to create test products")

	// Create test stocks (linked to products)
	suite.testStocks, err = suite.stockHelper.CreateTestStocks(suite.ctx, suite.testProducts, suite.testUsers, 100)
	require.NoError(suite.T(), err, "Failed to create test stocks")
}

// TearDownTest runs after each test
func (suite *OrderRepositoryTestSuite) TearDownTest() {
	// Clean up test data using database helper
	err := suite.dbHelper.CleanDatabase()
	require.NoError(suite.T(), err)
}

func (suite *OrderRepositoryTestSuite) TestSaveOrder_Success() {
	// Arrange - use valid foreign keys
	orderEntity := &order.Order{
		OrderID:    "550e8400-e29b-41d4-a716-446655440000",
		BuyerID:    suite.testUsers[0],
		StockID:    suite.testStocks[0],
		Quantity:   5,
		TotalPrice: 500.00,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	// Act
	err := suite.repository.SaveOrder(suite.ctx, orderEntity)

	// Assert
	assert.NoError(suite.T(), err)

	// Verify the order was saved correctly
	var savedOrder struct {
		OrderID   string
		BuyerID   string
		StockID   string
		Price     float64
		Quantity  int
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	query := `SELECT order_id, buyer_id, stock_id, total_price, quantity, created_at, updated_at 
			  FROM orders WHERE order_id = $1`
	err = suite.dbHelper.DB.QueryRowContext(suite.ctx, query, orderEntity.OrderID).Scan(
		&savedOrder.OrderID,
		&savedOrder.BuyerID,
		&savedOrder.StockID,
		&savedOrder.Price,
		&savedOrder.Quantity,
		&savedOrder.CreatedAt,
		&savedOrder.UpdatedAt,
	)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), orderEntity.OrderID, savedOrder.OrderID)
	assert.Equal(suite.T(), orderEntity.BuyerID, savedOrder.BuyerID)
	assert.Equal(suite.T(), orderEntity.StockID, savedOrder.StockID)
	assert.Equal(suite.T(), orderEntity.TotalPrice, savedOrder.Price)
	assert.Equal(suite.T(), orderEntity.Quantity, savedOrder.Quantity)
}

func (suite *OrderRepositoryTestSuite) TestSaveOrder_DuplicateId() {
	// Arrange - use valid foreign keys
	order1 := &order.Order{
		OrderID:    "550e8400-e29b-41d4-a716-446655440001",
		BuyerID:    suite.testUsers[0],
		StockID:    suite.testStocks[0],
		Quantity:   3,
		TotalPrice: 300.00,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	order2 := &order.Order{
		OrderID:    "550e8400-e29b-41d4-a716-446655440001",
		BuyerID:    suite.testUsers[1],
		StockID:    suite.testStocks[1],
		Quantity:   5,
		TotalPrice: 500.00,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	// Act
	err1 := suite.repository.SaveOrder(suite.ctx, order1)
	err2 := suite.repository.SaveOrder(suite.ctx, order2)

	// Assert
	assert.NoError(suite.T(), err1)
	assert.Error(suite.T(), err2) // Should fail due to duplicate primary key
}

func (suite *OrderRepositoryTestSuite) TestSaveOrder_LargeValues() {
	// Arrange - use valid foreign keys
	orderEntity := &order.Order{
		OrderID:    "550e8400-e29b-41d4-a716-446655440002",
		BuyerID:    suite.testUsers[1],
		StockID:    suite.testStocks[1],
		Quantity:   1000,
		TotalPrice: 999999.99,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	// Act
	err := suite.repository.SaveOrder(suite.ctx, orderEntity)

	// Assert
	assert.NoError(suite.T(), err)

	// Verify large values are stored correctly
	var savedPrice float64
	var savedQuantity int
	query := `SELECT total_price, quantity FROM orders WHERE order_id = $1`
	err = suite.dbHelper.DB.QueryRowContext(suite.ctx, query, orderEntity.OrderID).Scan(&savedPrice, &savedQuantity)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), orderEntity.TotalPrice, savedPrice)
	assert.Equal(suite.T(), orderEntity.Quantity, savedQuantity)
}

func (suite *OrderRepositoryTestSuite) TestSaveOrder_MultipleOrders() {
	// Arrange - use valid foreign keys from test data
	buyerID := suite.testUsers[2]
	orders := []*order.Order{
		{
			OrderID:    "550e8400-e29b-41d4-a716-446655440003",
			BuyerID:    buyerID,
			StockID:    suite.testStocks[0],
			Quantity:   2,
			TotalPrice: 200.00,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			OrderID:    "550e8400-e29b-41d4-a716-446655440004",
			BuyerID:    buyerID,
			StockID:    suite.testStocks[1],
			Quantity:   3,
			TotalPrice: 300.00,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			OrderID:    "550e8400-e29b-41d4-a716-446655440005",
			BuyerID:    buyerID,
			StockID:    suite.testStocks[2],
			Quantity:   1,
			TotalPrice: 100.00,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}
	// Act
	for _, orderEntity := range orders {
		err := suite.repository.SaveOrder(suite.ctx, orderEntity)
		assert.NoError(suite.T(), err)
	}

	// Assert
	// Count orders for this buyer
	var count int
	query := `SELECT COUNT(*) FROM orders WHERE buyer_id = $1`
	err := suite.dbHelper.DB.QueryRowContext(suite.ctx, query, buyerID).Scan(&count)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)
}

func (suite *OrderRepositoryTestSuite) TestSaveOrder_ForeignKeyConstraint() {
	// Arrange - test with invalid foreign keys
	orderEntity := &order.Order{
		OrderID:    "550e8400-e29b-41d4-a716-446655440006",
		BuyerID:    "550e8400-e29b-41d4-a716-446655440999",
		StockID:    "550e8400-e29b-41d4-a716-446655440998",
		Quantity:   1,
		TotalPrice: 100.00,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	// Act
	err := suite.repository.SaveOrder(suite.ctx, orderEntity)

	// Assert
	// Should fail due to foreign key constraint violations
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "foreign key")
}

func (suite *OrderRepositoryTestSuite) TestSaveOrder_DecimalPrecision() {
	// Arrange - test decimal precision with valid foreign keys
	orderEntity := &order.Order{
		OrderID:    "550e8400-e29b-41d4-a716-446655440007",
		BuyerID:    suite.testUsers[3],
		StockID:    suite.testStocks[0],
		Quantity:   1,
		TotalPrice: 123.456789,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	// Act
	err := suite.repository.SaveOrder(suite.ctx, orderEntity)

	// Assert
	assert.NoError(suite.T(), err)

	// Verify decimal precision (should be rounded to 2 decimal places)
	var savedPrice float64
	query := `SELECT total_price FROM orders WHERE order_id = $1`
	err = suite.dbHelper.DB.QueryRowContext(suite.ctx, query, orderEntity.OrderID).Scan(&savedPrice)

	assert.NoError(suite.T(), err)
	// PostgreSQL DECIMAL(15,2) should round to 2 decimal places
	assert.InDelta(suite.T(), 123.46, savedPrice, 0.01)
}

func (suite *OrderRepositoryTestSuite) TestRepository_ContextCancellation() {
	// Arrange - use valid foreign keys
	orderEntity := &order.Order{
		OrderID:    "550e8400-e29b-41d4-a716-446655440008",
		BuyerID:    suite.testUsers[4],
		StockID:    suite.testStocks[2],
		Quantity:   1,
		TotalPrice: 100.00,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	// Act
	err := suite.repository.SaveOrder(ctx, orderEntity)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "context canceled")
}

// TestOrderRepositoryTestSuite runs the test suite
func TestOrderRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(OrderRepositoryTestSuite))
}
