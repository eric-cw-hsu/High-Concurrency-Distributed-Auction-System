package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

func NewReader(brokers []string, topic string, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:         brokers,
		Topic:           topic,
		GroupID:         groupID,
		MinBytes:        10e3,
		MaxBytes:        10e6,
		CommitInterval:  time.Second,
		StartOffset:     kafka.FirstOffset,
		ReadLagInterval: -1,
	})
}
