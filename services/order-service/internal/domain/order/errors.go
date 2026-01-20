package order

import "errors"

var (
	// Order errors
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderAlreadyPaid      = errors.New("order already paid")
	ErrOrderAlreadyCancelled = errors.New("order already cancelled")
	ErrOrderExpired          = errors.New("order has expired")
	ErrInvalidOrderStatus    = errors.New("invalid order status")
	ErrInvalidQuantity       = errors.New("quantity must be positive")

	// Payment errors
	ErrPaymentFailed               = errors.New("payment failed")
	ErrPaymentAlreadyCompleted     = errors.New("payment already completed")
	ErrPaymentAlreadyFailed        = errors.New("payment already failed")
	ErrCannotCompleteFailedPayment = errors.New("cannot complete a failed payment")
	ErrCannotFailCompletedPayment  = errors.New("cannot fail a completed payment")
	ErrInvalidPaymentMethod        = errors.New("invalid payment method")
	ErrInsufficientFunds           = errors.New("insufficient funds")

	// Value object errors
	ErrEmptyOrderID         = errors.New("order id cannot be empty")
	ErrInvalidOrderIDFormat = errors.New("invalid order id format")
	ErrEmptyReservationID   = errors.New("reservation id cannot be empty")
	ErrEmptyUserID          = errors.New("user id cannot be empty")
	ErrEmptyProductID       = errors.New("product id cannot be empty")
	ErrEmptyPaymentID       = errors.New("payment id cannot be empty")
	ErrEmptyCurrency        = errors.New("currency is required")
	ErrNegativeAmount       = errors.New("amount cannot be negative")
)
