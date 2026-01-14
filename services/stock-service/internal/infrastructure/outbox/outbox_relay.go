package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/messaging/kafka"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/persistence/postgres"
	"go.uber.org/zap"
)

// OutboxRelay is responsible for relaying outbox events to Kafka
type OutboxRelay struct {
	outboxRepo *postgres.OutboxRepository
	producer   *kafka.Producer
	config     *config.OutboxConfig
}

// NewOutboxRelay creates a new OutboxRelay
func NewOutboxRelay(
	outboxRepo *postgres.OutboxRepository,
	producer *kafka.Producer,
	cfg *config.OutboxConfig,
) *OutboxRelay {
	return &OutboxRelay{
		outboxRepo: outboxRepo,
		producer:   producer,
		config:     cfg,
	}
}

// Start starts the outbox relay worker
func (r *OutboxRelay) Start(ctx context.Context) error {
	ticker := time.NewTicker(r.config.PollInterval)
	defer ticker.Stop()

	zap.L().Info("outbox relay started",
		zap.Int("batch_size", r.config.BatchSize),
		zap.Duration("poll_interval", r.config.PollInterval),
	)

	for {
		select {
		case <-ticker.C:
			if err := r.processOutbox(ctx); err != nil {
				zap.L().Error("error processing outbox",
					zap.Error(err),
				)
			}

		case <-ctx.Done():
			zap.L().Info("outbox relay stopped")
			return ctx.Err()
		}
	}
}

// StartCleanup starts the cleanup worker
func (r *OutboxRelay) StartCleanup(ctx context.Context) error {
	ticker := time.NewTicker(r.config.CleanupPeriod)
	defer ticker.Stop()

	zap.L().Info("outbox cleanup started",
		zap.Duration("cleanup_age", r.config.CleanupAge),
		zap.Duration("cleanup_period", r.config.CleanupPeriod),
	)

	for {
		select {
		case <-ticker.C:
			if err := r.outboxRepo.DeleteOldProcessed(ctx, r.config.CleanupAge); err != nil {
				zap.L().Error("error cleaning up outbox",
					zap.Error(err),
				)
			} else {
				zap.L().Debug("outbox cleanup completed")
			}

		case <-ctx.Done():
			zap.L().Info("outbox cleanup stopped")
			return ctx.Err()
		}
	}
}

// processOutbox processes pending outbox events
func (r *OutboxRelay) processOutbox(ctx context.Context) error {
	events, err := r.outboxRepo.FindPending(ctx, r.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to find pending events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	zap.L().Debug("processing outbox events",
		zap.Int("count", len(events)),
	)

	successCount := 0
	for _, event := range events {
		if err := r.processEvent(ctx, event); err != nil {
			zap.L().Error("failed to process outbox event",
				zap.String("event_id", event.EventID),
				zap.String("event_type", event.EventType),
				zap.Error(err),
			)
			if err := r.outboxRepo.IncrementRetry(ctx, event.ID, err.Error()); err != nil {
				zap.L().Error("failed to increment retry count",
					zap.String("outbox_id", event.ID),
					zap.Error(err),
				)
			}
			continue
		}

		if err := r.outboxRepo.MarkAsProcessed(ctx, event.ID); err != nil {
			zap.L().Error("failed to mark outbox event as processed",
				zap.String("outbox_id", event.ID),
				zap.Error(err),
			)
		} else {
			successCount++
		}
	}

	zap.L().Info("outbox events processed",
		zap.Int("success", successCount),
		zap.Int("total", len(events)),
	)

	return nil
}

// processEvent processes a single outbox event
func (r *OutboxRelay) processEvent(ctx context.Context, event *postgres.OutboxEvent) error {
	kafkaMsg := &kafka.EventMessage{
		EventID:     event.EventID,
		EventType:   event.EventType,
		AggregateID: event.AggregateID,
		OccurredAt:  event.CreatedAt,
		Data:        event.Payload,
	}

	if err := r.producer.Publish(ctx, kafkaMsg); err != nil {
		return fmt.Errorf("failed to publish to kafka: %w", err)
	}

	return nil
}
