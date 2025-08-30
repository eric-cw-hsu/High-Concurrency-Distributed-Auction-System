package api

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/api/bootstrap"
	globalBootstrap "eric-cw-hsu.github.io/scalable-auction-system/internal/app/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

func Run() error {
	ctx := context.Background()

	// Load Kafka configuration
	kafkaConfig := config.LoadKafkaConfig()

	// Create and set the global logger
	kafkaSender := logger.NewKafkaSender(kafkaConfig.Brokers, "service.logs")
	appLogger := logger.NewLogger("api-service", kafkaSender)
	logger.SetDefault(appLogger)
	defer appLogger.Close()

	// Use global logger functions
	logger.Info("Starting API service", map[string]interface{}{
		"kafka_brokers": kafkaConfig.Brokers,
	})

	// Initialize Kafka manager to ensure topics exist
	kafkaManager, err := globalBootstrap.NewKafkaManager(kafkaConfig)
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

	// Initialize all dependencies
	deps, err := bootstrap.InitDependencies(ctx)
	if err != nil {
		logger.Error("Failed to initialize dependencies", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Load API configuration
	appConfig := config.LoadAPIConfig()

	// Setup router
	r := bootstrap.RouterSetup(deps)

	logger.Info("API Server started successfully", map[string]interface{}{
		"port": appConfig.Port,
	})
	return r.Run(":" + appConfig.Port)
}
