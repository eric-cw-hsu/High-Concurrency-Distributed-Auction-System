package order

import "context"

type OrderRepository interface {
	SaveOrder(ctx context.Context, event OrderPlacedEvent) error
}
