package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/common/logger"
	"github.com/jmoiron/sqlx"
	"github.com/samborkent/uuidv7"
	"go.uber.org/zap"
)

// OutboxRepository manages outbox events
type OutboxRepository struct {
	db *sqlx.DB
}

// NewOutboxRepository creates a new OutboxRepository
func NewOutboxRepository(db *sqlx.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// Insert inserts a new outbox event
func (r *OutboxRepository) Insert(ctx context.Context, event *OutboxEvent) error {
	model, err := event.toModel()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO outbox_events (
			id, aggregate_type, aggregate_id, event_type, event_id,
			payload, status, created_at, retry_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.ExecContext(
		ctx, query,
		model.ID,
		model.AggregateType,
		model.AggregateID,
		model.EventType,
		model.EventID,
		model.Payload,
		model.Status,
		model.CreatedAt,
		model.RetryCount,
	)
	if err != nil {
		logger.ErrorContext(ctx, "failed to insert outbox event",
			zap.String("event_type", event.EventType),
			zap.String("aggregate_id", event.AggregateID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}

	return nil
}

// FindPending finds pending events
func (r *OutboxRepository) FindPending(ctx context.Context, limit int) ([]*OutboxEvent, error) {
	query := `
		SELECT id, aggregate_type, aggregate_id, event_type, event_id,
			   payload, status, created_at, processed_at,
			   retry_count, last_error, next_retry_at
		FROM outbox_events
		WHERE status IN ('PENDING', 'RETRY')
		  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at ASC
		LIMIT $1
	`

	var models []OutboxEventModel
	err := r.db.SelectContext(ctx, &models, query, limit)
	if err != nil {
		logger.ErrorContext(ctx, "failed to find pending outbox events",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to find pending events: %w", err)
	}

	events := make([]*OutboxEvent, 0, len(models))
	for _, model := range models {
		event, err := fromModel(&model)
		if err != nil {
			logger.ErrorContext(ctx, "failed to parse outbox event",
				zap.String("outbox_id", model.ID),
				zap.Error(err),
			)
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

// MarkAsProcessed marks event as processed
func (r *OutboxRepository) MarkAsProcessed(ctx context.Context, id string) error {
	if !uuidv7.IsValidString(id) {
		return fmt.Errorf("invalid uuid: %s", id)
	}

	query := `
		UPDATE outbox_events
		SET status = 'SENT', processed_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.ErrorContext(ctx, "failed to mark outbox event as processed",
			zap.String("outbox_id", id),
			zap.Error(err),
		)
		return fmt.Errorf("failed to mark as processed: %w", err)
	}

	return nil
}

// IncrementRetry increments retry count
func (r *OutboxRepository) IncrementRetry(ctx context.Context, id string, errorMsg string) error {
	if !uuidv7.IsValidString(id) {
		return fmt.Errorf("invalid uuid: %s", id)
	}

	query := `
		UPDATE outbox_events
		SET retry_count = retry_count + 1,
		    last_error = $1,
		    next_retry_at = NOW() + INTERVAL '1 minute' * POW(2, retry_count),
		    status = CASE WHEN retry_count >= 5 THEN 'FAILED' ELSE 'RETRY' END
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, errorMsg, id)
	if err != nil {
		logger.ErrorContext(ctx, "failed to increment outbox retry",
			zap.String("outbox_id", id),
			zap.Error(err),
		)
		return fmt.Errorf("failed to increment retry: %w", err)
	}

	return nil
}

// DeleteOldProcessed deletes old processed events
func (r *OutboxRepository) DeleteOldProcessed(ctx context.Context, olderThan time.Duration) error {
	query := `
		DELETE FROM outbox_events
		WHERE status = 'SENT'
		  AND processed_at < NOW() - $1::INTERVAL
	`

	result, err := r.db.ExecContext(ctx, query, olderThan.String())
	if err != nil {
		logger.ErrorContext(ctx, "failed to delete old outbox events",
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	rowsDeleted, _ := result.RowsAffected()
	if rowsDeleted > 0 {
		logger.InfoContext(ctx, "old outbox events deleted",
			zap.Int64("count", rowsDeleted),
		)
	}

	return nil
}
