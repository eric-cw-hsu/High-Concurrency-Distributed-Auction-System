package kafkaproducer

import (
	"context"
	"encoding/json"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/producer"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
	"github.com/samborkent/uuidv7"
	"github.com/segmentio/kafka-go"
)

// KafkaProducer handles all event publishing
// Implements the Single Responsibility Principle by focusing only on event publishing
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(writer *kafka.Writer) producer.EventProducer {
	return &KafkaProducer{
		writer: writer,
	}
}

// PublishEvent publishes a generic event to Kafka
func (p *KafkaProducer) PublishEvent(ctx context.Context, event message.Event) error {
	envelope := message.MessageEnvelope{
		MessageID:   uuidv7.New().String(),
		MessageType: event.EventType(),
		SentAt:      event.OccurredOn(),
		Version:     1,
		Event:       event,
	}

	msgBytes, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal event envelope: %w", err)
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(event.GetAggregateID()),
		Value: msgBytes,
		Time:  envelope.SentAt,
		Headers: []kafka.Header{
			{Key: "message-type", Value: []byte(envelope.MessageType)},
			{Key: "event-name", Value: []byte(event.EventName())},
			{Key: "version", Value: []byte(fmt.Sprintf("%d", envelope.Version))},
		},
	}

	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		return fmt.Errorf("failed to publish event to Kafka: %w", err)
	}

	return nil
}
