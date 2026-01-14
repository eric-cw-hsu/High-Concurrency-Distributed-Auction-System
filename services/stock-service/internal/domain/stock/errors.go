package stock

import "errors"

var (
	ErrStockNotFound      = errors.New("stock not found")
	ErrInvalidProductID   = errors.New("invalid product id")
	ErrProductIDRequired  = errors.New("product id is required")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrInvalidQuantity    = errors.New("invalid quantity")
	ErrNegativeQuantity   = errors.New("quantity cannot be negative")
	ErrExceedsMaxQuantity = errors.New("quantity exceeds maximum limit of 10")
)
