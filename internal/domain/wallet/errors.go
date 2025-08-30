package wallet

import (
	"errors"
	"fmt"
)

var (
	// ErrWalletNotFound is returned when a wallet is not found
	ErrWalletNotFound = errors.New("wallet not found")

	// ErrWalletAlreadyExists is returned when trying to create a wallet that already exists
	ErrWalletAlreadyExists = errors.New("wallet already exists")

	// ErrInsufficientBalance is returned when there's not enough balance for an operation
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrWalletSuspended is returned when trying to operate on a suspended wallet
	ErrWalletSuspended = errors.New("wallet is suspended")

	// ErrWalletInactive is returned when trying to operate on an inactive wallet
	ErrWalletInactive = errors.New("wallet is inactive")

	// ErrInvalidAmount is returned when an invalid amount is provided
	ErrInvalidAmount = errors.New("invalid amount")
)

// Infrastructure errors
var (
	ErrWalletRepositoryFailure = errors.New("wallet repository operation failed")
	ErrWalletSaveFailure       = errors.New("wallet save operation failed")
)

// InsufficientBalanceError provides detailed information about balance shortage
type InsufficientBalanceError struct {
	UserId    string
	Current   float64
	Required  float64
	Operation string
}

func (e *InsufficientBalanceError) Error() string {
	return fmt.Sprintf("insufficient balance for user %s (%s): current=%.2f, required=%.2f",
		e.UserId, e.Operation, e.Current, e.Required)
}

func (e *InsufficientBalanceError) Is(target error) bool {
	return target == ErrInsufficientBalance
}

// WalletNotFoundError provides detailed information about wallet lookup failures
type WalletNotFoundError struct {
	UserId string
}

func (e *WalletNotFoundError) Error() string {
	return fmt.Sprintf("wallet not found for user: %s", e.UserId)
}

func (e *WalletNotFoundError) Is(target error) bool {
	return target == ErrWalletNotFound
}

// WalletAlreadyExistsError provides detailed information about duplicate wallet creation
type WalletAlreadyExistsError struct {
	UserId string
}

func (e *WalletAlreadyExistsError) Error() string {
	return fmt.Sprintf("wallet already exists for user: %s", e.UserId)
}

func (e *WalletAlreadyExistsError) Is(target error) bool {
	return target == ErrWalletAlreadyExists
}

// WalletStatusError provides detailed information about wallet status issues
type WalletStatusError struct {
	UserId    string
	Status    int
	Operation string
}

func (e *WalletStatusError) Error() string {
	statusStr := "unknown"
	switch e.Status {
	case 0:
		statusStr = "inactive"
	case 1:
		statusStr = "active"
	case 2:
		statusStr = "suspended"
	}
	return fmt.Sprintf("wallet operation failed for user %s (%s): wallet status is %s",
		e.UserId, e.Operation, statusStr)
}

func (e *WalletStatusError) Is(target error) bool {
	switch e.Status {
	case 0:
		return target == ErrWalletInactive
	case 2:
		return target == ErrWalletSuspended
	default:
		return false
	}
}

// InvalidAmountError provides detailed information about invalid amounts
type InvalidAmountError struct {
	Amount    float64
	Operation string
}

func (e *InvalidAmountError) Error() string {
	return fmt.Sprintf("invalid amount for %s: %.2f", e.Operation, e.Amount)
}

func (e *InvalidAmountError) Is(target error) bool {
	return target == ErrInvalidAmount
}

// RepositoryError wraps repository operation failures
type RepositoryError struct {
	Operation string
	UserId    string
	Wrapped   error
}

func (e *RepositoryError) Error() string {
	if e.UserId != "" {
		if e.Wrapped != nil {
			return fmt.Sprintf("wallet repository operation failed (%s) for user %s: %v",
				e.Operation, e.UserId, e.Wrapped)
		}
		return fmt.Sprintf("wallet repository operation failed (%s) for user %s",
			e.Operation, e.UserId)
	}

	if e.Wrapped != nil {
		return fmt.Sprintf("wallet repository operation failed (%s): %v", e.Operation, e.Wrapped)
	}
	return fmt.Sprintf("wallet repository operation failed (%s)", e.Operation)
}

func (e *RepositoryError) Is(target error) bool {
	return target == ErrWalletRepositoryFailure || target == ErrWalletSaveFailure
}

func (e *RepositoryError) Unwrap() error {
	return e.Wrapped
}
