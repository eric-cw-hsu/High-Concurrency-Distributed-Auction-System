package kafka

import (
	"context"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/persistence/redis"
	"go.uber.org/zap"
)

// ProductEventHandler handles product lifecycle events
type ProductEventHandler struct {
	productStateRepo *redis.ProductStateRepository
}

// NewProductEventHandler creates a new product event handler
func NewProductEventHandler(productStateRepo *redis.ProductStateRepository) *ProductEventHandler {
	return &ProductEventHandler{
		productStateRepo: productStateRepo,
	}
}

// Handle processes product events
func (h *ProductEventHandler) Handle(ctx context.Context, msg *EventMessage) error {
	zap.L().Debug("handling product event",
		zap.String("event_type", msg.EventType),
		zap.String("event_id", msg.EventID),
		zap.String("aggregate_id", msg.AggregateID),
	)

	// Filter: only handle product lifecycle events
	switch msg.EventType {
	case "product.published":
		return h.handleProductPublished(ctx, msg)

	case "product.deactivated":
		return h.handleProductDeactivated(ctx, msg)

	case "product.deleted":
		return h.handleProductDeleted(ctx, msg)

	default:
		// Ignore other product events (info.updated, pricing.updated, etc.)
		zap.L().Debug("ignoring non-lifecycle event",
			zap.String("event_type", msg.EventType),
		)
		return nil
	}
}

func (h *ProductEventHandler) handleProductPublished(ctx context.Context, msg *EventMessage) error {
	productID, ok := msg.Data["product_id"].(string)
	if !ok {
		zap.L().Error("missing product_id in event data",
			zap.String("event_id", msg.EventID),
		)
		return nil // Skip this message
	}

	zap.L().Info("marking product as active",
		zap.String("product_id", productID),
		zap.String("event_id", msg.EventID),
	)

	if err := h.productStateRepo.MarkActive(ctx, productID); err != nil {
		zap.L().Error("failed to mark product as active",
			zap.String("product_id", productID),
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("product marked as active",
		zap.String("product_id", productID),
	)

	return nil
}

func (h *ProductEventHandler) handleProductDeactivated(ctx context.Context, msg *EventMessage) error {
	productID, ok := msg.Data["product_id"].(string)
	if !ok {
		zap.L().Error("missing product_id in event data",
			zap.String("event_id", msg.EventID),
		)
		return nil
	}

	zap.L().Info("marking product as inactive",
		zap.String("product_id", productID),
		zap.String("event_id", msg.EventID),
	)

	if err := h.productStateRepo.MarkInactive(ctx, productID); err != nil {
		zap.L().Error("failed to mark product as inactive",
			zap.String("product_id", productID),
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("product marked as inactive",
		zap.String("product_id", productID),
	)

	return nil
}

func (h *ProductEventHandler) handleProductDeleted(ctx context.Context, msg *EventMessage) error {
	productID, ok := msg.Data["product_id"].(string)
	if !ok {
		zap.L().Error("missing product_id in event data",
			zap.String("event_id", msg.EventID),
		)
		return nil
	}

	zap.L().Info("removing product from active set",
		zap.String("product_id", productID),
		zap.String("event_id", msg.EventID),
	)

	if err := h.productStateRepo.Remove(ctx, productID); err != nil {
		zap.L().Error("failed to remove product",
			zap.String("product_id", productID),
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("product removed from active set",
		zap.String("product_id", productID),
	)

	return nil
}
