package bootstrap

import (
	"context"
	"database/sql"
	"fmt"

	globalBootstrap "eric-cw-hsu.github.io/scalable-auction-system/internal/app/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/order_processor/handler"
	orderProcessorMetrics "eric-cw-hsu.github.io/scalable-auction-system/internal/app/order_processor/metrics"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/metrics"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres"
	pgOrder "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/order"
	kafkaconsumer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/consumer"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	metricsInterface "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
)

// OrderConsumerDependencies contains all dependencies for the order consumer service
type OrderConsumerDependencies struct {
	MetricsServer    *metrics.StandaloneMetricsServer
	MetricsCollector metricsInterface.MetricsCollector
	MetricsHelper    *orderProcessorMetrics.OrderProcessorMetricsHelper

	DB       *sql.DB
	Consumer kafkaconsumer.Consumer
}

// InitDependencies initializes all dependencies for the order consumer service
// Initialize all dependencies (config, logger, metrics, kafka, db, consumer)
func InitDependencies(ctx context.Context) (*OrderConsumerDependencies, error) {
	// Initialize config
	serviceConfig := config.LoadOrderProcessorServiceConfig()
	kafkaConfig := config.LoadKafkaConfig()
	pgConfig := config.LoadPostgresConfig()

	fmt.Printf("[DEBUG] ServiceConfig: %v\n", serviceConfig)

	// Check Kafka Topic Existence
	if err := globalBootstrap.EnsureAllKafkaTopics(ctx, kafkaConfig); err != nil {
		return nil, err
	}

	// Initialize logger
	kafkaSender := logger.NewKafkaSender(kafkaConfig.Brokers, "service.logs")
	logger.AddSender(kafkaSender)

	// Initialize metrics collector
	prometheusCollector := metrics.NewPrometheusCollector()
	metricsServer := metrics.NewStandaloneMetricsServer(serviceConfig.MetricsPort, prometheusCollector)
	// Create metrics helper compatible with existing consumer constructors
	metricsHelper := orderProcessorMetrics.NewOrderProcessorMetricsHelper(prometheusCollector)

	// Initialize Database
	logger.Info("Initializing PostgreSQL connection")
	pg, err := postgres.NewPostgresClient(pgConfig)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Initialize Repositories
	orderRepo := pgOrder.NewPostgresOrderRepository(pg)

	// Initialize Kafka Consumer
	orderConsumerReader := kafka.NewReader(kafkaConfig.Brokers, "order", "order-consumer-group")
	orderMessageHandler := handler.NewOrderMessageHandler(orderRepo)
	orderConsumer := kafkaconsumer.NewOrderProcessorConsumer(orderConsumerReader, orderMessageHandler, metricsHelper)

	return &OrderConsumerDependencies{
		MetricsServer:    metricsServer,
		MetricsCollector: prometheusCollector,
		MetricsHelper:    metricsHelper,
		Consumer:         orderConsumer,
		DB:               pg,
	}, nil
}

func (d *OrderConsumerDependencies) Shutdown() {
	if err := d.MetricsServer.Stop(); err != nil {
		logger.Error("Error stopping metrics server", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if err := d.Consumer.Stop(); err != nil {
		logger.Error("Error stopping Kafka consumer", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if err := d.DB.Close(); err != nil {
		logger.Error("Error closing database connection", map[string]interface{}{
			"error": err.Error(),
		})
	}

}
