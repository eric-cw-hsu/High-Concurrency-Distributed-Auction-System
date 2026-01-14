package clients

import (
	"context"

	stockv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/stock/v1"
	"google.golang.org/grpc"
)

type StockClient struct {
	cli stockv1.StockServiceClient
}

func NewStockClient(conn *grpc.ClientConn) *StockClient {
	return &StockClient{cli: stockv1.NewStockServiceClient(conn)}
}

func (c *StockClient) SetStock(ctx context.Context, productID string, quantity int32) (*stockv1.SetStockResponse, error) {
	return c.cli.SetStock(ctx, &stockv1.SetStockRequest{
		ProductId: productID,
		Quantity:  quantity,
	})
}

func (c *StockClient) GetStock(ctx context.Context, productID string) (*stockv1.GetStockResponse, error) {
	return c.cli.GetStock(ctx, &stockv1.GetStockRequest{
		ProductId: productID,
	})
}

func (c *StockClient) Reserve(ctx context.Context, productID, userID string, quantity int32) (*stockv1.ReserveResponse, error) {
	return c.cli.Reserve(ctx, &stockv1.ReserveRequest{
		ProductId: productID,
		UserId:    userID,
		Quantity:  quantity,
	})
}

func (c *StockClient) Release(ctx context.Context, reservationID string) (*stockv1.ReleaseResponse, error) {
	return c.cli.Release(ctx, &stockv1.ReleaseRequest{
		ReservationId: reservationID,
	})
}

func (c *StockClient) GetReservation(ctx context.Context, reservationID string) (*stockv1.GetReservationResponse, error) {
	return c.cli.GetReservation(ctx, &stockv1.GetReservationRequest{
		ReservationId: reservationID,
	})
}

func (c *StockClient) TriggerRecovery(ctx context.Context, recoveryType string) (*stockv1.TriggerRecoveryResponse, error) {
	return c.cli.TriggerRecovery(ctx, &stockv1.TriggerRecoveryRequest{
		RecoveryType: recoveryType,
	})
}
