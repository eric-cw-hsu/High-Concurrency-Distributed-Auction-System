package kafka

import "time"

// EventMessage represents a generic event message for Kafka
type EventMessage struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	AggregateType string                 `json:"aggregate_type"`
	AggregateID   string                 `json:"aggregate_id"`
	OccurredAt    time.Time              `json:"occurred_at"`
	Data          map[string]interface{} `json:"data"`
}
