package mocks

import (
	"context"

	walletUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/wallet"
	"github.com/stretchr/testify/mock"
)

// TestifyMockStockCache provides a testify-based mock for E2E tests
type TestifyMockStockCache struct {
	mock.Mock
}

func NewTestifyMockStockCache() *TestifyMockStockCache {
	return &TestifyMockStockCache{}
}

func (m *TestifyMockStockCache) DecreaseStock(ctx context.Context, stockId string, quantity int) (int64, error) {
	args := m.Called(ctx, stockId, quantity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *TestifyMockStockCache) RestoreStock(ctx context.Context, stockId string, quantity int) error {
	args := m.Called(ctx, stockId, quantity)
	return args.Error(0)
}

func (m *TestifyMockStockCache) GetPrice(ctx context.Context, stockId string) (float64, error) {
	args := m.Called(ctx, stockId)
	return args.Get(0).(float64), args.Error(1)
}

func (m *TestifyMockStockCache) GetStock(ctx context.Context, stockId string) (int, error) {
	args := m.Called(ctx, stockId)
	return args.Get(0).(int), args.Error(1)
}

func (m *TestifyMockStockCache) SetStock(ctx context.Context, stockId string, quantity int) error {
	args := m.Called(ctx, stockId, quantity)
	return args.Error(0)
}

func (m *TestifyMockStockCache) SetPrice(ctx context.Context, stockId string, price float64) error {
	args := m.Called(ctx, stockId, price)
	return args.Error(0)
}

func (m *TestifyMockStockCache) RemoveAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestifyMockWalletService provides a testify-based mock for E2E tests
type TestifyMockWalletService struct {
	mock.Mock
}

func NewTestifyMockWalletService() *TestifyMockWalletService {
	return &TestifyMockWalletService{}
}

func (m *TestifyMockWalletService) EnsureWalletExists(ctx context.Context, userId string) (*walletUsecase.WalletInfo, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletUsecase.WalletInfo), args.Error(1)
}

func (m *TestifyMockWalletService) CreateWallet(ctx context.Context, userId string) (*walletUsecase.WalletInfo, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletUsecase.WalletInfo), args.Error(1)
}

func (m *TestifyMockWalletService) ProcessPaymentWithSufficientFunds(ctx context.Context, userId, orderId string, amount float64) (*walletUsecase.WalletInfo, error) {
	args := m.Called(ctx, userId, orderId, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletUsecase.WalletInfo), args.Error(1)
}

func (m *TestifyMockWalletService) ProcessRefundSafely(ctx context.Context, userId, orderId string, amount float64) (*walletUsecase.WalletInfo, error) {
	args := m.Called(ctx, userId, orderId, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletUsecase.WalletInfo), args.Error(1)
}
