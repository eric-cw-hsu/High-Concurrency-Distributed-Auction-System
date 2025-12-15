package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	loggerDomain "eric-cw-hsu.github.io/scalable-auction-system/internal/domain/logger"
	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
	"github.com/segmentio/kafka-go"
)

// NewKafkaSender creates a Kafka log sender. It accepts a LogPayload and
// constructs the transport `message.LogMessage` before sending.
func NewKafkaSender(brokers []string, topic string) LogSender {
	writer := kafkaInfra.NewWriter(brokers, topic)
	return func(p LogPayload) error {
		msg := message.LogMessage{
			Timestamp: p.Timestamp,
			Level:     string(p.Level),
			Service:   p.Service,
			Message:   p.Message,
			Fields:    p.Fields,
		}

		// Allow senders to extract known fields into top-level properties
		if p.Fields != nil {
			if traceID, ok := p.Fields["trace_id"].(string); ok {
				msg.TraceID = traceID
			}
			if requestID, ok := p.Fields["request_id"].(string); ok {
				msg.RequestID = requestID
			}
			if userID, ok := p.Fields["user_id"].(string); ok {
				msg.UserID = userID
			}
			if operation, ok := p.Fields["operation"].(string); ok {
				msg.Operation = operation
			}
			if duration, ok := p.Fields["duration"].(time.Duration); ok {
				msg.Duration = &duration
			}
			if errorDetail, ok := p.Fields["error_detail"].(string); ok {
				msg.ErrorDetail = errorDetail
			}
		}

		logMessage := message.MessageEnvelope{
			MessageID:   fmt.Sprintf("%d", time.Now().UnixNano()),
			MessageType: "LogMessage",
			SentAt:      time.Now(),
			Version:     1,
			Event:       &msg,
		}

		data, err := json.Marshal(logMessage)
		if err != nil {
			return err
		}

		return writer.WriteMessages(context.Background(), kafka.Message{
			Value: data,
			Time:  time.Now(),
		})
	}
}

// NewConsoleSender creates a console log sender. It receives LogPayload and
// converts it to message.LogMessage for the console output helper.
func NewConsoleSender() LogSender {
	return func(p LogPayload) error {
		LogToConsole(p)
		return nil
	}
}

// NewStorageSender creates a sender that writes LogPayload into the provided
// domain LogStorage. This centralizes conversion from LogPayload -> domain.LogEntry
// and avoids duplicating conversion logic in multiple places.
func NewStorageSender(store loggerDomain.LogStorage) LogSender {
	return func(p LogPayload) error {
		// Build a domain LogEntry from the payload
		// Use the payload level as string for NewLogEntry
		operation := ""
		var metadata map[string]interface{}
		if p.Fields != nil {
			// Copy fields into metadata and pull known top-level fields
			metadata = make(map[string]interface{})
			for k, v := range p.Fields {
				switch k {
				case "operation":
					if s, ok := v.(string); ok {
						operation = s
					}
				default:
					metadata[k] = v
				}
			}
		}

		entry := loggerDomain.NewLogEntry(string(p.Level), p.Service, operation, p.Message)

		// Attach metadata and known IDs if present
		for k, v := range metadata {
			entry.AddMetadata(k, v)
		}
		if trace, ok := p.Fields["trace_id"].(string); ok {
			entry.SetTrace(trace)
		}
		if user, ok := p.Fields["user_id"].(string); ok {
			entry.SetUser(user)
		}

		// Store the entry
		return store.Store(context.Background(), entry)
	}
}
