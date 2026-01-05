package grpc

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapDomainErrorToGRPC maps domain errors to gRPC status codes
func mapDomainErrorToGRPC(err error) error {
	msg := err.Error()

	// User already exists
	if strings.Contains(msg, "already exists") {
		return status.Error(codes.AlreadyExists, "user already exists")
	}

	// Invalid credentials
	if strings.Contains(msg, "invalid credentials") || strings.Contains(msg, "invalid password") {
		return status.Error(codes.Unauthenticated, "invalid email or password")
	}

	// User not found
	if strings.Contains(msg, "not found") {
		return status.Error(codes.NotFound, "user not found")
	}

	// Invalid token
	if strings.Contains(msg, "invalid token") || strings.Contains(msg, "token") {
		return status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	// Validation errors
	if isValidationError(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Default: internal error
	return status.Error(codes.Internal, "internal server error")
}

func isValidationError(err error) bool {
	msg := strings.ToLower(err.Error())
	validationKeywords := []string{
		"invalid",
		"required",
		"must be",
		"cannot be empty",
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

func isBusinessError(code codes.Code) bool {
	switch code {
	case codes.AlreadyExists,
		codes.PermissionDenied,
		codes.FailedPrecondition,
		codes.Aborted,
		codes.ResourceExhausted:
		return true
	default:
		return false
	}
}
