package testutil

import (
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil/mocks"
)

// Export mock types for backward compatibility
type MockOrderProducer = mocks.MockOrderProducer
type MockWalletService = mocks.MockWalletService
type MockStockCache = mocks.MockStockCache
type MockWalletEventPublisher = mocks.MockWalletEventPublisher
type MockStockRepository = mocks.MockStockRepository
type MockStockEventConsumer = mocks.MockStockEventConsumer

// Export factory functions
var (
	NewMockOrderProducer        = mocks.NewMockOrderProducer
	NewMockWalletService        = mocks.NewMockWalletService
	NewMockStockCache           = mocks.NewMockStockCache
	NewMockWalletEventPublisher = mocks.NewMockWalletEventPublisher
	NewMockStockRepository      = mocks.NewMockStockRepository
	NewMockStockEventConsumer   = mocks.NewMockStockEventConsumer
)
