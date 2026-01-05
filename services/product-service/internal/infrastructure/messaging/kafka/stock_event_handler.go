package kafka

import (
	"context"
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/common/logger"
	"go.uber.org/zap"
)

// StockEventHandler handles stock events from Stock Service
type StockEventHandler struct {
	productService *service.ProductService
}

// NewStockEventHandler creates a new StockEventHandler
func NewStockEventHandler(productService *service.ProductService) *StockEventHandler {
	return &StockEventHandler{
		productService: productService,
	}
}

// Handle handles stock events
func (h *StockEventHandler) Handle(ctx context.Context, msg *EventMessage) error {
	switch msg.EventType {
	case "stock.depleted":
		return h.handleStockDepleted(ctx, msg)
	default:
		logger.DebugContext(ctx, "unknown stock event type",
			zap.String("event_type", msg.EventType),
			zap.String("event_id", msg.EventID),
		)
		return nil
	}
}

// handleStockDepleted handles stock.depleted event
func (h *StockEventHandler) handleStockDepleted(ctx context.Context, msg *EventMessage) error {
	productID, ok := msg.Data["product_id"].(string)
	if !ok {
		logger.ErrorContext(ctx, "missing or invalid product_id in stock.depleted event",
			zap.String("event_id", msg.EventID),
		)
		return fmt.Errorf("missing or invalid product_id in event data")
	}

	logger.InfoContext(ctx, "handling stock.depleted event",
		zap.String("product_id", productID),
		zap.String("event_id", msg.EventID),
	)

	if err := h.productService.MarkProductAsSoldOut(ctx, productID); err != nil {
		logger.ErrorContext(ctx, "failed to mark product as sold out",
			zap.String("product_id", productID),
			zap.String("event_id", msg.EventID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to mark product as sold out: %w", err)
	}

	logger.InfoContext(ctx, "product marked as sold out successfully",
		zap.String("product_id", productID),
		zap.String("event_id", msg.EventID),
	)

	return nil
}
