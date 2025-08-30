package logger

import (
	"context"
)

// LoggerRepository defines the interface for log storage operations
// This is an alias to LogStorage for backward compatibility
type LoggerRepository interface {
	LogStorage
}

// LoggerService defines the business logic interface for the logger
type LoggerService interface {
	// ProcessLog processes and stores a log entry
	ProcessLog(ctx context.Context, entry *LogEntry) error

	// QueryLogs retrieves log entries based on filter criteria
	QueryLogs(ctx context.Context, filter *LogFilter) ([]*LogEntry, error)

	// GetLogByID retrieves a specific log entry by its ID
	GetLogByID(ctx context.Context, id string) (*LogEntry, error)
}
