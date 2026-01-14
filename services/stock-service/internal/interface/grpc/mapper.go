package grpc

import (
	stockv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/stock/v1"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/stock"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// domainStockToProto converts domain Stock to proto Stock
func domainStockToProto(s *stock.Stock) *stockv1.Stock {
	return &stockv1.Stock{
		ProductId:       s.ProductID().String(),
		Quantity:        int32(s.Quantity()),
		InitialQuantity: int32(s.InitialQuantity()),
		UpdatedAt:       timestamppb.New(s.UpdatedAt()),
	}
}

// domainReservationToProto converts domain Reservation to proto Reservation
func domainReservationToProto(r *reservation.Reservation) *stockv1.Reservation {
	proto := &stockv1.Reservation{
		Id:         r.ID().String(),
		ProductId:  r.ProductID().String(),
		UserId:     r.UserID().String(),
		Quantity:   int32(r.Quantity()),
		Status:     string(r.Status()),
		ReservedAt: timestamppb.New(r.ReservedAt()),
		ExpiredAt:  timestamppb.New(r.ExpiredAt()),
	}

	if orderID := r.OrderID(); orderID != nil {
		proto.OrderId = orderID
	}

	return proto
}
