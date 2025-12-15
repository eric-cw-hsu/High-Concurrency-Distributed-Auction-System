package producer

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
)

// EventProducer defines the unified interface for publishing domain events
// Follows the Interface Segregation Principle by providing a single responsibility
type EventProducer interface {
	PublishEvent(ctx context.Context, event message.Event) error
}
