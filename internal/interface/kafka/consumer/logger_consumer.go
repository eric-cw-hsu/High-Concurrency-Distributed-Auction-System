package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
	loggerDomain "eric-cw-hsu.github.io/scalable-auction-system/internal/domain/logger"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/storage"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// LoggerConsumer handles log events from multiple Kafka topics
type LoggerConsumer struct {
	readers    map[string]*kafka.Reader
	logStorage storage.LogStorage
	logger     *config.Logger
}

// NewLoggerConsumer creates a new logger consumer
func NewLoggerConsumer(readers map[string]*kafka.Reader, logStorage storage.LogStorage, logger *config.Logger) EventConsumer {
	return &LoggerConsumer{
		readers:    readers,
		logStorage: logStorage,
		logger:     logger,
	}
}

// GetSupportedEventTypes returns the list of event types this consumer can handle
func (c *LoggerConsumer) GetSupportedEventTypes() []string {
	return []string{
		"audit.log",
		"order.placed",
		"order.failed",
		"wallet.deposited",
		"wallet.withdrawn",
		"stock.updated",
		"stock.depleted",
	}
}

// HandleEvent processes a domain event and converts it to a log entry
func (c *LoggerConsumer) HandleEvent(ctx context.Context, event domain.DomainEvent) error {
	logEntry := c.convertEventToLogEntry(event)

	if err := c.logStorage.Store(ctx, logEntry); err != nil {
		return fmt.Errorf("failed to store log entry: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"service":    logEntry.Service,
		"event_type": logEntry.EventType,
		"level":      "INFO",
	}).Info("Log stored successfully")

	return nil
}

// convertEventToLogEntry converts a domain event to a log entry
func (c *LoggerConsumer) convertEventToLogEntry(event domain.DomainEvent) *loggerDomain.LogEntry {
	logEntry := loggerDomain.NewLogEntry(
		"INFO",
		c.extractServiceFromEventType(event.EventType()),
		event.EventType(),
		fmt.Sprintf("Event processed: %s", event.EventType()),
	)

	// Add event-specific metadata
	logEntry.AddMetadata("aggregate_id", event.AggregateId())
	logEntry.AddMetadata("occurred_on", event.OccurredOn())

	return logEntry
}

// extractServiceFromEventType extracts service name from event type
func (c *LoggerConsumer) extractServiceFromEventType(eventType string) string {
	parts := strings.Split(eventType, ".")
	if len(parts) > 0 {
		switch parts[0] {
		case "order":
			return "order-service"
		case "wallet":
			return "wallet-service"
		case "stock":
			return "stock-service"
		case "audit":
			return "audit-service"
		default:
			return "unknown-service"
		}
	}
	return "unknown-service"
}

// Start starts the consumer with normal operation
func (c *LoggerConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting logger consumer...")

	// Create goroutines for each topic reader
	for topic, reader := range c.readers {
		go func(topic string, reader *kafka.Reader) {
			c.logger.WithField("topic", topic).Info("Starting consumer for topic")

			c.consumeFromTopic(ctx, topic, reader)
		}(topic, reader)
	}

	// Wait for context cancellation
	<-ctx.Done()
	c.logger.Info("Context done, stopping logger consumer...")

	return nil
}

// consumeFromTopic handles consumption from a specific topic
func (c *LoggerConsumer) consumeFromTopic(ctx context.Context, topic string, reader *kafka.Reader) {
	defer func() {
		if err := recover(); err != nil {
			c.logger.WithFields(logrus.Fields{
				"topic": topic,
				"error": err,
			}).Error("Consumer error")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			c.logger.WithField("topic", topic).Info("Context done, stopping consumer for topic")
			return

		default:
			message, err := reader.ReadMessage(ctx)
			if err != nil {
				if err == kafka.ErrGroupClosed {
					c.logger.WithField("topic", topic).Info("Consumer group closed for topic")
					return
				}
				c.logger.WithError(err).WithField("topic", topic).Error("Error reading message from topic")
				continue
			}

			if err := c.processMessage(ctx, topic, message); err != nil {
				c.logger.WithError(err).WithField("topic", topic).Error("Error processing message from topic")
				continue
			}

			// Commit the message
			if err := reader.CommitMessages(ctx, message); err != nil {
				c.logger.WithError(err).WithField("topic", topic).Error("Error committing message from topic")
			}
		}
	}
}

