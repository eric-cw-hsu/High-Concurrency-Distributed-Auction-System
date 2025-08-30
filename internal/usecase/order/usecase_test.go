package order

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil"
)

func TestPlaceOrderUsecase_Execute_Success(t *testing.T) {
	// Arrange
	mockProducer := testutil.NewMockOrderProducer()
	mockStockCache := testutil.NewMockStockCache()
	mockWalletService := testutil.NewMockWalletService()

	usecase := NewPlaceOrderUsecase(mockProducer, mockStockCache, mockWalletService)

	command := order.PlaceOrderCommand{
		BuyerId:  "buyer-123",
		StockId:  "stock-456",
		Quantity: 5,
	}

	stockPrice := 100.0
	availableStock := 10
	totalAmount := stockPrice * float64(command.Quantity) // 500.0

	// Setup initial stock and wallet state
	mockStockCache.SetInitialStock(command.StockId, availableStock, stockPrice)
	mockWalletService.SetBalance(command.BuyerId, 1000.0)

	// Act
	err := usecase.Execute(context.Background(), command)

	// Give a moment for async operation to complete
	time.Sleep(50 * time.Millisecond)

	// Assert
	assert.NoError(t, err)

	// Verify final state
	assert.Equal(t, availableStock-command.Quantity, mockStockCache.GetCurrentStock(command.StockId))
	assert.Equal(t, 1000.0-totalAmount, mockWalletService.GetBalance(command.BuyerId))
	assert.Equal(t, int64(1), mockProducer.GetPublishCount())
}

func TestPlaceOrderUsecase_Execute_InsufficientStock(t *testing.T) {
	// Arrange
	mockProducer := testutil.NewMockOrderProducer()
	mockStockCache := testutil.NewMockStockCache()
	mockWalletService := testutil.NewMockWalletService()

	usecase := NewPlaceOrderUsecase(mockProducer, mockStockCache, mockWalletService)

	command := order.PlaceOrderCommand{
		BuyerId:  "buyer-123",
		StockId:  "stock-456",
		Quantity: 5,
	}

	// Setup: stock price but insufficient quantity
	stockPrice := 100.0
	availableStock := 3 // Less than requested
	mockStockCache.SetInitialStock(command.StockId, availableStock, stockPrice)
	mockWalletService.SetBalance(command.BuyerId, 1000.0)

	// Act
	err := usecase.Execute(context.Background(), command)

	// Assert - using domain error
	assert.Error(t, err)
	assert.ErrorIs(t, err, order.ErrInsufficientStock)

	// Verify detailed error information
	var stockErr *order.InsufficientStockError
	assert.ErrorAs(t, err, &stockErr)
	assert.Equal(t, command.StockId, stockErr.StockId)
	assert.Equal(t, availableStock, stockErr.Available)
	assert.Equal(t, command.Quantity, stockErr.Requested)

	// Verify no stock was deducted and no events published
	assert.Equal(t, availableStock, mockStockCache.GetCurrentStock(command.StockId)) // Stock unchanged
	assert.Equal(t, 1000.0, mockWalletService.GetBalance(command.BuyerId))           // Balance unchanged
	assert.Equal(t, int64(0), mockProducer.GetPublishCount())                        // No events published
}

func TestPlaceOrderUsecase_Execute_GetPriceFailure(t *testing.T) {
	// Arrange
	mockProducer := testutil.NewMockOrderProducer()
	mockStockCache := testutil.NewMockStockCache()
	mockWalletService := testutil.NewMockWalletService()

	usecase := NewPlaceOrderUsecase(mockProducer, mockStockCache, mockWalletService)

	command := order.PlaceOrderCommand{
		BuyerId:  "buyer-123",
		StockId:  "stock-456",
		Quantity: 5,
	}

	// Setup: don't set initial stock, which will cause price lookup to fail
	// (no stock data available)

	// Act
	err := usecase.Execute(context.Background(), command)

	// Assert - using domain error
	assert.Error(t, err)
	assert.ErrorIs(t, err, order.ErrStockPriceUnavailable)

	// Verify detailed error information
	var stockErr *order.StockError
	assert.ErrorAs(t, err, &stockErr)
	assert.Equal(t, command.StockId, stockErr.StockId)
	assert.Equal(t, "get_price", stockErr.Operation)
}

func TestPlaceOrderUsecase_Execute_InsufficientFunds(t *testing.T) {
	// Arrange
	mockProducer := testutil.NewMockOrderProducer()
	mockStockCache := testutil.NewMockStockCache()
	mockWalletService := testutil.NewMockWalletService()

	usecase := NewPlaceOrderUsecase(mockProducer, mockStockCache, mockWalletService)

	command := order.PlaceOrderCommand{
		BuyerId:  "buyer-123",
		StockId:  "stock-456",
		Quantity: 5,
	}

	stockPrice := 100.0
	availableStock := 10

	// Setup: insufficient wallet balance
	mockStockCache.SetInitialStock(command.StockId, availableStock, stockPrice)
	mockWalletService.SetBalance(command.BuyerId, 300.0) // Less than needed

	// Act
	err := usecase.Execute(context.Background(), command)

	// Assert - using domain error
	assert.Error(t, err)
	assert.ErrorIs(t, err, order.ErrPaymentProcessingFailed)

	// Verify detailed error information
	var paymentErr *order.PaymentError
	assert.ErrorAs(t, err, &paymentErr)
	assert.Equal(t, command.BuyerId, paymentErr.UserId)
	assert.Equal(t, 500.0, paymentErr.Amount) // 100 * 5

	// Verify no stock was deducted and no events published
	assert.Equal(t, availableStock, mockStockCache.GetCurrentStock(command.StockId)) // Stock unchanged
	assert.Equal(t, 300.0, mockWalletService.GetBalance(command.BuyerId))            // Balance unchanged
	assert.Equal(t, int64(0), mockProducer.GetPublishCount())                        // No events published
}
