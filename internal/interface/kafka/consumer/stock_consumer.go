package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/stock"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"github.com/segmentio/kafka-go"
)

// StockConsumer handles all stock-related events
// Implements the Single Responsibility Principle by focusing only on stock operations
type StockConsumer struct {
	reader    *kafka.Reader
	stockRepo stock.StockRepository
}

// NewStockConsumer creates a new stock consumer
func NewStockConsumer(reader *kafka.Reader, stockRepo stock.StockRepository) EventConsumer {
	return &StockConsumer{
		reader:    reader,
		stockRepo: stockRepo,
	}
}

// GetSupportedEventTypes returns the list of event types this consumer can handle
func (c *StockConsumer) GetSupportedEventTypes() []string {
	return []string{
		"order.placed",
	}
}

// HandleEvent processes a domain event based on its type
func (c *StockConsumer) HandleEvent(ctx context.Context, event domain.DomainEvent) error {
	switch event.EventType() {
	case "order.placed":
		orderEvent, ok := event.(*order.OrderPlacedEvent)
		if !ok {
			return fmt.Errorf("invalid event type for order.placed: %T", event)
		}
		return c.handleOrderPlaced(ctx, *orderEvent)

	default:
		return fmt.Errorf("unsupported event type: %s", event.EventType())
	}
}

// handleOrderPlaced processes order placed events to decrease stock in database
func (c *StockConsumer) handleOrderPlaced(ctx context.Context, event order.OrderPlacedEvent) error {

	// Decrease stock in database
	updatedQuantity, err := c.stockRepo.DecreaseStock(ctx, event.StockId, event.Quantity)
	if err != nil {
		return fmt.Errorf("failed to decrease stock for order %s: %w", event.OrderId, err)
	}

	logger.Info("Stock decreased in database", map[string]interface{}{
		"stock_id":         event.StockId,
		"quantity":         event.Quantity,
		"updated_quantity": updatedQuantity,
		"order_id":         event.OrderId,
	})

	return nil
}

// Start starts the consumer with normal operation
func (c *StockConsumer) Start(ctx context.Context) error {
	logger.Info("Starting stock consumer")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context done, stopping stock consumer")
			return nil

		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if err == kafka.ErrGroupClosed {
					logger.Info("Consumer group closed, exiting")
					return nil
				}
				logger.Error("Error reading message", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

			if err := c.processMessage(ctx, message); err != nil {
				logger.Error("Error processing message", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

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
func (c *StockConsumer) StartWithRecovery(ctx context.Context) error {
	logger.Info("Recovery preparation completed, starting normal operation")
	return c.Start(ctx)
}

// Stop stops the consumer gracefully
func (c *StockConsumer) Stop() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

// processMessage processes a single Kafka message
func (c *StockConsumer) processMessage(ctx context.Context, message kafka.Message) error {
	// Determine event type from message headers or topic
	eventType := message.Topic

	var event domain.DomainEvent

	switch eventType {
	case "order.placed":
		var orderEvent order.OrderPlacedEvent
		if err := json.Unmarshal(message.Value, &orderEvent); err != nil {
			return fmt.Errorf("failed to unmarshal order placed event: %w", err)
		}
		event = &orderEvent

	default:
		// Skip unsupported events
		logger.Info("Skipping unsupported event type", map[string]interface{}{
			"event_type": eventType,
		})
		return nil
	}

	return c.HandleEvent(ctx, event)
}

// processMessageWithRecovery processes a message during recovery with additional error handling
func (c *StockConsumer) processMessageWithRecovery(ctx context.Context, message kafka.Message) error {
	// Determine event type from message headers or topic
	eventType := message.Topic

	var event domain.DomainEvent

	switch eventType {
	case "order.placed":
		var orderEvent order.OrderPlacedEvent
		if err := json.Unmarshal(message.Value, &orderEvent); err != nil {
			logger.Error("Failed to unmarshal order placed event during recovery, skipping", map[string]interface{}{
				"error": err.Error(),
			})
			return nil // Skip malformed messages during recovery
		}
		event = &orderEvent

	default:
		// Skip unsupported events during recovery
		logger.Info("Skipping unsupported event type during recovery", map[string]interface{}{
			"event_type": eventType,
		})
		return nil
	}

	// Handle the event with additional recovery context
	if err := c.HandleEvent(ctx, event); err != nil {
		// During recovery, log the error but don't fail completely
		// This helps handle cases where stock might already be adjusted
		logger.Warn("Error handling event during recovery (event may have been processed before crash)", map[string]interface{}{
			"error": err.Error(),
		})
		return nil
	}

	return nil
}
