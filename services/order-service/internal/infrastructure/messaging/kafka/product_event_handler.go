package kafka

import (
	"context"
	"fmt"

	productprice "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/product_price"
	"go.uber.org/zap"
)

type ProductEventHandler struct {
	// Dependent only on the domain interface
	syncer productprice.ProductPriceSyncer
}

func NewProductEventHandler(syncer productprice.ProductPriceSyncer) *ProductEventHandler {
	return &ProductEventHandler{
		syncer: syncer,
	}
}

func (h *ProductEventHandler) Handle(ctx context.Context, msg *EventMessage) error {
	// Filter for relevant events
	if msg.EventType != "product.published" && msg.EventType != "product.price_updated" {
		return nil
	}

	// 1. Extract data (Infra concern)
	productID := msg.AggregateID
	priceRaw, okPrice := msg.Data["price"].(float64)
	currency, okCurr := msg.Data["currency"].(string)

	if !okPrice || !okCurr {
		zap.L().Error("missing price data in product event", zap.String("event_id", msg.EventID))
		return fmt.Errorf("incomplete price data")
	}

	// 2. Delegate to the syncer (Application logic)
	return h.syncer.SyncProductPrice(ctx, productID, int64(priceRaw), currency)
}
