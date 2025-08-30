package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres"
	pgorder "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/order"
	kafkaconsumer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/consumer"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load service-specific .env file
	if err := godotenv.Load(".env.order-consumer"); err != nil {
		if err := godotenv.Load(); err != nil {
			// Use standard log before logger initialization
			log.Println("No .env file found, relying on system env vars")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configurations
	kafkaConfig := config.LoadKafkaConfig()
	pgConfig := config.LoadPostgresConfig()

	// Initialize global logger
	kafkaSender := logger.NewKafkaSender(kafkaConfig.Brokers, "service.logs")
	appLogger := logger.NewLogger("order-consumer", kafkaSender)
	logger.SetDefault(appLogger)

	logger.Info("Starting Order Consumer service", map[string]interface{}{
		"kafka_brokers": kafkaConfig.Brokers,
	})

	// Initialize Kafka manager to ensure topics exist
	kafkaManager, err := bootstrap.NewKafkaManager(kafkaConfig)
	if err != nil {
		logger.Fatal("Failed to create Kafka manager", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Ensure all required topics exist before starting the service
	logger.Info("Ensuring Kafka topics exist")
	if err := kafkaManager.EnsureAllTopics(ctx); err != nil {
		logger.Fatal("Failed to ensure Kafka topics", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Initialize PostgreSQL connection
	logger.Info("Initializing PostgreSQL connection")
	postgresDB, err := postgres.NewPostgresClient(pgConfig)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer func() {
		postgresDB.Close()
	}()
	logger.Info("PostgreSQL connection established")

	// Initialize order service
	logger.Info("Initializing Kafka consumer")
	orderRepo := pgorder.NewPostgresOrderRepository(postgresDB)
	// Use config values for consumer
	reader := kafkaInfra.NewReader(kafkaConfig.Brokers, "order.placed", "order-consumer-group")
	orderConsumer := kafkaconsumer.NewOrderConsumer(
		reader,
		orderRepo,
	)
	logger.Info("Kafka consumer initialized")

	// Setup graceful shutdown
	logger.Info("Starting order consumer")
	go func() {
		if err := orderConsumer.Start(ctx); err != nil {
			logger.Warn("Consumer stopped", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Order Consumer service is running")

	<-sigChan
	logger.Info("Received shutdown signal, stopping consumer")
	cancel()

	// Cleanup
	orderConsumer.Stop()

	logger.Info("Order Consumer service stopped gracefully")
}
