package reservation

import (
	"context"
	"time"
)

type PersistentRepository interface {
	Save(ctx context.Context, res *Reservation) error
	FindByID(ctx context.Context, id ReservationID) (*Reservation, error)
	UpdateStatus(ctx context.Context, id ReservationID, status ReservationStatus) error
	FindAllActive(ctx context.Context) ([]*Reservation, error)
	FindActiveByProductID(ctx context.Context, productID ProductID) ([]*Reservation, error)

	FindExpiredWithinWindow(ctx context.Context, windowStart, windowEnd time.Time, limit int) ([]*Reservation, error)
}
