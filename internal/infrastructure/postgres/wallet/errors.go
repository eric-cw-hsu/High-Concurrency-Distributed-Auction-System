package wallet

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
)

// WrapRepositoryError wraps infrastructure errors into domain repository errors
func WrapRepositoryError(operation, userId string, err error) error {
	return &wallet.RepositoryError{
		Operation: operation,
		UserId:    userId,
		Wrapped:   err,
	}
}

// WrapWalletNotFoundError wraps database not found errors
func WrapWalletNotFoundError(userId string) error {
	return &wallet.WalletNotFoundError{
		UserId: userId,
	}
}

// WrapInsufficientBalanceError wraps balance check errors
func WrapInsufficientBalanceError(userId string, current, required float64) error {
	return &wallet.InsufficientBalanceError{
		UserId:    userId,
		Current:   current,
		Required:  required,
		Operation: "repository_balance_check",
	}
}
