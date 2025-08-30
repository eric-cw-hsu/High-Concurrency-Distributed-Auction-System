package order

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
)

// WrapRepositoryError wraps infrastructure errors into domain repository errors
func WrapRepositoryError(operation, orderId string, err error) error {
	return &order.RepositoryError{
		Operation: operation,
		OrderId:   orderId,
		Wrapped:   err,
	}
}