// consumeTopic consumes messages from a specific topic
func (c *LoggerConsumer) consumeTopic(ctx context.Context, topic string, reader *kafka.Reader) error {
	for {
		select {
		case <-ctx.Done():
			c.logger.WithField("topic", topic).Info("Context done, stopping consumer for topic")
			return nil

		default:
			message, err := reader.ReadMessage(ctx)
			if err != nil {
				if err == kafka.ErrGroupClosed {
					c.logger.WithField("topic", topic).Info("Consumer group closed for topic")
					return nil
				}
				c.logger.WithError(err).WithField("topic", topic).Error("Error reading message from topic")
				continue
			}

			if err := c.processMessage(ctx, topic, message); err != nil {
				c.logger.WithError(err).WithField("topic", topic).Error("Error processing message from topic")
				continue
			}

			// Commit the message
			if err := reader.CommitMessages(ctx, message); err != nil {
				c.logger.WithError(err).WithField("topic", topic).Error("Error committing message from topic")
			}
		}
	}
}

// processMessage processes a single Kafka message
func (c *LoggerConsumer) processMessage(ctx context.Context, topic string, message kafka.Message) error {
	// Try to parse as a generic log entry first
	var logEntry loggerDomain.LogEntry
	if err := json.Unmarshal(message.Value, &logEntry); err == nil {
		// If it's already a structured log entry, store it directly
		logEntry.ID = loggerDomain.NewLogEntry("", "", "", "").ID // Generate new ID
		return c.logStorage.Store(ctx, &logEntry)
	}

	// If not a structured log entry, create one from the message
	logEntry = *loggerDomain.NewLogEntry(
		"INFO",
		c.getServiceFromTopic(topic),
		c.getEventTypeFromTopic(topic),
		string(message.Value),
	)

	// Add message metadata
	logEntry.AddMetadata("topic", topic)
	logEntry.AddMetadata("kafka_offset", message.Offset)
	logEntry.AddMetadata("kafka_partition", message.Partition)

	// Extract metadata from headers
	for _, header := range message.Headers {
		switch string(header.Key) {
		case "event-type":
			logEntry.EventType = string(header.Value)
		case "user-id":
			logEntry.UserID = string(header.Value)
		case "trace-id":
			logEntry.TraceID = string(header.Value)
		default:
			logEntry.AddMetadata(string(header.Key), string(header.Value))
		}
	}

	return c.logStorage.Store(ctx, &logEntry)
}

// getServiceFromTopic maps topic name to service name
func (c *LoggerConsumer) getServiceFromTopic(topic string) string {
	switch topic {
	case "audit-logs":
		return "audit-service"
	case "order-events":
		return "order-service"
	case "wallet-events":
		return "wallet-service"
	case "stock-events":
		return "stock-service"
	default:
		return "unknown-service"
	}
}

// getEventTypeFromTopic maps topic name to default event type
func (c *LoggerConsumer) getEventTypeFromTopic(topic string) string {
	switch topic {
	case "audit-logs":
		return "audit.log"
	case "order-events":
		return "order.event"
	case "wallet-events":
		return "wallet.event"
	case "stock-events":
		return "stock.event"
	default:
		return "unknown.event"
	}
}

// StartWithRecovery starts the consumer with crash recovery logic
func (c *LoggerConsumer) StartWithRecovery(ctx context.Context) error {
	c.logger.Info("Starting logger consumer with crash recovery...")
	return c.Start(ctx)
}

// Stop stops the consumer gracefully
func (c *LoggerConsumer) Stop() error {
	c.logger.Info("Stopping logger consumer...")

	// Close all readers
	for topic, reader := range c.readers {
		if err := reader.Close(); err != nil {
			c.logger.WithError(err).WithField("topic", topic).Error("Error closing reader for topic")
		}
	}

	// Close storage
	if err := c.logStorage.Close(); err != nil {
		c.logger.WithError(err).Error("Error closing log storage")
		return err
	}

	return nil
}
