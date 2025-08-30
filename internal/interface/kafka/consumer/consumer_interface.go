package kafkaconsumer

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
)

// EventConsumer defines the common interface for all domain event consumers
type EventConsumer interface {
	// HandleEvent processes a domain event
	HandleEvent(ctx context.Context, event domain.DomainEvent) error

	// Start starts the consumer with normal operation
	Start(ctx context.Context) error

	// StartWithRecovery starts the consumer with crash recovery logic
	StartWithRecovery(ctx context.Context) error

	// Stop stops the consumer gracefully
	Stop() error

	// GetSupportedEventTypes returns the list of event types this consumer can handle
	GetSupportedEventTypes() []string
}
