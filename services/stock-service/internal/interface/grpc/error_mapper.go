package grpc

import (
	"errors"
	"strings"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/stock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapDomainErrorToGRPC maps domain errors to gRPC status codes
func mapDomainErrorToGRPC(err error) error {
	// Stock errors
	if errors.Is(err, stock.ErrStockNotFound) {
		return status.Error(codes.NotFound, "stock not found")
	}
	if errors.Is(err, stock.ErrInsufficientStock) {
		return status.Error(codes.FailedPrecondition, "insufficient stock")
	}
	if errors.Is(err, stock.ErrInvalidProductID) {
		return status.Error(codes.InvalidArgument, "invalid product id")
	}
	if errors.Is(err, stock.ErrExceedsMaxQuantity) {
		return status.Error(codes.InvalidArgument, "quantity exceeds maximum limit of 10")
	}
	if errors.Is(err, stock.ErrInvalidQuantity) {
		return status.Error(codes.InvalidArgument, "invalid quantity")
	}

	// Reservation errors
	if errors.Is(err, reservation.ErrReservationNotFound) {
		return status.Error(codes.NotFound, "reservation not found")
	}
	if errors.Is(err, reservation.ErrReservationExpired) {
		return status.Error(codes.FailedPrecondition, "reservation has expired")
	}
	if errors.Is(err, reservation.ErrInvalidReservationID) {
		return status.Error(codes.InvalidArgument, "invalid reservation id")
	}
	if errors.Is(err, reservation.ErrExceedsMaxQuantity) {
		return status.Error(codes.InvalidArgument, "quantity exceeds maximum limit of 10")
	}
	if errors.Is(err, reservation.ErrCanOnlyConsumeReserved) {
		return status.Error(codes.FailedPrecondition, "only reserved reservations can be consumed")
	}
	if errors.Is(err, reservation.ErrCanOnlyReleaseReserved) {
		return status.Error(codes.FailedPrecondition, "only reserved reservations can be released")
	}

	// Validation errors
	if isValidationError(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Default: internal error
	return status.Error(codes.Internal, "internal server error")
}

// isValidationError checks if error is a validation error
func isValidationError(err error) bool {
	msg := strings.ToLower(err.Error())
	validationKeywords := []string{
		"invalid",
		"required",
		"must be",
		"cannot be",
		"too short",
		"too long",
	}

	for _, keyword := range validationKeywords {
		if strings.Contains(msg, keyword) {
			return true
		}
	}
	return false
}

// isSystemError checks if error code represents a system error
func isSystemError(code codes.Code) bool {
	switch code {
	case codes.Internal,
		codes.Unavailable,
		codes.DataLoss,
		codes.Unknown:
		return true
	default:
		return false
	}
}

// isBusinessError checks if error code represents a business error
func isBusinessError(code codes.Code) bool {
	switch code {
	case codes.FailedPrecondition,
		codes.ResourceExhausted,
		codes.AlreadyExists,
		codes.Aborted:
		return true
	default:
		return false
	}
}
