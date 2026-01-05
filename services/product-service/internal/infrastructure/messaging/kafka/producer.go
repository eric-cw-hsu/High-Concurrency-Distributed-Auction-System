package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/config"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// EventMessage represents a Kafka event message
type EventMessage struct {
	EventID     string                 `json:"event_id"`
	EventType   string                 `json:"event_type"`
	AggregateID string                 `json:"aggregate_id"`
	OccurredAt  time.Time              `json:"occurred_at"`
	Data        map[string]interface{} `json:"data"`
}

// Producer wraps Kafka producer for publishing events
type Producer struct {
	writer *kafka.Writer
	topic  string
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg *config.KafkaConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.ProducerTopic,
		Balancer:     &kafka.Hash{}, // Partition by key (aggregate ID)
		MaxAttempts:  cfg.ProducerMaxAttempts,
		BatchSize:    cfg.ProducerBatchSize,
		BatchTimeout: cfg.ProducerBatchTimeout,
		Compression:  kafka.Snappy,
		// Idempotent writes (at-least-once delivery)
		RequiredAcks: kafka.RequireAll,
	}

	return &Producer{
		writer: writer,
		topic:  cfg.ProducerTopic,
	}
}

// Publish publishes an event message to Kafka
func (p *Producer) Publish(ctx context.Context, msg *EventMessage) error {
	logger.DebugContext(ctx, "publishing kafka message",
		zap.String("topic", p.topic),
		zap.String("event_type", msg.EventType),
		zap.String("event_id", msg.EventID),
		zap.String("aggregate_id", msg.AggregateID),
	)

	payload, err := json.Marshal(msg)
	if err != nil {
		logger.ErrorContext(ctx, "failed to marshal kafka message",
			zap.String("event_type", msg.EventType),
			zap.String("error", err.Error()),
		)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(msg.AggregateID),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(msg.EventType)},
			{Key: "event_id", Value: []byte(msg.EventID)},
		},
		Time: msg.OccurredAt,
	}

	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		logger.ErrorContext(ctx, "kafka publish failed",
			zap.String("topic", p.topic),
			zap.String("event_type", msg.EventType),
			zap.String("event_id", msg.EventID),
			zap.String("error", err.Error()),
		)
		return fmt.Errorf("failed to write message: %w", err)
	}

	logger.InfoContext(ctx, "kafka message published",
		zap.String("topic", p.topic),
		zap.String("event_type", msg.EventType),
		zap.String("event_id", msg.EventID),
	)

	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %w", err)
	}
	return nil
}
