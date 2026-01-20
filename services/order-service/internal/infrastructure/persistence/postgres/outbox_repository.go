package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/order"
	"github.com/jmoiron/sqlx"
)

// OutboxRecord represents a raw row from the outbox table.
// It has zero knowledge of Kafka or specific messaging protocols.
type OutboxRecord struct {
	ID            string    `db:"id"`
	AggregateType string    `db:"aggregate_type"`
	AggregateID   string    `db:"aggregate_id"`
	EventType     string    `db:"event_type"`
	Payload       []byte    `db:"payload"` // Raw JSON bytes from DB
	OccurredAt    time.Time `db:"occurred_at"`
}

type OutboxRepository struct {
	db sqlx.ExtContext // Accepts both *sqlx.DB and *sqlx.Tx
}

// NewOutboxRepository creates a new repository using a connection pool
func NewOutboxRepository(db *sqlx.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func NewOutboxRepositoryWithTx(tx *sqlx.Tx) *OutboxRepository {
	return &OutboxRepository{db: tx}
}

func (r *OutboxRepository) SaveEvent(ctx context.Context, aggregateID string, event order.DomainEvent) error {
	payload, err := json.Marshal(event.ToPayload())
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	query := `
		INSERT INTO outbox (
			id, aggregate_type, aggregate_id, event_type, payload, occurred_at, status
		) VALUES (
			gen_random_uuid(), 'order', $1, $2, $3, $4, 'pending'
		)
	`
	_, err = r.db.ExecContext(ctx, query,
		aggregateID,
		event.EventType(),
		payload,
		event.OccurredAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert outbox record: %w", err)
	}
	return nil
}

// FetchPending retrieves a batch of events that haven't been published yet
func (r *OutboxRepository) FetchPending(ctx context.Context, limit int) ([]*OutboxRecord, error) {
	query := `
		SELECT id, event_type, aggregate_type, aggregate_id, payload, occurred_at
		FROM outbox
		WHERE status = 'pending'
		ORDER BY occurred_at ASC
		LIMIT $1
	`

	var records []*OutboxRecord
	if err := sqlx.SelectContext(ctx, r.db, &records, query, limit); err != nil {
		return nil, err
	}

	return records, nil
}

// MarkAsPublished updates the event status to prevent re-processing
func (r *OutboxRepository) MarkAsPublished(ctx context.Context, eventID string) error {
	query := `UPDATE outbox SET status = 'published', published_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, eventID)
	return err
}
