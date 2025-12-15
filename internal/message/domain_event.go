package message

import "time"

// DomainEvent is the standard event schema for all business events
// Versioned for schema evolution
// All services should use this struct for event-driven communication

type DomainEvent struct {
	Name        string      `json:"name"`         // e.g. "OrderPlaced", "StockReserved"
	AggregateID string      `json:"aggregate_id"` // e.g. orderId, stockId
	OccurredAt  time.Time   `json:"occurred_at"`
	Payload     interface{} `json:"payload"` // event-specific data
	Version     int         `json:"version"`
}

func (e *DomainEvent) OccurredOn() time.Time {
	return e.OccurredAt
}

func (e *DomainEvent) EventType() string {
	return "domain-event"
}

func (e *DomainEvent) EventName() string {
	return e.Name
}

func (e *DomainEvent) GetAggregateID() string {
	return e.AggregateID
}

// Ensure DomainEvent implements the Event interface
var _ Event = &DomainEvent{}
