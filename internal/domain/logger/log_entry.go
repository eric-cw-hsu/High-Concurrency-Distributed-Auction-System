package logger

import (
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LogEntry represents a structured log entry compatible with logrus
type LogEntry struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Level     logrus.Level  `json:"level"`      // logrus.InfoLevel, logrus.ErrorLevel, etc.
	Service   string        `json:"service"`    // order-service, wallet-service, etc.
	EventType string        `json:"event_type"` // order.placed, wallet.deposited, etc.
	Message   string        `json:"message"`
	UserID    string        `json:"user_id,omitempty"`
	TraceID   string        `json:"trace_id,omitempty"`
	Metadata  logrus.Fields `json:"metadata,omitempty"`
}

// LogLevel utility functions
func ParseLogLevel(level string) logrus.Level {
	switch level {
	case "DEBUG":
		return logrus.DebugLevel
	case "INFO":
		return logrus.InfoLevel
	case "WARN", "WARNING":
		return logrus.WarnLevel
	case "ERROR":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// NewLogEntry creates a new log entry with basic validation
func NewLogEntry(level, service, eventType, message string) *LogEntry {
	return &LogEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		Level:     ParseLogLevel(level),
		Service:   service,
		EventType: eventType,
		Message:   message,
		Metadata:  make(logrus.Fields),
	}
}

// AddMetadata adds metadata to the log entry
func (e *LogEntry) AddMetadata(key string, value interface{}) *LogEntry {
	if e.Metadata == nil {
		e.Metadata = make(logrus.Fields)
	}
	e.Metadata[key] = value
	return e
}

// SetUser sets the user ID for the log entry
func (e *LogEntry) SetUser(userID string) *LogEntry {
	e.UserID = userID
	return e
}

// SetTrace sets the trace ID for the log entry
func (e *LogEntry) SetTrace(traceID string) *LogEntry {
	e.TraceID = traceID
	return e
}

// generateID generates a unique ID for the log entry
func generateID() string {
	// Use UUID for better uniqueness
	return uuid.New().String()
}
