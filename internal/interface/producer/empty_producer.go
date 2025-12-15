package producer

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
)

type EmptyProducer struct{}

func NewEmptyProducer() *EmptyProducer {
	return &EmptyProducer{}
}

func (ep *EmptyProducer) PublishEvent(ctx context.Context, event message.Event) error {
	// No operation performed
	return nil
}

var _ EventProducer = (*EmptyProducer)(nil)
