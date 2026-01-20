package order

import (
	"context"
	"time"
)

// Repository defines the interface for order persistence
type Repository interface {
	// Save saves an order
	Save(ctx context.Context, order *Order) error

	// FindByID finds an order by ID
	FindByID(ctx context.Context, id OrderID) (*Order, error)

	// FindByReservationID finds an order by reservation ID
	FindByReservationID(ctx context.Context, reservationID ReservationID) (*Order, error)

	// FindByUserID finds orders by user ID
	FindByUserID(ctx context.Context, userID UserID, limit, offset int) ([]*Order, error)

	// FindExpired finds expired orders
	FindExpired(ctx context.Context, now time.Time, limit int) ([]*Order, error)

	// UpdateStatus updates order status
	UpdateStatus(ctx context.Context, id OrderID, status OrderStatus) error
}
