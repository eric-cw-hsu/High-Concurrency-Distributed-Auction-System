package mocks

import (
	"context"
	"sync"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/stock"
	kafkaConsumer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/consumer"
)

// MockStockEventConsumer is a mock implementation for testing
type MockStockEventConsumer struct {
	stockCache   stock.StockCache
	processedIds map[string]bool // For idempotency
	mu           sync.RWMutex
}

// NewMockStockEventConsumer creates a new mock stock event consumer
func NewMockStockEventConsumer(stockCache stock.StockCache) kafkaConsumer.EventConsumer {
	return &MockStockEventConsumer{
		stockCache:   stockCache,
		processedIds: make(map[string]bool),
	}
}

// GetSupportedEventTypes returns the list of event types this consumer can handle
func (c *MockStockEventConsumer) GetSupportedEventTypes() []string {
	return []string{"order.placed", "order.cancelled"}
}

// HandleEvent processes a domain event
func (c *MockStockEventConsumer) HandleEvent(ctx context.Context, event domain.DomainEvent) error {
	switch event.EventType() {
	case "order.placed":
		orderEvent, ok := event.(*order.OrderPlacedEvent)
		if !ok {
			return nil
		}
		return c.handleOrderPlaced(ctx, *orderEvent)
	case "order.cancelled":
		cancelEvent, ok := event.(*order.OrderCancelledEvent)
		if !ok {
			return nil
		}
		return c.handleOrderCancelled(ctx, *cancelEvent)
	}
	return nil
}

// handleOrderPlaced processes an order placed event to update stock
func (c *MockStockEventConsumer) handleOrderPlaced(ctx context.Context, event order.OrderPlacedEvent) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check for idempotency
	if c.processedIds[event.OrderId] {
		return nil // Already processed
	}

	// Decrease stock in Redis
	_, err := c.stockCache.DecreaseStock(ctx, event.StockId, event.Quantity)
	if err != nil {
		return err
	}

	// Mark as processed
	c.processedIds[event.OrderId] = true
	return nil
}

// handleOrderCancelled processes an order cancelled event to restore stock
func (c *MockStockEventConsumer) handleOrderCancelled(ctx context.Context, event order.OrderCancelledEvent) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check for idempotency
	cancelId := event.OrderId + "_cancelled"
	if c.processedIds[cancelId] {
		return nil // Already processed
	}

	// Restore stock in Redis
	err := c.stockCache.RestoreStock(ctx, event.StockId, event.Quantity)
	if err != nil {
		return err
	}

	// Mark as processed
	c.processedIds[cancelId] = true
	return nil
}

// StartWithRecovery starts the consumer with crash recovery logic
func (c *MockStockEventConsumer) StartWithRecovery(ctx context.Context) error {
	// Mock implementation - in real version this would:
	// 1. Process all pending messages from Kafka
	// 2. Then sync final state to Redis
	return nil
}

// Start starts the consumer normally
func (c *MockStockEventConsumer) Start(ctx context.Context) error {
	return nil
}

// Stop stops the consumer gracefully
func (c *MockStockEventConsumer) Stop() error {
	return nil
}
