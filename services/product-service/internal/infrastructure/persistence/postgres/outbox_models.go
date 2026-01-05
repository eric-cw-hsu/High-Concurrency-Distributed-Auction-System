package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/samborkent/uuidv7"
)

// OutboxEvent represents an outbox event (business structure used by repository and relay)
type OutboxEvent struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	EventID       string
	Payload       map[string]interface{}
	Status        string
	CreatedAt     time.Time
	ProcessedAt   *time.Time
	RetryCount    int
	LastError     *string
	NextRetryAt   *time.Time
}

// OutboxEventModel represents the database model
type OutboxEventModel struct {
	ID            string         `db:"id"`
	AggregateType string         `db:"aggregate_type"`
	AggregateID   string         `db:"aggregate_id"`
	EventType     string         `db:"event_type"`
	EventID       string         `db:"event_id"`
	Payload       []byte         `db:"payload"`
	Status        string         `db:"status"`
	CreatedAt     time.Time      `db:"created_at"`
	ProcessedAt   sql.NullTime   `db:"processed_at"`
	RetryCount    int            `db:"retry_count"`
	LastError     sql.NullString `db:"last_error"`
	NextRetryAt   sql.NullTime   `db:"next_retry_at"`
}

// NewOutboxEvent creates a new OutboxEvent
func NewOutboxEvent(
	aggregateType string,
	aggregateID string,
	eventType string,
	payload map[string]interface{},
) *OutboxEvent {
	return &OutboxEvent{
		ID:            uuidv7.New().String(),
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		EventID:       uuidv7.New().String(),
		Payload:       payload,
		Status:        "PENDING",
		CreatedAt:     time.Now(),
		RetryCount:    0,
	}
}

// toModel converts OutboxEvent to OutboxEventModel for database operations
func (e *OutboxEvent) toModel() (*OutboxEventModel, error) {
	payloadJSON, err := json.Marshal(e.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	model := &OutboxEventModel{
		ID:            e.ID,
		AggregateType: e.AggregateType,
		AggregateID:   e.AggregateID,
		EventType:     e.EventType,
		EventID:       e.EventID,
		Payload:       payloadJSON,
		Status:        e.Status,
		CreatedAt:     e.CreatedAt,
		RetryCount:    e.RetryCount,
	}

	if e.ProcessedAt != nil {
		model.ProcessedAt = sql.NullTime{Time: *e.ProcessedAt, Valid: true}
	}
	if e.LastError != nil {
		model.LastError = sql.NullString{String: *e.LastError, Valid: true}
	}
	if e.NextRetryAt != nil {
		model.NextRetryAt = sql.NullTime{Time: *e.NextRetryAt, Valid: true}
	}

	return model, nil
}

// fromModel converts OutboxEventModel to OutboxEvent
func fromModel(model *OutboxEventModel) (*OutboxEvent, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(model.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	event := &OutboxEvent{
		ID:            model.ID,
		AggregateType: model.AggregateType,
		AggregateID:   model.AggregateID,
		EventType:     model.EventType,
		EventID:       model.EventID,
		Payload:       payload,
		Status:        model.Status,
		CreatedAt:     model.CreatedAt,
		RetryCount:    model.RetryCount,
	}

	if model.ProcessedAt.Valid {
		event.ProcessedAt = &model.ProcessedAt.Time
	}
	if model.LastError.Valid {
		event.LastError = &model.LastError.String
	}
	if model.NextRetryAt.Valid {
		event.NextRetryAt = &model.NextRetryAt.Time
	}

	return event, nil
}
