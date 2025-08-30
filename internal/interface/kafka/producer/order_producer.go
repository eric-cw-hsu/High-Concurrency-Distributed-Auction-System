package kafkaproducer

import (
	"context"
	"encoding/json"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"github.com/segmentio/kafka-go"
)

// OrderProducer handles all order-related event publishing
// Implements the Single Responsibility Principle by focusing only on order events
type OrderProducer struct {
	writer *kafka.Writer
}

// NewOrderProducer creates a new order producer
func NewOrderProducer(writer *kafka.Writer) EventProducer {
	return &OrderProducer{
		writer: writer,
	}
}

// GetSupportedEventTypes returns the list of event types this producer can handle
func (p *OrderProducer) GetSupportedEventTypes() []string {
	return []string{
		"order.placed",
		"order.cancelled",
	}
}

// PublishEvent publishes a domain event based on its type
func (p *OrderProducer) PublishEvent(ctx context.Context, event domain.DomainEvent) error {
	switch event.EventType() {
	case "order.placed":
		orderEvent, ok := event.(*order.OrderPlacedEvent)
		if !ok {
			return fmt.Errorf("invalid event type for order.placed: %T", event)
		}
		return p.publishOrderPlaced(ctx, *orderEvent)

	case "order.cancelled":
		orderEvent, ok := event.(*order.OrderCancelledEvent)
		if !ok {
			return fmt.Errorf("invalid event type for order.cancelled: %T", event)
		}
		return p.publishOrderCancelled(ctx, *orderEvent)

	default:
		return fmt.Errorf("unsupported event type: %s", event.EventType())
	}
}

// publishOrderPlaced publishes order placed events
func (p *OrderProducer) publishOrderPlaced(ctx context.Context, event order.OrderPlacedEvent) error {
	msg, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal order placed event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.AggregateId()),
		Value: msg,
		Time:  event.OccurredOn(),
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.EventType())},
			{Key: "aggregate-id", Value: []byte(event.AggregateId())},
		},
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to publish order placed event: %w", err)
	}

	return nil
}

// publishOrderCancelled publishes order cancelled events
func (p *OrderProducer) publishOrderCancelled(ctx context.Context, event order.OrderCancelledEvent) error {
	msg, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal order cancelled event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.AggregateId()),
		Value: msg,
		Time:  event.OccurredOn(),
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.EventType())},
			{Key: "aggregate-id", Value: []byte(event.AggregateId())},
		},
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to publish order cancelled event: %w", err)
	}

	return nil
}

// PublishOrder maintains backward compatibility
// Deprecated: Use PublishEvent instead
func (p *OrderProducer) PublishOrder(ctx context.Context, event domain.DomainEvent) error {
	return p.PublishEvent(ctx, event)
}
