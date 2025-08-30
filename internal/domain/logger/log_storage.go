package logger

import (
	"context"
	"time"
)

// LogStorage defines the interface for log storage operations
type LogStorage interface {
	// Store stores a log entry
	Store(ctx context.Context, entry *LogEntry) error

	// Query retrieves log entries based on filter criteria
	Query(ctx context.Context, filter *LogFilter) ([]*LogEntry, error)

	// Close closes the storage connection
	Close() error
}

// LogFilter defines filter criteria for querying logs
type LogFilter struct {
	// Time range
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`

	// Log level filter
	Levels []string `json:"levels,omitempty"`

	// Service filter
	Services []string `json:"services,omitempty"`

	// Event type filter
	EventTypes []string `json:"event_types,omitempty"`

	// User ID filter
	UserIDs []string `json:"user_ids,omitempty"`

	// Trace ID filter
	TraceIDs []string `json:"trace_ids,omitempty"`

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// NewLogFilter creates a new log filter with default values
func NewLogFilter() *LogFilter {
	return &LogFilter{
		Limit: 100, // Default limit
	}
}

// SetTimeRange sets the time range for the filter
func (f *LogFilter) SetTimeRange(start, end time.Time) *LogFilter {
	f.StartTime = &start
	f.EndTime = &end
	return f
}

// SetLevels sets the log levels to filter by
func (f *LogFilter) SetLevels(levels ...string) *LogFilter {
	f.Levels = levels
	return f
}

// SetServices sets the services to filter by
func (f *LogFilter) SetServices(services ...string) *LogFilter {
	f.Services = services
	return f
}

// SetEventTypes sets the event types to filter by
func (f *LogFilter) SetEventTypes(eventTypes ...string) *LogFilter {
	f.EventTypes = eventTypes
	return f
}

// SetPagination sets the pagination parameters
func (f *LogFilter) SetPagination(limit, offset int) *LogFilter {
	f.Limit = limit
	f.Offset = offset
	return f
}
