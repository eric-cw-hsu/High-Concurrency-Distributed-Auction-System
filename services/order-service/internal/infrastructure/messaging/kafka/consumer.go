package kafka

import (
	"context"
	"encoding/json"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/config"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type EventHandler interface {
	Handle(ctx context.Context, msg *EventMessage) error
}

type Consumer struct {
	reader  *kafka.Reader
	handler EventHandler
}

func NewConsumer(cfg *config.KafkaConfig, handler EventHandler) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.ConsumerGroupID,
		Topic:    cfg.ConsumerTopic, // Simplified for single topic
		MinBytes: cfg.ConsumerMinBytes,
		MaxBytes: cfg.ConsumerMaxBytes,
		MaxWait:  cfg.ConsumerMaxWait,
	})

	return &Consumer{
		reader:  reader,
		handler: handler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	zap.L().Info("starting kafka consumer")

	for {
		// FetchMessage blocks until a message is available
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // Context cancelled, exit normally
			}
			zap.L().Error("failed to fetch message from kafka", zap.Error(err))
			continue
		}

		var event EventMessage
		if err := json.Unmarshal(m.Value, &event); err != nil {
			zap.L().Error("failed to unmarshal kafka message", zap.Error(err))
			// Commit invalid message to skip it
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		// Execute business logic via handler
		if err := c.handler.Handle(ctx, &event); err != nil {
			zap.L().Error("failed to handle event", zap.String("event_id", event.EventID), zap.Error(err))
			// In enterprise systems, we might send this to a DLQ (Dead Letter Queue) instead of retrying forever
			continue
		}

		// Commit offset only after successful processing (At-least-once)
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			zap.L().Error("failed to commit message offset", zap.Error(err))
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
