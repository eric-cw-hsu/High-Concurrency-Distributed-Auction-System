package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/messaging/kafka"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/outbox"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/persistence/postgres"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/persistence/redis"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/recovery"
	grpcserver "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/interface/grpc"

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

	log.Info("starting stock service",
		zap.String("environment", cfg.Logger.Environment),
	)

	// Initialize database
	db := postgres.MustConnect(cfg.Database)
	defer db.Close()

	// Initialize Redis
	redisClient := redis.MustConnect(cfg.Redis)
	defer redisClient.Close()

	// Initialize repositories
	// Redis
	stockRepo := redis.NewStockRepository(redisClient)
	reservationRedisRepo := redis.NewReservationRepository(redisClient)
	stockReservationCoordinator := redis.NewStockReservationCoordinator(redisClient)
	productStateRepo := redis.NewProductStateRepository(redisClient)
	// Postgres
	reservationPostgresRepo := postgres.NewReservationRepository(db)

	outboxRepo := postgres.NewOutboxRepository(db)

	// recovery redis
	redisRecovery := recovery.NewRedisRecovery(redisClient, reservationPostgresRepo, reservationRedisRepo)
	ctx := context.Background()
	needsRecovery, err := redisRecovery.CheckRedisHealth(ctx)
	if err != nil {
		zap.L().Error("failed to check redis health", zap.Error(err))
	} else if needsRecovery {
		zap.L().Warn("redis appears empty, attempting recovery")

		if err := redisRecovery.FullRecovery(ctx); err != nil {
			zap.L().Error("redis recovery failed", zap.Error(err))
			// Continue startup (allow manual recovery later)
		}
	}

	productStateRecovery := recovery.NewProductStateRecovery(redisClient, productStateRepo, cfg.Kafka.Brokers, cfg.Kafka.ProductEventsTopic)
	if err := productStateRecovery.CheckAndRecover(ctx); err != nil {
		zap.L().Error("product state recovery failed", zap.Error(err))
	}

	// Initialize application services
	reservationPersistQueue := service.NewReservationPersistQueue(&cfg.Service)
	stockService := service.NewStockService(&cfg.Service, stockRepo, reservationRedisRepo, reservationPostgresRepo, stockReservationCoordinator, outboxRepo, reservationPersistQueue, productStateRepo)
	reservationPersistWorker := service.NewReservationPersistWorker(&cfg.Service, reservationPostgresRepo, reservationPersistQueue)

	// Initialize Kafka producer
	producer := kafka.NewProducer(&cfg.Kafka)
	defer producer.Close()

	// Initialize outbox relay worker
	outboxRelay := outbox.NewOutboxRelay(outboxRepo, producer, &cfg.Outbox)

	// Initialize Kafka consumer
	orderEventHandler := kafka.NewOrderEventHandler(stockService)
	orderConsumer := kafka.NewConsumer(&cfg.Kafka, orderEventHandler)
	defer orderConsumer.Close()
	productKafkaConfig := &config.KafkaConfig{
		Brokers:         cfg.Kafka.Brokers,
		ConsumerTopic:   cfg.Kafka.ProductEventsTopic, // "product-events"
		ConsumerGroupID: "stock-service-product-consumer",
	}
	productEventHandler := kafka.NewProductEventHandler(productStateRepo)
	productConsumer := kafka.NewConsumer(productKafkaConfig, productEventHandler)
	defer productConsumer.Close()

	// Initialize gRPC server
	grpcServer := grpcserver.NewServer(&cfg.Server, stockService, redisRecovery)

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
		zap.L().Info("starting kafka order consumer")
		if err := orderConsumer.Start(ctx); err != nil && ctx.Err() == nil {
			zap.L().Error("kafka order consumer error", zap.Error(err))
		}
	}()

	go func() {
		zap.L().Info("starting kafka product consumer")
		if err := productConsumer.Start(ctx); err != nil && ctx.Err() == nil {
			zap.L().Error("kafka product consumer error", zap.Error(err))
		}
	}()

	go func() {
		zap.L().Info("starting reservation persistent worker")
		reservationPersistWorker.Start(ctx)
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
