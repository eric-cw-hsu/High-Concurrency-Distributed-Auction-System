package bootstrap

import (
	"context"
	"database/sql"

	globalBootstrap "eric-cw-hsu.github.io/scalable-auction-system/internal/app/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/inventory_processor/handler"
	stockMetrics "eric-cw-hsu.github.io/scalable-auction-system/internal/app/inventory_processor/metrics"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/metrics"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres"
	pgstock "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/stock"
	kafkaconsumer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/consumer"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	metricsInterface "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
)

// StockConsumerDependencies contains all dependencies for the stock consumer service
type StockConsumerDependencies struct {
	MetricsServer    *metrics.StandaloneMetricsServer
	MetricsCollector metricsInterface.MetricsCollector
	MetricsHelper    *stockMetrics.InventoryProcessorMetricsHelper

	DB       *sql.DB
	Consumer kafkaconsumer.Consumer
}

// InitDependencies initializes all dependencies for the stock consumer service
func InitDependencies(ctx context.Context) (*StockConsumerDependencies, error) {
	// Initial configuration
	stockServiceConfig := config.LoadInventoryProcessorServiceConfig()
	kafkaConfig := config.LoadKafkaConfig()
	pgConfig := config.LoadPostgresConfig()

	// Ensure Kafka Topics Exist
	if err := globalBootstrap.EnsureAllKafkaTopics(ctx, kafkaConfig); err != nil {
		return nil, err
	}

	// Initialize logger
	kafkaLogSender := logger.NewKafkaSender(kafkaConfig.Brokers, "service.logs")
	logger.AddSender(kafkaLogSender)

	// Initialize metrics collector
	prometheusCollector := metrics.NewPrometheusCollector()
	metricsServer := metrics.NewStandaloneMetricsServer(stockServiceConfig.MetricsPort, prometheusCollector)
	metricsHelper := stockMetrics.NewInventoryProcessorMetricsHelper(prometheusCollector)

	// Initialize Database
	logger.Info("Initializing PostgreSQL connection")
	pg, err := postgres.NewPostgresClient(pgConfig)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Initialize stock repository
	stockRepo := pgstock.NewPostgresStockRepository(pg)
	logger.Info("PostgreSQL connection established")

	// Initialize Kafka reader for stock events
	logger.Info("Initializing Kafka consumer")
	kafkaReader := kafkaInfra.NewReader(kafkaConfig.Brokers, "order.placed", "stock-service")
	defer kafkaReader.Close()

	// Create stock consumer
	stockMessageHandler := handler.NewInventoryMessageHandler(stockRepo)
	stockConsumer := kafkaconsumer.NewInventoryProcessorConsumer(kafkaReader, stockMessageHandler, metricsHelper)
	logger.Info("Kafka consumer initialized")

	return &StockConsumerDependencies{
		MetricsServer:    metricsServer,
		MetricsCollector: prometheusCollector,
		MetricsHelper:    metricsHelper,

		DB:       pg,
		Consumer: stockConsumer,
	}, nil
}

func (d *StockConsumerDependencies) Shutdown() {
	logger.Info("Shutting down Stock Consumer dependencies")

	if d.Consumer != nil {
		if err := d.Consumer.Stop(); err != nil {
			logger.Error("Error stopping Kafka consumer", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			logger.Error("Error closing database connection", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if d.MetricsServer != nil {
		if err := d.MetricsServer.Stop(); err != nil {
			logger.Error("Error stopping metrics server", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	logger.Info("Stock Consumer dependencies shut down completed")
}
