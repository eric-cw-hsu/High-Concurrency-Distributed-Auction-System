package mocks

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
)

// MockWalletEventPublisher provides a simple mock for wallet event publishing
type MockWalletEventPublisher struct {
	mu           sync.RWMutex
	publishCount int64
	shouldFail   bool
	events       []domain.DomainEvent
}

func NewMockWalletEventPublisher() *MockWalletEventPublisher {
	return &MockWalletEventPublisher{
		events: make([]domain.DomainEvent, 0),
	}
}

func (m *MockWalletEventPublisher) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
}

func (m *MockWalletEventPublisher) GetPublishCount() int64 {
	return atomic.LoadInt64(&m.publishCount)
}

func (m *MockWalletEventPublisher) GetEvents() []domain.DomainEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	events := make([]domain.DomainEvent, len(m.events))
	copy(events, m.events)
	return events
}

func (m *MockWalletEventPublisher) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	atomic.StoreInt64(&m.publishCount, 0)
	m.shouldFail = false
	m.events = make([]domain.DomainEvent, 0)
}

func (m *MockWalletEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return fmt.Errorf("mock wallet event publisher configured to fail")
	}

	m.events = append(m.events, event)
	atomic.AddInt64(&m.publishCount, 1)
	return nil
}
