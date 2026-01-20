package order

import "context"

// OrderCreator defines the interface for creating orders
type Creator interface {
	CreateOrderFromReservation(ctx context.Context, reservationID, userID, productID string, quantity int) error
}

type Service interface {
	CreateOrder(ctx context.Context, reservationID string, userID string, productID string, quantity int, unitPrice int64, currency string) (*Order, error)
	CancelExpiredOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	ListUserOrders(ctx context.Context, userID string, limit, offset int) ([]*Order, error)
}
