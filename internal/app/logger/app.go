package app

import (
	"context"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/storage"
	kafkaconsumer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/consumer"
	"github.com/segmentio/kafka-go"
)

// LoggerApp represents the logger application
type LoggerApp struct {
	config   *config.LoggerConfig
	logger   *config.Logger
	storage  storage.LogStorage
	consumer kafkaconsumer.EventConsumer
	readers  map[string]*kafka.Reader
}

// NewLoggerApp creates a new logger application instance
func NewLoggerApp(cfg *config.LoggerConfig, logger *config.Logger) (*LoggerApp, error) {
	app := &LoggerApp{
		config: cfg,
		logger: logger,
	}

	if err := app.initializeStorage(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	if err := app.initializeKafkaReaders(); err != nil {
		return nil, fmt.Errorf("failed to initialize kafka readers: %w", err)
	}

	app.initializeConsumer()

	return app, nil
}

// initializeStorage initializes the storage backend
func (app *LoggerApp) initializeStorage() error {
	switch app.config.LogStorageType {
	case "file":
		baseDir := app.config.GetFileDir()
		config := storage.FileStorageConfig{
			BaseDir:     baseDir,
			MaxFileSize: 100 * 1024 * 1024, // 100MB
		}
		var err error
		app.storage, err = storage.NewFileStorage(config, app.logger)
		if err != nil {
			return fmt.Errorf("failed to create file storage: %w", err)
		}
	default:
		return fmt.Errorf("unsupported storage type: %s (only 'file' is supported)", app.config.LogStorageType)
	}

	app.logger.WithField("storage_type", app.config.LogStorageType).Info("Storage initialized")
	return nil
}

// initializeKafkaReaders initializes Kafka readers for all configured topics
func (app *LoggerApp) initializeKafkaReaders() error {
	if len(app.config.Topics) == 0 {
		return fmt.Errorf("no topics configured")
	}

	app.readers = make(map[string]*kafka.Reader)
	for _, topic := range app.config.Topics {
		reader := kafkaInfra.NewReader([]string{app.config.KafkaBroker}, topic, "logger-service")
		app.readers[topic] = reader
	}

	app.logger.WithField("topics", app.config.Topics).Info("Kafka readers initialized")
	return nil
}

// initializeConsumer initializes the logger consumer
func (app *LoggerApp) initializeConsumer() {
	app.consumer = kafkaconsumer.NewLoggerConsumer(app.readers, app.storage, app.logger)
	app.logger.Info("Logger consumer initialized")
}

// Start starts the logger application
func (app *LoggerApp) Start(ctx context.Context) error {
	app.logger.WithField("topics", app.config.Topics).Info("Starting Logger Service...")

	if err := app.consumer.StartWithRecovery(ctx); err != nil {
		return fmt.Errorf("consumer failed: %w", err)
	}

	return nil
}

// Stop stops the logger application gracefully
func (app *LoggerApp) Stop() error {
	app.logger.Info("Stopping Logger Service...")

	if app.consumer != nil {
		if err := app.consumer.Stop(); err != nil {
			app.logger.WithError(err).Error("Error stopping consumer")
			return err
		}
	}

	// Close Kafka readers
	for topic, reader := range app.readers {
		if err := reader.Close(); err != nil {
			app.logger.WithError(err).WithField("topic", topic).Error("Error closing Kafka reader")
		}
	}

	app.logger.Info("Logger Service stopped gracefully")
	return nil
}
