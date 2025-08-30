package kafkaproducer

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
)

// EventProducer defines the unified interface for publishing domain events
// Follows the Interface Segregation Principle by providing a single responsibility
type EventProducer interface {
	PublishEvent(ctx context.Context, event domain.DomainEvent) error
	GetSupportedEventTypes() []string
}
