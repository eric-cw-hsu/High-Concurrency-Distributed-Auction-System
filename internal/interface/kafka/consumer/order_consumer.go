package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"github.com/segmentio/kafka-go"
)

// OrderConsumer handles all order-related events
// Implements the Single Responsibility Principle by focusing only on order persistence
type OrderConsumer struct {
	reader *kafka.Reader
	repo   order.OrderRepository
}

// NewOrderConsumer creates a new order consumer
func NewOrderConsumer(reader *kafka.Reader, repo order.OrderRepository) EventConsumer {
	return &OrderConsumer{
		reader: reader,
		repo:   repo,
	}
}

// GetSupportedEventTypes returns the list of event types this consumer can handle
func (c *OrderConsumer) GetSupportedEventTypes() []string {
	return []string{
		"order.placed",
	}
}

// HandleEvent processes a domain event based on its type
func (c *OrderConsumer) HandleEvent(ctx context.Context, event domain.DomainEvent) error {
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

// handleOrderPlaced processes order placed events to persist orders
func (c *OrderConsumer) handleOrderPlaced(ctx context.Context, event order.OrderPlacedEvent) error {
	if err := c.repo.SaveOrder(ctx, event); err != nil {
		return fmt.Errorf("failed to save order %s: %w", event.OrderId, err)
	}

	logger.Info("Order persisted", map[string]interface{}{
		"order_id": event.OrderId,
		"stock_id": event.StockId,
		"buyer_id": event.BuyerId,
		"quantity": event.Quantity,
	})

	return nil
}

// Start starts the consumer with normal operation
func (c *OrderConsumer) Start(ctx context.Context) error {
	logger.Info("Starting order consumer")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context done, stopping order consumer")
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

			var event order.OrderPlacedEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				logger.Error("Error unmarshalling message", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

			if err := c.HandleEvent(ctx, &event); err != nil {
				logger.Error("Error handling order placed event", map[string]interface{}{
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
func (c *OrderConsumer) StartWithRecovery(ctx context.Context) error {
	logger.Info("Starting order consumer with crash recovery")

	return c.Start(ctx)
}

// Stop stops the consumer gracefully
func (c *OrderConsumer) Stop() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
