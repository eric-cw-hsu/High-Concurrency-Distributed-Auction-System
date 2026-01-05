package grpc

import (
	"errors"
	"strings"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/domain/product"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapDomainErrorToGRPC maps domain errors to gRPC status codes
func mapDomainErrorToGRPC(err error) error {
	// Product not found
	if errors.Is(err, product.ErrProductNotFound) {
		return status.Error(codes.NotFound, "product not found")
	}

	// Invalid input
	if errors.Is(err, product.ErrInvalidProductID) {
		return status.Error(codes.InvalidArgument, "invalid product id format")
	}

	// Business rule violations (FailedPrecondition)
	if errors.Is(err, product.ErrCannotPublishProduct) {
		return status.Error(codes.FailedPrecondition,
			"only draft or inactive products can be published")
	}
	if errors.Is(err, product.ErrCannotDeactivateProduct) {
		return status.Error(codes.FailedPrecondition,
			"only active products can be deactivated")
	}
	if errors.Is(err, product.ErrCannotMarkAsSoldOut) {
		return status.Error(codes.FailedPrecondition,
			"only active products can be marked as sold out")
	}
	if errors.Is(err, product.ErrCannotUpdateActiveProduct) {
		return status.Error(codes.FailedPrecondition,
			"cannot update active product information")
	}
	if errors.Is(err, product.ErrCannotUpdatePricingForActiveProduct) {
		return status.Error(codes.FailedPrecondition,
			"cannot update pricing for active product")
	}

	// Authorization errors
	if errors.Is(err, product.ErrUnauthorizedDelete) {
		return status.Error(codes.PermissionDenied,
			"you are not authorized to delete this product")
	}

	// Validation errors (check error message)
	if isValidationError(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Default: internal error
	return status.Error(codes.Internal, "internal server error")
}

// isValidationError checks if error is a validation error based on message
func isValidationError(err error) bool {
	msg := strings.ToLower(err.Error())
	validationKeywords := []string{
		"invalid",
		"cannot be empty",
		"too long",
		"too short",
		"must be",
		"required",
		"greater than",
		"less than",
	}

	for _, keyword := range validationKeywords {
		if strings.Contains(msg, keyword) {
			return true
		}
	}
	return false
}

// isDomainError checks if error is a domain/business error (not system error)
func isDomainError(err error) bool {
	// Domain errors
	domainErrors := []error{
		product.ErrProductNotFound,
		product.ErrInvalidProductID,
		product.ErrCannotPublishProduct,
		product.ErrCannotDeactivateProduct,
		product.ErrCannotMarkAsSoldOut,
		product.ErrCannotUpdateActiveProduct,
		product.ErrCannotUpdatePricingForActiveProduct,
		product.ErrUnauthorizedDelete,
	}

	for _, domainErr := range domainErrors {
		if errors.Is(err, domainErr) {
			return true
		}
	}

	return isValidationError(err)
}

// isSystemError checks if gRPC code represents a system error (needs immediate attention)
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

// isBusinessError checks if gRPC code represents a business rule violation (expected)
func isBusinessError(code codes.Code) bool {
	switch code {
	case codes.FailedPrecondition,
		codes.PermissionDenied,
		codes.AlreadyExists,
		codes.Aborted,
		codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

// isClientError checks if gRPC code represents a client error (user's fault)
func isClientError(code codes.Code) bool {
	switch code {
	case codes.InvalidArgument,
		codes.NotFound,
		codes.Unauthenticated,
		codes.OutOfRange,
		codes.Canceled,
		codes.DeadlineExceeded:
		return true
	default:
		return false
	}
}
