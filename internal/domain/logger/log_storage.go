package logger

import (
	"context"
)

// LogStorage defines the interface for log storage operations
type LogStorage interface {
	// Store stores a log entry
	Store(ctx context.Context, entry *LogEntry) error

	// Close closes the storage connection
	Close() error
}
