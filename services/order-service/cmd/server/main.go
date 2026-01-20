package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/application/worker"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/infrastructure/messaging/kafka"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/infrastructure/persistence/postgres"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/infrastructure/persistence/redis"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/interface/grpc"
	grpcserver "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/interface/grpc"
	productv1pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/product/v1"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// 1. Load environment and configuration
	if os.Getenv("ENV") != "production" {
		_ = godotenv.Load()
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize core logging
	log := logger.Init(&cfg.Logger)
	defer log.Sync()

	zap.L().Info("starting order service",
		zap.String("env", cfg.Logger.Environment),
	)

	// 3. Initialize Infrastructure
	db := postgres.MustConnect(cfg.Database)
	defer db.Close()

	redisClient := redis.MustConnect(cfg.Redis)
	defer redisClient.Close()

	productClientConn := grpc.MustConnProductClient(cfg.GRPC)
	defer productClientConn.Close()

	pbProductClient := productv1pb.NewProductServiceClient(productClientConn)
	productGRPCClient := grpc.NewProductClient(pbProductClient)

	// 4. Initialize Repositories & Managers
	txManager := postgres.NewTxManager(db)
	orderRepo := postgres.NewOrderRepository(db)
	outboxRepo := postgres.NewOutboxRepository(db)
	productPriceRepo := postgres.NewProductPriceRepository(db)
	timeoutQueue := redis.NewTimeoutQueue(redisClient)

	// 5. Initialize Application Services
	orderAppService := service.NewOrderAppService(txManager, timeoutQueue, orderRepo, productPriceRepo, productGRPCClient)
	productAppService := service.NewProductAppService(productPriceRepo)

	// 6. Initialize Workers & Messaging
	// Kafka Producer for Domain Events
	producer := kafka.NewProducer(&cfg.Kafka)
	defer producer.Close()

	// Outbox Relay Worker (Relays DB events to Kafka)
	outboxRelayWorker := worker.NewOutboxRelayWorker(
		outboxRepo,
		producer,
		cfg.Outbox.Interval,
		cfg.Outbox.BatchSize,
	)

	// Order Timeout Worker (Scans Redis for expired orders)
	timeoutWorker := worker.NewOrderTimeoutWorker(
		orderAppService,
		timeoutQueue,
		&cfg.OrderTimeoutWorker,
	)

	// Kafka Consumer (Listens to Stock Service reservations)
	reservationHandler := kafka.NewReservationEventHandler(orderAppService)
	kafkaConsumer := kafka.NewConsumer(&cfg.Kafka, reservationHandler)
	defer kafkaConsumer.Close()

	productEventHandler := kafka.NewProductEventHandler(productAppService)
	productKafkaConfig := &config.KafkaConfig{
		Brokers:         cfg.Kafka.Brokers,
		ConsumerTopic:   "product-events", // The topic where Product Service sends updates
		ConsumerGroupID: "order-service-product-sync",
	}
	productConsumer := kafka.NewConsumer(productKafkaConfig, productEventHandler)
	defer productConsumer.Close()

	// 7. Initialize gRPC Server
	grpcHandler := grpcserver.NewOrderHandler(orderAppService)
	grpcServer := grpcserver.NewServer(&cfg.GRPC, grpcHandler)

	// 8. Lifecycle Management
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 9. Start Concurrent Components
	// Start Kafka Consumer
	go func() {
		if err := kafkaConsumer.Start(ctx); err != nil && ctx.Err() == nil {
			zap.L().Error("kafka consumer failed", zap.Error(err))
		}
	}()

	go func() {
		if err := productConsumer.Start(ctx); err != nil {
			zap.L().Error("product consumer failed", zap.Error(err))
		}
	}()

	// Start Outbox Relay
	go func() {
		if err := outboxRelayWorker.Start(ctx); err != nil && ctx.Err() == nil {
			zap.L().Error("outbox relay worker failed", zap.Error(err))
		}
	}()

	// Start Timeout Scanning
	go func() {
		if err := timeoutWorker.Start(ctx); err != nil && ctx.Err() == nil {
			zap.L().Error("timeout worker failed", zap.Error(err))
		}
	}()

	// Start gRPC Server
	go func() {
		zap.L().Info("grpc server listening", zap.Int("port", cfg.GRPC.Server.Port))
		if err := grpcServer.Start(); err != nil {
			zap.L().Error("grpc server failed", zap.Error(err))
			os.Exit(1)
		}
	}()

	// 10. Graceful Shutdown Signal Handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("shutting down order service gracefully...")

	// Stop all workers via context cancellation
	cancel()

	// Stop gRPC server
	grpcServer.GracefulStop()

	zap.L().Info("order service stopped cleanly")
}
