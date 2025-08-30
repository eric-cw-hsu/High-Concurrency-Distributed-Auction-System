package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/bootstrap"
	loggerapp "eric-cw-hsu.github.io/scalable-auction-system/internal/app/logger"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load service-specific .env file
	if err := godotenv.Load(".env.logger"); err != nil {
		if err := godotenv.Load(); err != nil {
			// Use standard log before logger initialization
			log.Println("No .env file found, relying on system env vars")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configurations
	kafkaConfig := config.LoadKafkaConfig()
	cfg := config.LoadConfig()

	// Initialize global logger
	consoleSender := logger.NewConsoleSender()
	appLogger := logger.NewLogger("logger", consoleSender)
	logger.SetDefault(appLogger)

	logger.Info("Starting Logger service", map[string]interface{}{
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

	logger.Info("Starting Logger Service", map[string]interface{}{
		"kafka_broker":  cfg.KafkaBroker,
		"storage_type":  cfg.LogStorageType,
		"log_file_path": cfg.LogFilePath,
		"topics":        cfg.Topics,
	})

	// Create logger application
	app, err := loggerapp.NewLoggerApp(cfg, nil) // Pass nil since we're using global logger now
	if err != nil {
		logger.Fatal("Failed to initialize logger app", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Setup graceful shutdown
	go handleGracefulShutdown(cancel)

	logger.Info("Logger Service started successfully")

	// Start the application
	if err := app.Start(ctx); err != nil {
		logger.Fatal("Logger service failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Cleanup
	if err := app.Stop(); err != nil {
		logger.Error("Error stopping logger service", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("Logger Service stopped gracefully")
}

func handleGracefulShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal, stopping logger service")
	cancel()
}
