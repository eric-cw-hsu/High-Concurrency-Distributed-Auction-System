package handler

import (
	"context"
	"encoding/json"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

// OrderMessageHandler handles order-related messages
// Responsible for event/payload type assertion and persistence

type OrderMessageHandler struct {
	Repo order.OrderRepository
}

func NewOrderMessageHandler(repo order.OrderRepository) message.Handler {
	return &OrderMessageHandler{Repo: repo}
}

func (h *OrderMessageHandler) Handle(ctx context.Context, msg message.MessageEnvelopeRaw) error {
	switch msg.MessageType {
	case "domain-event":
		var domainEvent message.DomainEvent
		if err := json.Unmarshal(msg.Event, &domainEvent); err != nil {
			logger.Error("Failed to unmarshal domain event", map[string]interface{}{
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

func (h *OrderMessageHandler) handleDomainEvent(ctx context.Context, event *message.DomainEvent) error {
	switch event.EventName() {
	case "order.reserved":
		payload, ok := event.Payload.(order.OrderReservedPayload)
		if !ok {
			logger.Error("DomainEvent payload is not OrderReservedPayload", map[string]interface{}{
				"payload": event.Payload,
			})
			return nil
		}
		return h.handleOrderReserved(ctx, payload)
	default:
		logger.Error("Unsupported domain event type", map[string]interface{}{
			"event_name": event.EventName(),
		})
		return nil
	}
}

func (h *OrderMessageHandler) handleOrderReserved(ctx context.Context, payload order.OrderReservedPayload) error {
	order := &order.Order{
		OrderID:    payload.OrderID,
		BuyerID:    payload.BuyerID,
		StockID:    payload.StockID,
		Quantity:   payload.Quantity,
		TotalPrice: payload.TotalPrice,
		CreatedAt:  payload.CreatedAt,
		UpdatedAt:  payload.UpdatedAt,
	}
	if err := h.Repo.SaveOrder(ctx, order); err != nil {
		logger.Error("Failed to create order in repository", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	return nil
}
