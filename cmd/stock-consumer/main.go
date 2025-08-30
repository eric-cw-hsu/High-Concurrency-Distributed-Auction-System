package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres"
	pgstock "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/stock"
	kafkaconsumer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/consumer"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load service-specific .env file
	if err := godotenv.Load(".env.stock-consumer"); err != nil {
		if err := godotenv.Load(); err != nil {
			// Will log this after logger is initialized
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configurations
	kafkaConfig := config.LoadKafkaConfig()
	pgConfig := config.LoadPostgresConfig()

	// Initialize global logger
	kafkaSender := logger.NewKafkaSender(kafkaConfig.Brokers, "service.logs")
	appLogger := logger.NewLogger("stock-consumer", kafkaSender)
	logger.SetDefault(appLogger)

	logger.Info("Starting Stock Consumer service", map[string]interface{}{
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
	pg, err := postgres.NewPostgresClient(pgConfig)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer pg.Close()

	// Initialize stock repository
	stockRepo := pgstock.NewPostgresStockRepository(pg)
	logger.Info("PostgreSQL connection established")

	// Initialize Kafka reader for stock events
	logger.Info("Initializing Kafka consumer")
	kafkaReader := kafkaInfra.NewReader(kafkaConfig.Brokers, "order.placed", "stock-service")
	defer kafkaReader.Close()

	// Create stock consumer
	stockConsumer := kafkaconsumer.NewStockConsumer(kafkaReader, stockRepo)
	logger.Info("Kafka consumer initialized")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal, stopping stock consumer")
		cancel()
	}()

	logger.Info("Starting Stock Consumer Service")
	logger.Info("Listening for order.placed events to update stock")

	// Start consumer with recovery logic
	if err := stockConsumer.StartWithRecovery(ctx); err != nil {
		logger.Fatal("Stock consumer failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Cleanup
	if err := stockConsumer.Stop(); err != nil {
		logger.Error("Error stopping stock consumer", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("Stock Consumer Service stopped gracefully")
}
