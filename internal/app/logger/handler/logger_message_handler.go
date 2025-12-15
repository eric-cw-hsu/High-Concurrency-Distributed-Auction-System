package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	sharedLogger "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

type LoggerMessageHandler struct {
}

func NewLoggerMessageHandler() *LoggerMessageHandler {
	return &LoggerMessageHandler{}
}

func (h *LoggerMessageHandler) Handle(ctx context.Context, msg message.MessageEnvelopeRaw) error {
	// Convert LogMessage to LogEntry
	var logMsg message.LogMessage
	if err := json.Unmarshal(msg.Event, &logMsg); err != nil {
		return fmt.Errorf("invalid message type: expected LogMessage")
	}

	// Convert incoming message.LogMessage to LogPayload and use StorageSender

	payload := sharedLogger.LogPayload{
		Timestamp: logMsg.Timestamp,
		Level:     sharedLogger.LogLevel(logMsg.Level),
		Service:   logMsg.Service,
		Message:   logMsg.Message,
		Fields:    logMsg.Fields,
	}

	logger.LogWithPayload(payload)

	return nil
}

var _ message.Handler = (*LoggerMessageHandler)(nil)
