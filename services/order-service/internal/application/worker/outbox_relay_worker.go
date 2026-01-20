package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/infrastructure/messaging/kafka"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/infrastructure/persistence/postgres"
)

type OutboxRelayWorker struct {
	repo      *postgres.OutboxRepository
	producer  *kafka.Producer
	interval  time.Duration
	batchSize int
}

func NewOutboxRelayWorker(
	repo *postgres.OutboxRepository,
	producer *kafka.Producer,
	interval time.Duration,
	batchSize int,
) *OutboxRelayWorker {
	return &OutboxRelayWorker{
		repo:      repo,
		producer:  producer,
		interval:  interval,
		batchSize: batchSize,
	}
}

func (w *OutboxRelayWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			w.processEvents(ctx)
		}
	}
}

func (w *OutboxRelayWorker) processEvents(ctx context.Context) {
	// 1. Fetch pending events from Postgres
	records, err := w.repo.FetchPending(ctx, w.batchSize)
	if err != nil {
		return
	}

	for _, rec := range records {
		// 2. Map raw database record to Kafka-specific envelope
		var data map[string]interface{}
		if err := json.Unmarshal(rec.Payload, &data); err != nil {
			continue
		}

		msg := &kafka.EventMessage{
			EventID:       rec.ID,
			EventType:     rec.EventType,
			AggregateType: rec.AggregateType,
			AggregateID:   rec.AggregateID,
			OccurredAt:    rec.OccurredAt,
			Data:          data,
		}

		// 3. Publish via the messaging infrastructure
		if err := w.producer.Publish(ctx, msg); err != nil {
			return // Stop current batch on network failure
		}

		// 4. Finalize state in DB
		_ = w.repo.MarkAsPublished(ctx, rec.ID)
	}
}
