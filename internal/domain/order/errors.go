package order

import (
	"errors"
	"fmt"
)

// Domain validation errors
var (
	ErrEmptyBuyerID      = errors.New("buyer ID cannot be empty")
	ErrEmptyStockID      = errors.New("stock ID cannot be empty")
	ErrInvalidQuantity   = errors.New("quantity must be positive")
	ErrOrderNotConfirmed = errors.New("order cannot be confirmed")
	ErrOrderNotCancelled = errors.New("order cannot be cancelled")
)

// Business logic errors
var (
	ErrInsufficientStock = errors.New("insufficient stock available")
	ErrOutOfStock        = errors.New("out of stock")
)

// Infrastructure errors
var (
	ErrStockPriceUnavailable    = errors.New("stock price unavailable")
	ErrStockQuantityUnavailable = errors.New("stock quantity unavailable")
	ErrPaymentProcessingFailed  = errors.New("payment processing failed")
	ErrStockUpdateFailed        = errors.New("stock update failed")
)

// InsufficientStockError provides detailed information about stock shortage
type InsufficientStockError struct {
	StockID   string
	Available int
	Requested int
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("insufficient stock for %s: available %d, requested %d",
		e.StockID, e.Available, e.Requested)
}

func (e *InsufficientStockError) Is(target error) bool {
	return target == ErrInsufficientStock
}

// PaymentError provides detailed information about payment failures
type PaymentError struct {
	UserId  string
	Amount  float64
	Reason  string
	Wrapped error
}

func (e *PaymentError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("payment failed for user %s (amount: %.2f): %s: %v",
			e.UserId, e.Amount, e.Reason, e.Wrapped)
	}
	return fmt.Sprintf("payment failed for user %s (amount: %.2f): %s",
		e.UserId, e.Amount, e.Reason)
}

func (e *PaymentError) Is(target error) bool {
	return target == ErrPaymentProcessingFailed
}

func (e *PaymentError) Unwrap() error {
	return e.Wrapped
}

// StockError provides detailed information about stock-related failures
type StockError struct {
	StockID   string
	Operation string
	Wrapped   error
}

func (e *StockError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("stock operation failed for %s (%s): %v",
			e.StockID, e.Operation, e.Wrapped)
	}
	return fmt.Sprintf("stock operation failed for %s (%s)",
		e.StockID, e.Operation)
}

func (e *StockError) Is(target error) bool {
	switch e.Operation {
	case "get_price":
		return target == ErrStockPriceUnavailable
	case "get_quantity":
		return target == ErrStockQuantityUnavailable
	case "update":
		return target == ErrStockUpdateFailed
	default:
		return false
	}
}

func (e *StockError) Unwrap() error {
	return e.Wrapped
}

// RepositoryError wraps infrastructure repository errors
type RepositoryError struct {
	Operation string
	OrderId   string
	Wrapped   error
}

func (e *RepositoryError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("repository operation '%s' failed for order %s: %v",
			e.Operation, e.OrderId, e.Wrapped)
	}
	return fmt.Sprintf("repository operation '%s' failed for order %s",
		e.Operation, e.OrderId)
}

func (e *RepositoryError) Unwrap() error {
	return e.Wrapped
}
