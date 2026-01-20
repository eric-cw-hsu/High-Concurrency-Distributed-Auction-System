package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/config"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

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
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  cfg.ProducerMaxAttempts,
		BatchSize:    cfg.ProducerBatchSize,
		BatchTimeout: cfg.ProducerBatchTimeout,
		Async:        false,
	}

	return &Producer{
		writer: writer,
		topic:  cfg.ProducerTopic,
	}
}

// Publish publishes an event message to Kafka
func (p *Producer) Publish(ctx context.Context, event *EventMessage) error {
	zap.L().Debug("publishing event to kafka",
		zap.String("topic", p.topic),
		zap.String("event_type", event.EventType),
		zap.String("event_id", event.EventID),
	)

	// Marshal event to JSON
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Key:   []byte(event.AggregateID), // Partition by aggregate ID
		Value: value,
	}

	// Publish
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		zap.L().Error("failed to publish event to kafka",
			zap.String("topic", p.topic),
			zap.String("event_type", event.EventType),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	zap.L().Info("event published to kafka",
		zap.String("topic", p.topic),
		zap.String("event_type", event.EventType),
		zap.String("event_id", event.EventID),
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
