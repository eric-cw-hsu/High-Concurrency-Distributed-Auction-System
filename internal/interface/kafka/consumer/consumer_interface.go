package kafkaconsumer

import (
	"context"
)

// Consumer defines the common interface for all message consumers (event, log, command, etc.)
type Consumer interface {
	// Start starts the consumer with normal operation
	Start(ctx context.Context) error

	// StartWithRecovery starts the consumer with crash recovery logic
	StartWithRecovery(ctx context.Context) error

	// Stop stops the consumer gracefully
	Stop() error
}
