package bootstrap

import (
	"context"

	globalBootstrap "eric-cw-hsu.github.io/scalable-auction-system/internal/app/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/logger/handler"
	loggerMetrics "eric-cw-hsu.github.io/scalable-auction-system/internal/app/logger/metrics"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/metrics"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/storage"
	kafkaconsumer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/consumer"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	metricsInterface "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
)

// LoggerDependencies contains all dependencies for the logger service
type LoggerDependencies struct {
	MetricsServer    *metrics.StandaloneMetricsServer
	MetricsCollector metricsInterface.MetricsCollector
	MetricsHelper    *loggerMetrics.LoggerMetricsHelper

	LoggerConsumer kafkaconsumer.Consumer
}

// InitDependencies initializes all dependencies for the logger service
func InitDependencies(ctx context.Context) (*LoggerDependencies, error) {
	// Initial configuration
	loggerAppServiceConfig := config.LoadLoggerAppServiceConfig()
	kafkaConfig := config.LoadKafkaConfig()

	// Ensure Kafka Topics Exist
	if err := globalBootstrap.EnsureAllKafkaTopics(ctx, kafkaConfig); err != nil {
		return nil, err
	}

	// Initialize logger
	logStorageConfig := storage.FileStorageConfig{
		BaseDir:     loggerAppServiceConfig.LogDirPath,
		MaxFileSize: 10 * 1024 * 1024, // 10 MB
	}

	logStorage, err := storage.NewFileStorage(logStorageConfig)
	if err != nil {
		return nil, err
	}
	storageSender := logger.NewStorageSender(logStorage)
	logger.AddSender(storageSender)

	// Initialize metrics collector
	prometheusCollector := metrics.NewPrometheusCollector()

	// Initialize standalone metrics server for logger service
	metricsServer := metrics.NewStandaloneMetricsServer(loggerAppServiceConfig.MetricsPort, prometheusCollector)

	// Initialize metrics helper
	metricsHelper := loggerMetrics.NewLoggerMetricsHelper(prometheusCollector)

	// Initialize Kafka Consumer
	loggerMessageHandler := handler.NewLoggerMessageHandler()
	loggerAppReader := kafkaInfra.NewReader(kafkaConfig.Brokers, "service.logs", "logger-group")
	loggerConsumer := kafkaconsumer.NewLoggerConsumer(loggerAppReader, loggerMessageHandler, metricsHelper)

	return &LoggerDependencies{
		MetricsServer:    metricsServer,
		MetricsCollector: prometheusCollector,
		MetricsHelper:    metricsHelper,
		LoggerConsumer:   loggerConsumer,
	}, nil
}

func (d *LoggerDependencies) Shutdown() {
	logger.Info("Shutting down Logger dependencies", nil)

	if d.LoggerConsumer != nil {
		if err := d.LoggerConsumer.Stop(); err != nil {
			logger.Error("Error stopping logger consumer", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if err := d.MetricsServer.Stop(); err != nil {
		logger.Error("Failed to stop Metrics Server", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("Logger dependencies shut down successfully", nil)
}
