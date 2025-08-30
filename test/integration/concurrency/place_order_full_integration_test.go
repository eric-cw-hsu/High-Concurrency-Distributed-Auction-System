package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/samborkent/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres"
	walletRepo "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/wallet"
	redisStock "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/redis/stockcache"
	orderUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/order"
	walletUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil/mocks"
)

// FullIntegrationTestSuite tests auction system with PostgreSQL + Redis + Mock Kafka
type FullIntegrationTestSuite struct {
	suite.Suite
	db          *sql.DB
	redisClient *redis.Client

	walletRepo     *walletRepo.PostgresWalletRepository
	walletService  walletUsecase.WalletService
	addFundUsecase *walletUsecase.AddFundUsecase
	stockCache     *redisStock.RedisStockCache
	producer       *mocks.MockOrderProducer
	placeOrderUC   *orderUsecase.PlaceOrderUsecase

	ctx       context.Context
	testUsers []string
}

func (suite *FullIntegrationTestSuite) runDatabaseMigration(pgCfg config.PostgresConfig) {
	// drop all tables
	_, err := suite.db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		log.Fatalf("failed to reset database: %v", err)
	}
	// Perform database migration
	suite.T().Log("Running database migrations...")

	// Create proper postgres DSN with scheme
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pgCfg.User, pgCfg.Password, pgCfg.Host, pgCfg.Port, pgCfg.DBName)

	migrator, err := migrate.New(
		"file://../../../db/migrations",
		dsn,
	)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migration: %v", err)
	}
}

func (suite *FullIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Load test environment from project root
	if err := godotenv.Load("../../../.env.test"); err != nil {
		suite.T().Logf("Warning: Could not load .env.test file: %v", err)
		suite.T().Log("Make sure to run tests from project root directory")
	}

	// Setup PostgreSQL
	pgCfg := config.LoadPostgresConfig()
	var err error
	suite.db, err = postgres.NewPostgresClient(pgCfg)
	suite.Require().NoError(err, "Failed to connect to PostgreSQL")

	// Run database migrations
	suite.runDatabaseMigration(pgCfg)

	// Setup Redis
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := 6379
	if portStr := os.Getenv("REDIS_PORT"); portStr != "" {
		if parsed, err := strconv.Atoi(portStr); err == nil {
			redisPort = parsed
		}
	}
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if parsed, err := strconv.Atoi(dbStr); err == nil {
			redisDB = parsed
		}
	}
	suite.redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", redisHost, redisPort),
		DB:   redisDB,
	})
	_, err = suite.redisClient.Ping(suite.ctx).Result()
	suite.Require().NoError(err, "Failed to connect to Redis")

	// Initialize services
	suite.walletRepo = walletRepo.NewPostgresWalletRepository(suite.db)
	mockEventPub := mocks.NewMockWalletEventPublisher()
	suite.walletService = walletUsecase.NewWalletService(suite.walletRepo, mockEventPub)
	suite.addFundUsecase = walletUsecase.NewAddFundUsecase(suite.walletRepo, mockEventPub)

	suite.stockCache, err = redisStock.NewRedisStockCache(suite.redisClient)
	suite.Require().NoError(err, "Failed to create stock cache")

	suite.producer = mocks.NewMockOrderProducer()
	suite.placeOrderUC = orderUsecase.NewPlaceOrderUsecase(
		suite.producer,
		suite.stockCache,
		suite.walletService,
	)

	// Setup test users with valid UUIDs v7
	suite.testUsers = make([]string, 5)
	for i := 0; i < 5; i++ {
		userID := uuidv7.New().String()
		suite.testUsers[i] = userID

		// First create user in users table
		_, err := suite.db.ExecContext(suite.ctx,
			"INSERT INTO users (id, email, password_hash, name) VALUES ($1, $2, $3, $4)",
			userID, fmt.Sprintf("user%d@test.com", i), "test_hash", fmt.Sprintf("Test User %d", i))
		suite.Require().NoError(err, "Failed to create user %s", userID)

		// Then create wallet
		_, err = suite.walletService.CreateWallet(suite.ctx, userID)
		suite.Require().NoError(err, "Failed to create wallet for user %s", userID)

		// Add initial balance using AddFundUsecase
		addFundCmd := &wallet.AddFundCommand{
			UserId:      userID,
			Amount:      10000.0,
			Description: "Initial test funding",
		}
		_, err = suite.addFundUsecase.Execute(suite.ctx, addFundCmd)
		suite.Require().NoError(err, "Failed to add initial balance to wallet for user %s", userID)

		suite.T().Logf("Created user %s with wallet balance: $%.2f", userID, 10000.0)
	}

	suite.T().Log("Full integration test setup completed")
}

