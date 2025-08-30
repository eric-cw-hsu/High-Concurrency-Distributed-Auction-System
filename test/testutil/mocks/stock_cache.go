package mocks

import (
	"context"
	"fmt"
	"sync"
)

// MockStockCache provides a thread-safe mock implementation for stock operations
type MockStockCache struct {
	mu     sync.RWMutex
	stocks map[string]int
	prices map[string]float64
}

func NewMockStockCache() *MockStockCache {
	return &MockStockCache{
		stocks: make(map[string]int),
		prices: make(map[string]float64),
	}
}

func (m *MockStockCache) SetInitialStock(stockId string, quantity int, price float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stocks[stockId] = quantity
	m.prices[stockId] = price
}

func (m *MockStockCache) GetCurrentStock(stockId string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stocks[stockId]
}

func (m *MockStockCache) DecreaseStock(ctx context.Context, stockId string, quantity int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentStock, exists := m.stocks[stockId]
	if !exists {
		return 0, fmt.Errorf("stock not found: %s", stockId)
	}

	if currentStock < quantity {
		return 0, fmt.Errorf("insufficient stock: have %d, need %d", currentStock, quantity)
	}

	m.stocks[stockId] = currentStock - quantity
	return int64(m.stocks[stockId]), nil
}

func (m *MockStockCache) RestoreStock(ctx context.Context, stockId string, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentStock := m.stocks[stockId]
	m.stocks[stockId] = currentStock + quantity
	return nil
}

func (m *MockStockCache) GetPrice(ctx context.Context, stockId string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	price, exists := m.prices[stockId]
	if !exists {
		return 0, fmt.Errorf("price not found for stock: %s", stockId)
	}
	return price, nil
}

func (m *MockStockCache) GetStock(ctx context.Context, stockId string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stock, exists := m.stocks[stockId]
	if !exists {
		return 0, fmt.Errorf("stock not found: %s", stockId)
	}
	return stock, nil
}

func (m *MockStockCache) SetStock(ctx context.Context, stockId string, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stocks[stockId] = quantity
	return nil
}

func (m *MockStockCache) SetPrice(ctx context.Context, stockId string, price float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prices[stockId] = price
	return nil
}

func (m *MockStockCache) RemoveAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stocks = make(map[string]int)
	m.prices = make(map[string]float64)
	return nil
}
