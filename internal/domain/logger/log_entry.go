package logger

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LogEntry represents a structured log entry compatible with logrus
type LogEntry struct {
	Id        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Level     logrus.Level  `json:"level"`     // logrus.InfoLevel, logrus.ErrorLevel, etc.
	Service   string        `json:"service"`   // order-service, wallet-service, etc.
	Operation string        `json:"operation"` // CreateOrder, ReserveStock, etc.
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
func NewLogEntry(level, service, operation, message string) *LogEntry {
	return &LogEntry{
		Id:        generateId(),
		Timestamp: time.Now(),
		Level:     ParseLogLevel(level),
		Service:   service,
		Operation: operation,
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

func (e *LogEntry) String() string {

	timestamp := e.Timestamp.Format("15:04:05")

	// Add color to log level
	var colorLevel string
	switch e.Level {
	case logrus.DebugLevel:
		colorLevel = "\033[36m[DEBUG]\033[0m" // Cyan
	case logrus.InfoLevel:
		colorLevel = "\033[32m[INFO]\033[0m" // Green
	case logrus.WarnLevel:
		colorLevel = "\033[33m[WARN]\033[0m" // Yellow
	case logrus.ErrorLevel:
		colorLevel = "\033[31m[ERROR]\033[0m" // Red
	case logrus.FatalLevel:
		colorLevel = "\033[35m[FATAL]\033[0m" // Magenta
	default:
		colorLevel = fmt.Sprintf("[%s]", e.Level.String())
	}

	if e.Metadata != nil && len(e.Metadata) > 0 {
		return fmt.Sprintf("%s %s: %s (%s) %+v", colorLevel, e.Service, e.Message, timestamp, e.Metadata)
	} else {
		return fmt.Sprintf("%s %s: %s (%s)", colorLevel, e.Service, e.Message, timestamp)
	}
}

// generateID generates a unique ID for the log entry
func generateId() string {
	// Use UUID for better uniqueness
	return uuid.New().String()
}
