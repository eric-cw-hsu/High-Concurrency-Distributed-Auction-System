package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
)

// MockEventPublisher is a mock implementation of wallet.EventPublisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}
