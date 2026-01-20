package clients

import (
	"context"

	"google.golang.org/grpc"

	pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/order/v1"
)

// OrderClient wraps Order Service gRPC client for synchronous queries
type OrderClient struct {
	client pb.OrderServiceClient
}

// NewOrderClient creates a new OrderClient instance
func NewOrderClient(conn *grpc.ClientConn) *OrderClient {
	return &OrderClient{
		client: pb.NewOrderServiceClient(conn),
	}
}

// GetOrder fetches a single order by ID from the Order Service
func (c *OrderClient) GetOrder(
	ctx context.Context,
	req *pb.GetOrderRequest,
) (*pb.OrderResponse, error) {
	return c.client.GetOrder(ctx, req)
}

// ListUserOrders fetches a paginated list of orders for a specific user
func (c *OrderClient) ListUserOrders(
	ctx context.Context,
	req *pb.ListUserOrdersRequest,
) (*pb.ListUserOrdersResponse, error) {
	return c.client.ListUserOrders(ctx, req)
}
