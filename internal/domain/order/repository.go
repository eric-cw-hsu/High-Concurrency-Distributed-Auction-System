package order

import "context"

type OrderRepository interface {
	SaveOrder(ctx context.Context, order *Order) error
}
