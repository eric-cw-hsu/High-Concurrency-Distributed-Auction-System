package grpc

import (
	"context"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/order"
	pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/order/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	service order.Service
}

func NewOrderHandler(svc order.Service) *OrderHandler {
	return &OrderHandler{
		service: svc,
	}
}

// GetOrder handles the synchronous retrieval of a single order by ID
func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	o, err := h.service.GetOrder(ctx, req.OrderId)
	if err != nil {
		if err == order.ErrOrderNotFound {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, "failed to retrieve order")
	}

	return domainOrderToProto(o), nil
}

// ListUserOrders handles paginated queries for orders belonging to a specific user
func (h *OrderHandler) ListUserOrders(ctx context.Context, req *pb.ListUserOrdersRequest) (*pb.ListUserOrdersResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // Default page size
	}
	offset := int(req.Offset)

	orders, err := h.service.ListUserOrders(ctx, req.UserId, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list user orders")
	}

	resp := &pb.ListUserOrdersResponse{
		Orders: make([]*pb.OrderResponse, 0, len(orders)),
	}

	for _, o := range orders {
		resp.Orders = append(resp.Orders, domainOrderToProto(o))
	}

	return resp, nil
}
