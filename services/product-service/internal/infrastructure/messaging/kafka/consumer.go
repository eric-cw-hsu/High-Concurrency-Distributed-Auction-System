package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/config"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// EventHandler handles consumed events
type EventHandler interface {
	Handle(ctx context.Context, msg *EventMessage) error
}

// Consumer wraps Kafka consumer for consuming events
type Consumer struct {
	reader  *kafka.Reader
	handler EventHandler
	topic   string
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg *config.KafkaConfig, handler EventHandler) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.ConsumerTopic,
		GroupID:        cfg.ConsumerGroupID,
		MinBytes:       1e3,
		MaxBytes:       10e6,
		StartOffset:    kafka.LastOffset,
		CommitInterval: 0,
	})

	return &Consumer{
		reader:  reader,
		handler: handler,
		topic:   cfg.ConsumerTopic,
	}
}

// Start starts consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	zap.L().Info("kafka consumer started",
		zap.String("topic", c.topic),
	)

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				zap.L().Info("kafka consumer shutting down")
				return nil
			}
			zap.L().Error("failed to fetch kafka message",
				zap.String("topic", c.topic),
				zap.Error(err),
			)
			continue
		}

		var eventMsg EventMessage
		if err := json.Unmarshal(msg.Value, &eventMsg); err != nil {
			zap.L().Error("failed to unmarshal kafka message",
				zap.String("topic", c.topic),
				zap.Error(err),
			)
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				zap.L().Error("failed to commit bad message",
					zap.Error(err),
				)
			}
			continue
		}

		zap.L().Info("kafka message received",
			zap.String("topic", c.topic),
			zap.String("event_type", eventMsg.EventType),
			zap.String("event_id", eventMsg.EventID),
		)

		if err := c.handler.Handle(ctx, &eventMsg); err != nil {
			zap.L().Error("failed to handle kafka message",
				zap.String("topic", c.topic),
				zap.String("event_type", eventMsg.EventType),
				zap.String("event_id", eventMsg.EventID),
				zap.Error(err),
			)
			// TODO: Implement retry logic or dead letter queue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			zap.L().Error("failed to commit kafka message",
				zap.String("topic", c.topic),
				zap.String("event_id", eventMsg.EventID),
				zap.Error(err),
			)
		}
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}
	return nil
}
