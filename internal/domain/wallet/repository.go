package wallet

import (
	"context"
)

type WalletRepository interface {
	// Aggregate operations
	Save(ctx context.Context, aggregate *WalletAggregate) error
	GetByUserID(ctx context.Context, userID string) (*WalletAggregate, error)
	CreateWallet(ctx context.Context, userID string) (*WalletAggregate, error)
}
