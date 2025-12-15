package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrderAggregate_Success(t *testing.T) {
	// Arrange
	buyerId := "buyer-123"
	stockId := "stock-456"
	price := 100.0
	quantity := 5

	// Act
	aggregate, err := NewOrderAggregate(buyerId, stockId, price, quantity)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, aggregate)
	assert.NotEmpty(t, aggregate.ID)
	assert.Equal(t, buyerId, aggregate.BuyerID)
	assert.Equal(t, stockId, aggregate.StockID)
	assert.Equal(t, quantity, aggregate.Quantity)
	assert.Equal(t, 500.0, aggregate.Price) // price * quantity
	assert.Equal(t, OrderStatusPending, aggregate.Status())
	assert.NotZero(t, aggregate.CreatedAt)
	assert.NotZero(t, aggregate.UpdatedAt)
	assert.Empty(t, aggregate.PopEventPayloads()) // No events initially
}

func TestNewOrderAggregate_Validation(t *testing.T) {
	tests := []struct {
		name        string
		buyerId     string
		stockId     string
		price       float64
		quantity    int
		expectedErr string
	}{
		{
			name:        "Empty buyer ID",
			buyerId:     "",
			stockId:     "stock-123",
			price:       100.0,
			quantity:    1,
			expectedErr: "buyer ID cannot be empty",
		},
		{
			name:        "Empty stock ID",
			buyerId:     "buyer-123",
			stockId:     "",
			price:       100.0,
			quantity:    1,
			expectedErr: "stock ID cannot be empty",
		},
		{
			name:        "Zero quantity",
			buyerId:     "buyer-123",
			stockId:     "stock-123",
			price:       100.0,
			quantity:    0,
			expectedErr: "quantity must be positive",
		},
		{
			name:        "Negative quantity",
			buyerId:     "buyer-123",
			stockId:     "stock-123",
			price:       100.0,
			quantity:    -1,
			expectedErr: "quantity must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			aggregate, err := NewOrderAggregate(tt.buyerId, tt.stockId, tt.price, tt.quantity)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, aggregate)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestOrderAggregate_ConfirmAfterStockDeduction_Success(t *testing.T) {
	// Arrange
	aggregate, err := NewOrderAggregate("buyer-123", "stock-456", 100.0, 2)
	require.NoError(t, err)

	confirmTime := time.Now()

	// Act
	err = aggregate.ConfirmAfterStockDeduction(confirmTime)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusConfirmed, aggregate.Status())
	assert.True(t, aggregate.UpdatedAt.After(aggregate.CreatedAt))

	// Check event generation
	events := aggregate.PopEventPayloads()
	assert.Len(t, events, 1)

	event, ok := events[0].(*OrderReservedPayload)
	assert.True(t, ok)
	assert.Equal(t, aggregate.ID, event.OrderID)
	assert.Equal(t, aggregate.BuyerID, event.BuyerID)
	assert.Equal(t, aggregate.StockID, event.StockID)
	assert.Equal(t, aggregate.Quantity, event.Quantity)
	assert.Equal(t, aggregate.Price, event.TotalPrice)
	assert.Equal(t, confirmTime, event.Timestamp)
}

func TestOrderAggregate_ConfirmAfterStockDeduction_InvalidStatus(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus OrderStatus
		setupFunc     func(*OrderAggregate) error
	}{
		{
			name:          "Already confirmed",
			initialStatus: OrderStatusConfirmed,
			setupFunc: func(o *OrderAggregate) error {
				return o.ConfirmAfterStockDeduction(time.Now())
			},
		},
		{
			name:          "Already cancelled",
			initialStatus: OrderStatusCancelled,
			setupFunc: func(o *OrderAggregate) error {
				return o.Cancel()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			aggregate, err := NewOrderAggregate("buyer-123", "stock-456", 100.0, 2)
			require.NoError(t, err)

			// Setup initial status
			err = tt.setupFunc(aggregate)
			require.NoError(t, err)

			// Act
			err = aggregate.ConfirmAfterStockDeduction(time.Now())

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "order cannot be confirmed")
		})
	}
}

