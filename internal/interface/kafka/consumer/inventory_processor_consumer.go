package kafkaconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/inventory_processor/metrics"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"github.com/segmentio/kafka-go"
)

// InventoryProcessorConsumer handles all inventory-related events
// Implements the Single Responsibility Principle by focusing only on inventory operations
type InventoryProcessorConsumer struct {
	reader        *kafka.Reader
	metricsHelper *metrics.InventoryProcessorMetricsHelper
	handler       message.Handler
}

// NewInventoryProcessorConsumer creates a new inventory processor consumer
func NewInventoryProcessorConsumer(reader *kafka.Reader, handler message.Handler, metricsHelper *metrics.InventoryProcessorMetricsHelper) Consumer {
	return &InventoryProcessorConsumer{
		reader:        reader,
		handler:       handler,
		metricsHelper: metricsHelper,
	}
}

// Start starts the consumer with normal operation
func (c *InventoryProcessorConsumer) Start(ctx context.Context) error {
	logger.Info("Starting inventory consumer")

	// Set active consumers metric
	c.metricsHelper.SetActiveInventoryConsumers(1)
	defer c.metricsHelper.SetActiveInventoryConsumers(0)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context done, stopping inventory consumer")
			return nil

		default:
			message, err := c.reader.ReadMessage(ctx)
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
				c.metricsHelper.RecordInventoryProcessingError("order.placed", "read_error")
				continue
			}

			// Record message received
			c.metricsHelper.RecordInventoryMessageReceived("order.placed")

			if err := c.processMessage(ctx, message); err != nil {
				logger.Error("Error processing message", map[string]interface{}{
					"error": err.Error(),
				})
				c.metricsHelper.RecordInventoryProcessingError("order.placed", "process_error")
				continue
			}

			// Record successful processing
			c.metricsHelper.RecordInventoryMessageProcessed("order.placed", "inventory_update")

			// Commit the message
			if err := c.reader.CommitMessages(ctx, message); err != nil {
				logger.Error("Error committing message", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}
}

// StartWithRecovery starts the consumer with crash recovery logic
func (c *InventoryProcessorConsumer) StartWithRecovery(ctx context.Context) error {
	logger.Info("Recovery preparation completed, starting normal operation")
	return c.Start(ctx)
}

// Stop stops the consumer gracefully
func (c *InventoryProcessorConsumer) Stop() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

// processMessage processes a single Kafka message
func (c *InventoryProcessorConsumer) processMessage(ctx context.Context, msg kafka.Message) error {
	var messageEnvelope message.MessageEnvelopeRaw
	if err := json.Unmarshal(msg.Value, &messageEnvelope); err != nil {
		logger.Error("Error unmarshalling message envelope", map[string]interface{}{
			"error": err.Error(),
		})
		c.metricsHelper.RecordInventoryProcessingError("order.placed", "unmarshal_error")

		return err
	}

	// handle message
	if err := c.handler.Handle(ctx, messageEnvelope); err != nil {
		logger.Error("Error handling order message", map[string]interface{}{
			"error": err.Error(),
		})
		c.metricsHelper.RecordInventoryProcessingError("order.placed", "handle_error")

		return err
	}

	return nil
}
