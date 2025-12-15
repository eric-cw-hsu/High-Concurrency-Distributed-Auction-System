package message

import "time"

// LogMessage is the standard schema for all log events
// All services should use this struct for centralized logging

type LogMessage struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`     // "INFO", "ERROR", etc.
	Service     string                 `json:"service"`   // e.g. "order-service"
	Operation   string                 `json:"operation"` // e.g. "CreateOrder"
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Duration    *time.Duration         `json:"duration,omitempty"`
	ErrorDetail string                 `json:"error_detail,omitempty"`
}

func (e *LogMessage) OccurredOn() time.Time {
	return e.Timestamp
}

func (e *LogMessage) EventType() string {
	return "log-message"
}

func (e *LogMessage) EventName() string {
	return e.Service + "." + e.Operation
}

func (e *LogMessage) GetAggregateID() string {
	return e.RequestID
}

// Ensure LogMessage implements the Event interface
var _ Event = &LogMessage{}
