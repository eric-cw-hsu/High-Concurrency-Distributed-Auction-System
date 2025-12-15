package kafkaconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/order_processor/metrics"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"github.com/segmentio/kafka-go"
)

// OrderProcessorConsumer handles all order-related events
// Implements the Single Responsibility Principle by focusing only on order persistence
type OrderProcessorConsumer struct {
	reader        *kafka.Reader
	handler       message.Handler
	metricsHelper *metrics.OrderProcessorMetricsHelper
}

// NewOrderProcessorConsumer creates a new order consumer
func NewOrderProcessorConsumer(reader *kafka.Reader, handler message.Handler, metricsHelper *metrics.OrderProcessorMetricsHelper) Consumer {
	return &OrderProcessorConsumer{
		reader:        reader,
		handler:       handler,
		metricsHelper: metricsHelper,
	}
}

// Start starts the consumer with normal operation
func (c *OrderProcessorConsumer) Start(ctx context.Context) error {
	logger.Info("Starting order consumer")

	// Set active consumers metric
	c.metricsHelper.SetActiveProcessorOrderConsumers(1)
	defer c.metricsHelper.SetActiveProcessorOrderConsumers(0)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context done, stopping order consumer")
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
				c.metricsHelper.RecordOrderProcessingError("order.reserved", "read_error")
				continue
			}

			// Record message received
			c.metricsHelper.RecordOrderMessageReceived("order.reserved")

			var messageEnvelope message.MessageEnvelopeRaw
			if err := json.Unmarshal(msg.Value, &messageEnvelope); err != nil {
				logger.Error("Error unmarshalling message envelope", map[string]interface{}{
					"error": err.Error(),
					"topic": msg.Topic,
					"key":   string(msg.Key),
				})
				c.metricsHelper.RecordOrderProcessingError("order.reserved", "unmarshal_error")
				continue
			}

			// handle message
			if err := c.handler.Handle(ctx, messageEnvelope); err != nil {
				logger.Error("Error handling order message", map[string]interface{}{
					"error": err.Error(),
				})
				c.metricsHelper.RecordOrderProcessingError("order.reserved", "handle_error")
				continue
			}

			// Record successful processing
			c.metricsHelper.RecordOrderMessageProcessed("order.reserved", "order_placed")

			// Commit the message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				logger.Error("Error committing message", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}
}

// StartWithRecovery starts the consumer with crash recovery logic
func (c *OrderProcessorConsumer) StartWithRecovery(ctx context.Context) error {
	logger.Info("Starting order consumer with crash recovery")

	return c.Start(ctx)
}

// Stop stops the consumer gracefully
func (c *OrderProcessorConsumer) Stop() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
