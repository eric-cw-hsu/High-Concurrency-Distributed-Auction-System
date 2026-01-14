package reservation

import "context"

type CacheRepository interface {
	Save(ctx context.Context, res *Reservation) error
	FindByID(ctx context.Context, id ReservationID) (*Reservation, error)
	Delete(ctx context.Context, id ReservationID) error
	FindActiveByProductID(ctx context.Context, productID ProductID) ([]*Reservation, error)
}
