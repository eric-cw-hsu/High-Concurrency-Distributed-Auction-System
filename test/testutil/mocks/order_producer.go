package mocks

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
)

// MockOrderProducer provides a thread-safe mock implementation for order event publishing
type MockOrderProducer struct {
	mu           sync.RWMutex
	publishCount int64
	shouldFail   bool
	events       []domain.DomainEvent
	messages     []order.OrderPlacedEvent
}

func NewMockOrderProducer() *MockOrderProducer {
	return &MockOrderProducer{
		events:   make([]domain.DomainEvent, 0),
		messages: make([]order.OrderPlacedEvent, 0),
	}
}

func (m *MockOrderProducer) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
}

func (m *MockOrderProducer) GetPublishCount() int64 {
	return atomic.LoadInt64(&m.publishCount)
}

func (m *MockOrderProducer) GetEvents() []domain.DomainEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	events := make([]domain.DomainEvent, len(m.events))
	copy(events, m.events)
	return events
}

func (m *MockOrderProducer) GetMessages() []order.OrderPlacedEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	messages := make([]order.OrderPlacedEvent, len(m.messages))
	copy(messages, m.messages)
	return messages
}

func (m *MockOrderProducer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	atomic.StoreInt64(&m.publishCount, 0)
	m.shouldFail = false
	m.events = make([]domain.DomainEvent, 0)
	m.messages = make([]order.OrderPlacedEvent, 0)
}

func (m *MockOrderProducer) PublishOrder(ctx context.Context, event domain.DomainEvent) error {
	return m.PublishEvent(ctx, event)
}

func (m *MockOrderProducer) PublishEvent(ctx context.Context, event domain.DomainEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return fmt.Errorf("mock order producer configured to fail")
	}

	m.events = append(m.events, event)
	if orderEvent, ok := event.(*order.OrderPlacedEvent); ok {
		m.messages = append(m.messages, *orderEvent)
	}
	atomic.AddInt64(&m.publishCount, 1)
	return nil
}

func (m *MockOrderProducer) PublishOrderEvent(event order.OrderPlacedEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return fmt.Errorf("mock order producer configured to fail")
	}

	m.messages = append(m.messages, event)
	atomic.AddInt64(&m.publishCount, 1)
	return nil
}
