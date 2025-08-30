package wallet

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
)

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event domain.DomainEvent) error
}
