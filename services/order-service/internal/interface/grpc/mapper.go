package grpc

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/order"
	orderv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/order/v1"
)

// domainOrderToProto maps the Domain Aggregate to the Protobuf Response message.
// This ensures that domain-specific logic and types stay internal to the service.
func domainOrderToProto(o *order.Order) *orderv1.OrderResponse {
	pricing := o.Pricing()

	return &orderv1.OrderResponse{
		OrderId:       o.ID().String(),
		ReservationId: o.ReservationID().String(),
		UserId:        o.UserID().String(),
		ProductId:     o.ProductID().String(),
		Quantity:      int32(o.Quantity()),

		// Monetary values extracted from the Money Value Object
		UnitPrice:  pricing.UnitPrice().Amount(),
		TotalPrice: pricing.TotalPrice().Amount(),
		Currency:   pricing.UnitPrice().Currency(),

		Status: string(o.Status()),

		// We use Unix timestamps for efficiency in high-concurrency systems
		CreatedAt: o.CreatedAt().Unix(),
		ExpiresAt: o.ExpiresAt().Unix(),
		UpdatedAt: o.UpdatedAt().Unix(),
	}
}
