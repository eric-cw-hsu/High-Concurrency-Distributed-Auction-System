package order

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
)

type EventProducer interface {
	PublishEvent(ctx context.Context, event domain.DomainEvent) error
}
