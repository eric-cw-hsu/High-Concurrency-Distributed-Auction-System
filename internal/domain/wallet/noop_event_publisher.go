package wallet

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
)

// NoOpEventPublisher is a no-operation implementation of EventPublisher
// It satisfies the interface but does nothing, following the Null Object Pattern
type NoOpEventPublisher struct{}

// NewNoOpEventPublisher creates a new no-operation event publisher
func NewNoOpEventPublisher() EventPublisher {
	return &NoOpEventPublisher{}
}

// Publish implements EventPublisher interface but does nothing
func (p *NoOpEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	// Do nothing - this is a no-op implementation
	return nil
}
