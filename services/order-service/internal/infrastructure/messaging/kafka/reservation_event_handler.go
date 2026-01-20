package kafka

import (
	"context"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/order"
	"go.uber.org/zap"
)

// ReservationEventHandler handles reservation events from Stock Service
type ReservationEventHandler struct {
	orderCreator order.Creator
}

// NewReservationEventHandler creates a new reservation event handler
func NewReservationEventHandler(orderCreator order.Creator) *ReservationEventHandler {
	return &ReservationEventHandler{
		orderCreator: orderCreator,
	}
}

// Handle processes reservation events
func (h *ReservationEventHandler) Handle(ctx context.Context, msg *EventMessage) error {
	zap.L().Debug("handling reservation event",
		zap.String("event_type", msg.EventType),
		zap.String("event_id", msg.EventID),
	)

	switch msg.EventType {
	case "stock.reserved":
		return h.handleReservationCreated(ctx, msg)

	default:
		// Ignore other reservation events
		zap.L().Debug("ignoring reservation event",
			zap.String("event_type", msg.EventType),
		)
		return nil
	}
}

func (h *ReservationEventHandler) handleReservationCreated(ctx context.Context, msg *EventMessage) error {
	// Extract data from event
	reservationID, ok := msg.Data["reservation_id"].(string)
	if !ok {
		zap.L().Error("missing reservation_id in event")
		return nil // Skip this message
	}

	userID, ok := msg.Data["user_id"].(string)
	if !ok {
		zap.L().Error("missing user_id in event")
		return nil
	}

	productID, ok := msg.Data["product_id"].(string)
	if !ok {
		zap.L().Error("missing product_id in event")
		return nil
	}

	quantity, ok := msg.Data["quantity"].(float64) // JSON numbers are float64
	if !ok {
		zap.L().Error("missing quantity in event")
		return nil
	}

	zap.L().Info("creating order from reservation",
		zap.String("reservation_id", reservationID),
		zap.String("user_id", userID),
		zap.String("product_id", productID),
		zap.Int("quantity", int(quantity)),
	)

	// Create order
	err := h.orderCreator.CreateOrderFromReservation(
		ctx,
		reservationID,
		userID,
		productID,
		int(quantity),
	)

	if err != nil {
		zap.L().Error("failed to create order from reservation",
			zap.String("reservation_id", reservationID),
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("order created from reservation",
		zap.String("reservation_id", reservationID),
	)

	return nil
}