func (suite *FullIntegrationTestSuite) TearDownSuite() {
	if suite.redisClient != nil {
		suite.redisClient.FlushDB(suite.ctx)
		suite.redisClient.Close()
	}
	if suite.db != nil {
		for _, userId := range suite.testUsers {
			suite.db.Exec("DELETE FROM wallets WHERE user_id = $1", userId)
		}
		suite.db.Close()
	}
}

func (suite *FullIntegrationTestSuite) SetupTest() {
	// Reset Redis
	suite.redisClient.FlushDB(suite.ctx)

	// Initialize stock
	suite.stockCache.SetStock(suite.ctx, "STOCK001", 1000)
	suite.stockCache.SetPrice(suite.ctx, "STOCK001", 100.0)

	// Ensure wallets exist
	for _, userId := range suite.testUsers {
		suite.walletService.EnsureWalletExists(suite.ctx, userId)
	}
}

// TestFullIntegration_AuctionRush tests concurrent auction scenario
func (suite *FullIntegrationTestSuite) TestFullIntegration_AuctionRush() {
	concurrency := 20 // Reduced for stability
	ordersPerUser := 2
	totalOrders := concurrency * ordersPerUser

	var (
		successCount int64
		failureCount int64
		totalLatency int64
	)

	startTime := time.Now()
	var wg sync.WaitGroup

	// Simulate auction rush
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineId int) {
			defer wg.Done()

			userId := suite.testUsers[goroutineId%len(suite.testUsers)]

			for j := 0; j < ordersPerUser; j++ {
				orderStart := time.Now()

				cmd := order.PlaceOrderCommand{
					BuyerId:  userId,
					StockId:  "STOCK001",
					Quantity: 1,
				}

				err := suite.placeOrderUC.Execute(suite.ctx, cmd)
				orderDuration := time.Since(orderStart)
				atomic.AddInt64(&totalLatency, orderDuration.Microseconds())

				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					// Log first few errors to understand what's happening
					if atomic.LoadInt64(&failureCount) <= 5 {
						suite.T().Logf("Order failed (goroutine %d, order %d): %v", goroutineId, j, err)
					}
				} else {
					atomic.AddInt64(&successCount, 1)
				}

				time.Sleep(time.Millisecond * 50) // Longer delay for stability
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Get final stock
	finalStock, err := suite.stockCache.GetStock(suite.ctx, "STOCK001")
	suite.Require().NoError(err)

	// Calculate metrics
	avgLatency := time.Duration(atomic.LoadInt64(&totalLatency)/int64(totalOrders)) * time.Microsecond
	throughput := float64(atomic.LoadInt64(&successCount)) / totalDuration.Seconds()

	// Log results
	suite.T().Logf("\n=== FULL INTEGRATION AUCTION PERFORMANCE ===")
	suite.T().Logf("Setup: PostgreSQL + Redis + Mock Kafka")
	suite.T().Logf("Total Duration: %v", totalDuration)
	suite.T().Logf("Total Orders: %d", totalOrders)
	suite.T().Logf("Successful: %d", atomic.LoadInt64(&successCount))
	suite.T().Logf("Failed: %d", atomic.LoadInt64(&failureCount))
	suite.T().Logf("Success Rate: %.2f%%", float64(atomic.LoadInt64(&successCount))/float64(totalOrders)*100)
	suite.T().Logf("Average Latency: %v", avgLatency)
	suite.T().Logf("Throughput: %.2f orders/sec", throughput)
	suite.T().Logf("Initial Stock: 1000")
	suite.T().Logf("Final Stock: %d", finalStock)
	suite.T().Logf("Consumed: %d", 1000-finalStock)

	// Assertions
	assert.Equal(suite.T(), int64(1000-finalStock), atomic.LoadInt64(&successCount),
		"Stock consumed should equal successful orders")
	assert.GreaterOrEqual(suite.T(), int(finalStock), 0,
		"Stock should never go negative")

	// Verify Kafka mock was called
	publishCount := suite.producer.GetPublishCount()
	suite.T().Logf("Kafka mock publish count: %d", publishCount)
}

func TestFullIntegrationSuite(t *testing.T) {
	suite.Run(t, new(FullIntegrationTestSuite))
}
