package wallet

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
)

// WrapRepositoryError wraps infrastructure errors into domain repository errors
func WrapRepositoryError(operation, userID string, err error) error {
	return &wallet.RepositoryError{
		Operation: operation,
		UserID:    userID,
		Wrapped:   err,
	}
}

// WrapWalletNotFoundError wraps database not found errors
func WrapWalletNotFoundError(userID string) error {
	return &wallet.WalletNotFoundError{
		UserID: userID,
	}
}

// WrapInsufficientBalanceError wraps balance check errors
func WrapInsufficientBalanceError(userID string, current, required float64) error {
	return &wallet.InsufficientBalanceError{
		UserID:    userID,
		Current:   current,
		Required:  required,
		Operation: "repository_balance_check",
	}
}
