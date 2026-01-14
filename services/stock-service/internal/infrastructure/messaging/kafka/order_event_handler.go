package kafka

import (
	"context"
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/common/logger"
	"go.uber.org/zap"
)

// OrderEventHandler handles order events from Order Service
type OrderEventHandler struct {
	stockService *service.StockService
}

// NewOrderEventHandler creates a new OrderEventHandler
func NewOrderEventHandler(stockService *service.StockService) *OrderEventHandler {
	return &OrderEventHandler{
		stockService: stockService,
	}
}

// Handle handles order events
func (h *OrderEventHandler) Handle(ctx context.Context, msg *EventMessage) error {
	switch msg.EventType {
	case "order.cancelled":
		return h.handleOrderCancelled(ctx, msg)
	default:
		logger.DebugContext(ctx, "unknown order event type",
			zap.String("event_type", msg.EventType),
			zap.String("event_id", msg.EventID),
		)
		return nil
	}
}

// handleOrderCancelled handles order.cancelled event
func (h *OrderEventHandler) handleOrderCancelled(ctx context.Context, msg *EventMessage) error {
	// Extract reservation_id from event data
	reservationID, ok := msg.Data["reservation_id"].(string)
	if !ok {
		logger.ErrorContext(ctx, "missing or invalid reservation_id in order.cancelled event",
			zap.String("event_id", msg.EventID),
		)
		return fmt.Errorf("missing or invalid reservation_id in event data")
	}

	logger.InfoContext(ctx, "handling order.cancelled event",
		zap.String("reservation_id", reservationID),
		zap.String("event_id", msg.EventID),
	)

	// Release reservation (return stock)
	if _, err := h.stockService.Release(ctx, reservationID); err != nil {
		logger.ErrorContext(ctx, "failed to release reservation",
			zap.String("reservation_id", reservationID),
			zap.String("event_id", msg.EventID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to release reservation: %w", err)
	}

	logger.InfoContext(ctx, "reservation released successfully",
		zap.String("reservation_id", reservationID),
		zap.String("event_id", msg.EventID),
	)

	return nil
}
