package handler

import (
	"context"
	"encoding/json"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/stock"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

// InventoryMessageHandler handles inventory-related messages
// Responsible for event/payload type assertion and persistence

type InventoryMessageHandler struct {
	Repo stock.StockRepository
}

func NewInventoryMessageHandler(repo stock.StockRepository) message.Handler {
	return &InventoryMessageHandler{Repo: repo}
}

func (h *InventoryMessageHandler) Handle(ctx context.Context, msg message.MessageEnvelopeRaw) error {
	switch msg.MessageType {
	case "domain-event":
		var domainEvent message.DomainEvent
		if err := json.Unmarshal(msg.Event, &domainEvent); err != nil {
			logger.Error("Envelope event is not DomainEvent", map[string]interface{}{
				"event": msg.Event,
				"error": err.Error(),
			})
			return nil
		}

		return h.handleDomainEvent(ctx, &domainEvent)
	default:
		logger.Error("Unsupported message type", map[string]interface{}{
			"message_type": msg.MessageType,
		})
		return nil
	}
}

func (h *InventoryMessageHandler) handleDomainEvent(ctx context.Context, event *message.DomainEvent) error {
	switch event.EventName() {
	case "order.reserved":
		payload, ok := event.Payload.(stock.StockUpdatedPayload)
		if !ok {
			logger.Error("DomainEvent payload is not StockUpdatedPayload", map[string]interface{}{
				"payload": event.Payload,
			})
			return nil
		}
		return h.handleStockUpdated(ctx, payload)
	default:
		logger.Error("Unsupported domain event type", map[string]interface{}{
			"event_name": event.EventName(),
		})
		return nil
	}
}

func (h *InventoryMessageHandler) handleStockUpdated(ctx context.Context, payload stock.StockUpdatedPayload) error {
	stockEntity := &stock.Stock{
		ID:        payload.StockID,
		Quantity:  payload.Quantity,
		UpdatedAt: payload.UpdatedAt,
	}
	_, err := h.Repo.UpdateStockQuantity(ctx, stockEntity)
	if err != nil {
		logger.Error("Failed to update stock in repository", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	return nil
}
