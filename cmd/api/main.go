package main

import (
	"log"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/api"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load service-specific .env file
	if err := godotenv.Load(".env.api"); err != nil {
		if err := godotenv.Load(); err != nil {
			// Use standard log before logger initialization
			log.Println("No .env file found, relying on system env vars")
		}
	}

	// Initialize global logger
	kafkaConfig := config.LoadKafkaConfig()
	kafkaSender := logger.NewKafkaSender(kafkaConfig.Brokers, "service.logs")
	appLogger := logger.NewLogger("api", kafkaSender)
	logger.SetDefault(appLogger)

	logger.Info("API service initializing")

	// Start the API service
	if err := api.Run(); err != nil {
		logger.Fatal("Failed to run the application", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
