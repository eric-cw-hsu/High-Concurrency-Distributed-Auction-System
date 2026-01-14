package reservation

import "errors"

var (
	ErrReservationNotFound    = errors.New("reservation not found")
	ErrInvalidReservationID   = errors.New("invalid reservation id")
	ErrInvalidProductID       = errors.New("invalid product id")
	ErrProductIDRequired      = errors.New("product id is required")
	ErrInvalidUserID          = errors.New("invalid user id")
	ErrUserIDRequired         = errors.New("user id is required")
	ErrReservationExpired     = errors.New("reservation has expired")
	ErrExceedsMaxQuantity     = errors.New("quantity exceeds maximum limit of 10")
	ErrInvalidQuantity        = errors.New("invalid quantity")
	ErrCanOnlyConsumeReserved = errors.New("only reserved reservations can be consumed")
	ErrCanOnlyReleaseReserved = errors.New("only reserved reservations can be released")
	ErrCanOnlyExpireReserved  = errors.New("only reserved reservations can expire")
)
