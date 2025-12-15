package message

import (
	"encoding/json"
	"time"
)

// MessageEnvelope is the standard wrapper for all Kafka messages
// It provides type, id, version, and payload for extensibility

type MessageEnvelope struct {
	MessageID   string    `json:"message_id"`   // identifier
	MessageType string    `json:"message_type"` // "DomainEvent", "LogMessage", "Command", ...
	SentAt      time.Time `json:"sent_at"`
	Version     int       `json:"version"` // schema version
	Event       Event     `json:"event"`   // specific event（DomainEvent, LogMessage, Command...）
}

type MessageEnvelopeRaw struct {
	MessageID   string          `json:"message_id"`   // identifier
	MessageType string          `json:"message_type"` // "DomainEvent", "LogMessage", "Command", ...
	SentAt      time.Time       `json:"sent_at"`
	Version     int             `json:"version"` // schema version
	Event       json.RawMessage `json:"event"`   // specific event（DomainEvent, LogMessage, Command...）
}
