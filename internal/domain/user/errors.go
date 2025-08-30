package user

import (
	"errors"
	"fmt"
)

// Domain validation errors (already defined in register_user_command.go)
// Re-exported here for consistency
var (
	ErrInvalidEmail    = errors.New("email is invalid")
	ErrInvalidName     = errors.New("name is invalid")
	ErrInvalidPassword = errors.New("password is invalid")
)

// Business logic errors
var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Infrastructure errors
var (
	ErrUserRepositoryFailure = errors.New("user repository operation failed")
	ErrPasswordHashingFailed = errors.New("password hashing failed")
)

// UserNotFoundError provides detailed information about user lookup failures
type UserNotFoundError struct {
	Email  string
	UserId string
}

func (e *UserNotFoundError) Error() string {
	if e.Email != "" {
		return fmt.Sprintf("user not found with email: %s", e.Email)
	}
	if e.UserId != "" {
		return fmt.Sprintf("user not found with ID: %s", e.UserId)
	}
	return "user not found"
}

func (e *UserNotFoundError) Is(target error) bool {
	return target == ErrUserNotFound
}

// UserAlreadyExistsError provides detailed information about duplicate user registration
type UserAlreadyExistsError struct {
	Email string
}

func (e *UserAlreadyExistsError) Error() string {
	return fmt.Sprintf("user already exists with email: %s", e.Email)
}

func (e *UserAlreadyExistsError) Is(target error) bool {
	return target == ErrUserAlreadyExists
}

// RepositoryError wraps repository operation failures
type RepositoryError struct {
	Operation string
	Wrapped   error
}

func (e *RepositoryError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("user repository operation failed (%s): %v", e.Operation, e.Wrapped)
	}
	return fmt.Sprintf("user repository operation failed (%s)", e.Operation)
}

func (e *RepositoryError) Is(target error) bool {
	return target == ErrUserRepositoryFailure
}

func (e *RepositoryError) Unwrap() error {
	return e.Wrapped
}
