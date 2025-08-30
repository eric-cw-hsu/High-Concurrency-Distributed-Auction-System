package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
)

// MockWalletRepository is a mock implementation of wallet.WalletRepository
type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) Save(ctx context.Context, aggregate *wallet.WalletAggregate) error {
	args := m.Called(ctx, aggregate)
	return args.Error(0)
}

func (m *MockWalletRepository) GetByUserId(ctx context.Context, userId string) (*wallet.WalletAggregate, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.WalletAggregate), args.Error(1)
}

func (m *MockWalletRepository) CreateWallet(ctx context.Context, userId string) (*wallet.WalletAggregate, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.WalletAggregate), args.Error(1)
}
