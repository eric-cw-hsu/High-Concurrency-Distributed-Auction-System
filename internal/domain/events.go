package domain

import (
	"context"
	"time"
)

// Base Domain Event interface
type DomainEvent interface {
	OccurredOn() time.Time
	EventType() string
	AggregateId() string
}

// Domain Event Publisher interface
type EventPublisher interface {
	Publish(ctx context.Context, events []DomainEvent) error
}

// Domain Event Handler interface
type EventHandler interface {
	Handle(ctx context.Context, event DomainEvent) error
	CanHandle(eventType string) bool
}

// Event Bus for coordinating domain events
type EventBus interface {
	RegisterHandler(handler EventHandler)
	PublishEvents(ctx context.Context, events []DomainEvent) error
}
