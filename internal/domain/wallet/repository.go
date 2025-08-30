package wallet

import (
	"context"
)

type WalletRepository interface {
	// Aggregate operations
	Save(ctx context.Context, aggregate *WalletAggregate) error
	GetByUserId(ctx context.Context, userId string) (*WalletAggregate, error)
	CreateWallet(ctx context.Context, userId string) (*WalletAggregate, error)
}
