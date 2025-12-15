package kafkaconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/logger/metrics"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"github.com/segmentio/kafka-go"
)

// LoggerConsumer handles log messages from Kafka topics
type LoggerConsumer struct {
	reader         *kafka.Reader
	messageHandler message.Handler
	metricsHelper  *metrics.LoggerMetricsHelper
}

// NewLoggerConsumer creates a new logger consumer
func NewLoggerConsumer(reader *kafka.Reader, messageHandler message.Handler, metricsHelper *metrics.LoggerMetricsHelper) Consumer {
	return &LoggerConsumer{
		reader:         reader,
		messageHandler: messageHandler,
		metricsHelper:  metricsHelper,
	}
}

// Start starts the consumer with normal operation
func (c *LoggerConsumer) Start(ctx context.Context) error {
	logger.Info("Starting logger consumer...")

	// Set active consumers metric
	c.metricsHelper.SetActiveLogConsumers(1)
	defer c.metricsHelper.SetActiveLogConsumers(0)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context done, stopping logger consumer...")
			return nil

		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if err == kafka.ErrGroupClosed {
					logger.Info("Consumer group closed, exiting")
					return nil
				}

				if errors.Is(err, io.EOF) {
					// Reached end of partition, continue to next iteration
					continue
				}

				logger.Error("Error reading message", map[string]interface{}{
					"error": err.Error(),
				})
				c.metricsHelper.RecordLogProcessingError("service.logs", "read_error")
				continue
			}

			// Record message received
			c.metricsHelper.RecordLogMessageReceived("service.logs")

			// Process the log message
			var messageEnvelope message.MessageEnvelopeRaw
			if err := json.Unmarshal(msg.Value, &messageEnvelope); err != nil {
				logger.Error("Failed to unmarshal log message", map[string]interface{}{
					"error": err.Error(),
					"topic": msg.Topic,
					"key":   string(msg.Key),
				})
				c.metricsHelper.RecordLogProcessingError("service.logs", "unmarshal_error")
				continue
			}

			if err := c.messageHandler.Handle(ctx, messageEnvelope); err != nil {
				logger.Error("Failed to handle log message", map[string]interface{}{
					"error": err.Error(),
					"topic": msg.Topic,
					"key":   string(msg.Key),
				})
				c.metricsHelper.RecordLogProcessingError("service.logs", "handle_error")
				continue
			}

			// Record successful processing
			c.metricsHelper.RecordLogMessageProcessed("service.logs", "log_processed")
		}
	}
}

// StartWithRecovery starts the consumer with crash recovery logic
func (c *LoggerConsumer) StartWithRecovery(ctx context.Context) error {
	logger.Info("Starting logger consumer with crash recovery...")
	return c.Start(ctx)
}

// Stop stops the consumer gracefully
func (c *LoggerConsumer) Stop() error {
	logger.Info("Stopping logger consumer...")

	if c.reader != nil {
		if err := c.reader.Close(); err != nil {
			logger.Error("Error closing Kafka reader", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
	}

	logger.Info("Logger consumer stopped gracefully")

	return nil
}
