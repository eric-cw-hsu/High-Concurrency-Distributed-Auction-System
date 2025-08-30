package mocks

import (
	"context"
	"sync"
)

type MockStockRepository struct {
	mu     sync.RWMutex
	stocks map[string]int
	prices map[string]float64
}

func NewMockStockRepository() *MockStockRepository {
	return &MockStockRepository{
		stocks: make(map[string]int),
		prices: make(map[string]float64),
	}
}

func (m *MockStockRepository) SetInitialStock(stockId string, quantity int, price float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stocks[stockId] = quantity
	m.prices[stockId] = price
}

func (m *MockStockRepository) GetCurrentStock(stockId string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stocks[stockId]
}

func (m *MockStockRepository) GetPrice(stockId string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.prices[stockId]
}

func (m *MockStockRepository) RemoveAll(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stocks = make(map[string]int)
	m.prices = make(map[string]float64)
}