func TestOrderAggregate_Cancel_Success(t *testing.T) {
	// Arrange
	aggregate, err := NewOrderAggregate("buyer-123", "stock-456", 100.0, 2)
	require.NoError(t, err)

	// Act
	err = aggregate.Cancel()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusCancelled, aggregate.Status())
	assert.True(t, aggregate.UpdatedAt.After(aggregate.CreatedAt))

	// No events should be generated for cancellation
	events := aggregate.PopEventPayloads()
	assert.Empty(t, events)
}

func TestOrderAggregate_Cancel_InvalidStatus(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*OrderAggregate) error
		expectedError string
	}{
		{
			name: "Already confirmed",
			setupFunc: func(o *OrderAggregate) error {
				return o.ConfirmAfterStockDeduction(time.Now())
			},
			expectedError: "order cannot be cancelled",
		},
		{
			name: "Already cancelled",
			setupFunc: func(o *OrderAggregate) error {
				return o.Cancel()
			},
			expectedError: "order cannot be cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			aggregate, err := NewOrderAggregate("buyer-123", "stock-456", 100.0, 2)
			require.NoError(t, err)

			// Setup status
			err = tt.setupFunc(aggregate)
			require.NoError(t, err)

			// Act
			err = aggregate.Cancel()

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestOrderAggregate_TotalPriceCalculation(t *testing.T) {
	tests := []struct {
		name          string
		price         float64
		quantity      int
		expectedTotal float64
	}{
		{
			name:          "Integer price",
			price:         100.0,
			quantity:      3,
			expectedTotal: 300.0,
		},
		{
			name:          "Decimal price",
			price:         99.99,
			quantity:      2,
			expectedTotal: 199.98,
		},
		{
			name:          "Single quantity",
			price:         150.5,
			quantity:      1,
			expectedTotal: 150.5,
		},
		{
			name:          "Large quantity",
			price:         10.0,
			quantity:      100,
			expectedTotal: 1000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			aggregate, err := NewOrderAggregate("buyer-123", "stock-456", tt.price, tt.quantity)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, aggregate.Price)
		})
	}
}

func TestOrderAggregate_StatusTransitions(t *testing.T) {
	// Test valid status transition: Pending -> Confirmed
	t.Run("Pending to Confirmed", func(t *testing.T) {
		aggregate, err := NewOrderAggregate("buyer-123", "stock-456", 100.0, 2)
		require.NoError(t, err)
		assert.Equal(t, OrderStatusPending, aggregate.Status())

		err = aggregate.ConfirmAfterStockDeduction(time.Now())
		assert.NoError(t, err)
		assert.Equal(t, OrderStatusConfirmed, aggregate.Status())
	})

	// Test valid status transition: Pending -> Cancelled
	t.Run("Pending to Cancelled", func(t *testing.T) {
		aggregate, err := NewOrderAggregate("buyer-123", "stock-456", 100.0, 2)
		require.NoError(t, err)
		assert.Equal(t, OrderStatusPending, aggregate.Status())

		err = aggregate.Cancel()
		assert.NoError(t, err)
		assert.Equal(t, OrderStatusCancelled, aggregate.Status())
	})

	// Test invalid transitions from Confirmed
	t.Run("No transitions from Confirmed", func(t *testing.T) {
		aggregate, err := NewOrderAggregate("buyer-123", "stock-456", 100.0, 2)
		require.NoError(t, err)

		err = aggregate.ConfirmAfterStockDeduction(time.Now())
		require.NoError(t, err)

		// Cannot confirm again
		err = aggregate.ConfirmAfterStockDeduction(time.Now())
		assert.Error(t, err)

		// Cannot cancel
		err = aggregate.Cancel()
		assert.Error(t, err)
	})

	// Test invalid transitions from Cancelled
	t.Run("No transitions from Cancelled", func(t *testing.T) {
		aggregate, err := NewOrderAggregate("buyer-123", "stock-456", 100.0, 2)
		require.NoError(t, err)

		err = aggregate.Cancel()
		require.NoError(t, err)

		// Cannot confirm
		err = aggregate.ConfirmAfterStockDeduction(time.Now())
		assert.Error(t, err)

		// Cannot cancel again
		err = aggregate.Cancel()
		assert.Error(t, err)
	})
}
