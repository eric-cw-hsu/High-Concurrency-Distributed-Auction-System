package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/infrastructure/messaging/kafka"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/infrastructure/outbox"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/infrastructure/persistence/postgres"
	grpcserver "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/interface/grpc"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	// Load .env file (if not in production)
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Warning: .env file not found\n")
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.Init(&cfg.Logger)
	defer log.Sync()

	log.Info("starting product service",
		zap.String("environment", cfg.Logger.Environment),
	)

	// Initialize database
	db, err := initDatabase(&cfg.Database)
	if err != nil {
		log.Error("failed to initialize database", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repositories
	productRepo := postgres.NewProductRepository(db)
	productWriter := postgres.NewProductWriter(db)
	outboxRepo := postgres.NewOutboxRepository(db)

	// Initialize application services
	productService := service.NewProductService(productRepo, productWriter)

	// Initialize Kafka producer
	producer := kafka.NewProducer(&cfg.Kafka)
	defer producer.Close()

	// Initialize outbox relay worker
	outboxRelay := outbox.NewOutboxRelay(outboxRepo, producer, &cfg.Outbox)

	// Initialize Kafka consumer
	stockEventHandler := kafka.NewStockEventHandler(productService)
	consumer := kafka.NewConsumer(&cfg.Kafka, stockEventHandler)
	defer consumer.Close()

	// Initialize gRPC server
	grpcServer := grpcserver.NewServer(&cfg.Server, productService)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background workers
	go func() {
		zap.L().Info("starting outbox relay worker")
		if err := outboxRelay.Start(ctx); err != nil && ctx.Err() == nil {
			zap.L().Error("outbox relay error", zap.Error(err))
		}
	}()

	go func() {
		zap.L().Info("starting outbox cleanup worker")
		if err := outboxRelay.StartCleanup(ctx); err != nil && ctx.Err() == nil {
			zap.L().Error("outbox cleanup error", zap.Error(err))
		}
	}()

	go func() {
		zap.L().Info("starting kafka consumer")
		if err := consumer.Start(ctx); err != nil && ctx.Err() == nil {
			zap.L().Error("kafka consumer error", zap.Error(err))
		}
	}()

	// Start gRPC server
	go func() {
		zap.L().Info("starting grpc server",
			zap.Int("port", cfg.Server.GRPCPort),
		)
		if err := grpcServer.Start(); err != nil {
			zap.L().Error("grpc server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("shutting down gracefully")

	// Cancel context to stop workers
	cancel()

	// Stop gRPC server
	grpcServer.Stop()

	zap.L().Info("server stopped")
}

// initDatabase initializes database connection
func initDatabase(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	zap.L().Info("connecting to database",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
	)

	db, err := sqlx.Connect("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	zap.L().Info("database connected successfully")

	return db, nil
}
