package mocks

import (
	"context"
	"fmt"
	"sync"

	walletUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/wallet"
)

// MockWalletService provides a thread-safe mock implementation for wallet operations
type MockWalletService struct {
	mu       sync.RWMutex
	balances map[string]float64
}

func NewMockWalletService() *MockWalletService {
	return &MockWalletService{
		balances: make(map[string]float64),
	}
}

func (m *MockWalletService) SetBalance(userId string, balance float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.balances[userId] = balance
}

func (m *MockWalletService) GetBalance(userId string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.balances[userId]
}

func (m *MockWalletService) EnsureWalletExists(ctx context.Context, userId string) (*walletUsecase.WalletInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	balance, exists := m.balances[userId]
	if !exists {
		m.balances[userId] = 0.0
		balance = 0.0
	}

	return &walletUsecase.WalletInfo{
		Balance: balance,
	}, nil
}

func (m *MockWalletService) CreateWallet(ctx context.Context, userId string) (*walletUsecase.WalletInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.balances[userId] = 0.0
	return &walletUsecase.WalletInfo{
		Balance: 0.0,
	}, nil
}

func (m *MockWalletService) ProcessPaymentWithSufficientFunds(ctx context.Context, userId, orderId string, amount float64) (*walletUsecase.WalletInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	balance, exists := m.balances[userId]
	if !exists {
		return nil, fmt.Errorf("wallet not found for user: %s", userId)
	}

	if balance < amount {
		return nil, fmt.Errorf("insufficient funds: have %.2f, need %.2f", balance, amount)
	}

	m.balances[userId] = balance - amount
	return &walletUsecase.WalletInfo{
		Balance: m.balances[userId],
	}, nil
}

func (m *MockWalletService) ProcessRefundSafely(ctx context.Context, userId, orderId string, amount float64) (*walletUsecase.WalletInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	balance := m.balances[userId]
	m.balances[userId] = balance + amount
	return &walletUsecase.WalletInfo{
		Balance: m.balances[userId],
	}, nil
}
