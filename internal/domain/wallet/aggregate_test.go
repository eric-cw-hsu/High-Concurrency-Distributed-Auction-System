package wallet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNewWallet(t *testing.T) {
	userId := "user-123"

	wallet := CreateNewWallet(userId)

	// ID might not be set by CreateNewWallet, that's ok for business operations
	assert.Equal(t, userId, wallet.UserId)
	assert.Equal(t, 0.0, wallet.Balance)
	assert.Equal(t, WalletStatusActive, wallet.Status)
	assert.NotZero(t, wallet.CreatedAt)
	assert.NotZero(t, wallet.UpdatedAt)
	assert.Empty(t, wallet.Transactions)

	// Check events
	events := wallet.GetUncommittedEvents()
	require.Len(t, events, 1)

	event, ok := events[0].(*WalletCreatedEvent)
	require.True(t, ok)
	assert.Equal(t, userId, event.UserId)
}

func TestWalletAggregate_AddFund(t *testing.T) {
	tests := []struct {
		name        string
		amount      float64
		description string
		expectError bool
	}{
		{
			name:        "Valid amount",
			amount:      100.0,
			description: "Initial deposit",
			expectError: false,
		},
		{
			name:        "Zero amount",
			amount:      0.0,
			description: "Zero deposit",
			expectError: true,
		},
		{
			name:        "Negative amount",
			amount:      -50.0,
			description: "Negative deposit",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := CreateNewWallet("user-123")
			wallet.MarkEventsAsCommitted() // Clear creation events

			err := wallet.AddFund(tt.amount, tt.description)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, 0.0, wallet.Balance)
				assert.Empty(t, wallet.GetUncommittedEvents())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.amount, wallet.Balance)
				assert.Len(t, wallet.Transactions, 1)

				transaction := wallet.Transactions[0]
				assert.Equal(t, TransactionTypeDeposit, transaction.Type)
				assert.Equal(t, tt.amount, transaction.Amount)
				assert.Equal(t, tt.description, transaction.Description)

				// Check events
				events := wallet.GetUncommittedEvents()
				require.Len(t, events, 1)

				event, ok := events[0].(*FundAddedEvent)
				require.True(t, ok)
				assert.Equal(t, "user-123", event.UserId)
				assert.Equal(t, tt.amount, event.Amount)
				assert.Equal(t, tt.amount, event.NewBalance)
			}
		})
	}
}

func TestWalletAggregate_ProcessPayment(t *testing.T) {
	tests := []struct {
		name           string
		initialBalance float64
		paymentAmount  float64
		orderId        string
		expectError    bool
	}{
		{
			name:           "Sufficient balance",
			initialBalance: 100.0,
			paymentAmount:  50.0,
			orderId:        "order-123",
			expectError:    false,
		},
		{
			name:           "Exact balance",
			initialBalance: 100.0,
			paymentAmount:  100.0,
			orderId:        "order-123",
			expectError:    false,
		},
		{
			name:           "Insufficient balance",
			initialBalance: 50.0,
			paymentAmount:  100.0,
			orderId:        "order-123",
			expectError:    true,
		},
		{
			name:           "Zero payment",
			initialBalance: 100.0,
			paymentAmount:  0.0,
			orderId:        "order-123",
			expectError:    true,
		},
		{
			name:           "Negative payment",
			initialBalance: 100.0,
			paymentAmount:  -50.0,
			orderId:        "order-123",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := CreateNewWallet("user-123")
			if tt.initialBalance > 0 {
				_ = wallet.AddFund(tt.initialBalance, "Initial fund")
			}
			wallet.MarkEventsAsCommitted() // Clear setup events

			err := wallet.ProcessPayment(tt.paymentAmount, tt.orderId)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.initialBalance, wallet.Balance)
			} else {
				assert.NoError(t, err)
				expectedBalance := tt.initialBalance - tt.paymentAmount
				assert.Equal(t, expectedBalance, wallet.Balance)

				// Check transaction
				foundTransaction := false
				for _, transaction := range wallet.Transactions {
					if transaction.Type == TransactionTypeWithdrawal && transaction.Amount == tt.paymentAmount {
						foundTransaction = true
						break
					}
				}
				assert.True(t, foundTransaction, "Payment transaction not found")

				// Check events
				events := wallet.GetUncommittedEvents()
				require.Len(t, events, 1)

				event, ok := events[0].(*FundSubtractedEvent)
				require.True(t, ok)
				assert.Equal(t, "user-123", event.UserId)
				assert.Equal(t, tt.paymentAmount, event.Amount)
				assert.Equal(t, expectedBalance, event.NewBalance)
			}
		})
	}
}

func TestWalletAggregate_ProcessRefund(t *testing.T) {
	wallet := CreateNewWallet("user-123")
	_ = wallet.AddFund(100.0, "Initial fund")
	_ = wallet.ProcessPayment(50.0, "order-123")
	wallet.MarkEventsAsCommitted() // Clear setup events

	err := wallet.ProcessRefund(30.0, "order-123")

	assert.NoError(t, err)
	assert.Equal(t, 80.0, wallet.Balance) // 100 - 50 + 30

	// Check transaction
	foundTransaction := false
	for _, transaction := range wallet.Transactions {
		if transaction.Type == TransactionTypeRefund && transaction.Amount == 30.0 {
			foundTransaction = true
			assert.Contains(t, transaction.Description, "Refund")
			break
		}
	}
	assert.True(t, foundTransaction, "Refund transaction not found")

	// Check events
	events := wallet.GetUncommittedEvents()
	require.Len(t, events, 1)

	event, ok := events[0].(*RefundProcessedEvent)
	require.True(t, ok)
	assert.Equal(t, "user-123", event.UserId)
	assert.Equal(t, 30.0, event.Amount)
	assert.Equal(t, "order-123", event.OrderId)
	assert.Equal(t, 80.0, event.NewBalance)
}

func TestReconstructWalletAggregate(t *testing.T) {
	id := "wallet-123"
	userId := "user-123"
	balance := 150.0
	status := WalletStatusActive
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()
	transactions := []Transaction{
		{
			Type:        TransactionTypeDeposit,
			Amount:      100.0,
			Description: "Initial deposit",
			CreatedAt:   createdAt,
		},
		{
			Type:        TransactionTypePayment,
			Amount:      50.0,
			Description: "Payment for order-123",
			CreatedAt:   time.Now(),
		},
	}

	wallet := ReconstructWalletAggregate(id, userId, balance, status, createdAt, updatedAt, transactions)

	assert.Equal(t, id, wallet.Id)
	assert.Equal(t, userId, wallet.UserId)
	assert.Equal(t, balance, wallet.Balance)
	assert.Equal(t, status, wallet.Status)
	assert.Equal(t, createdAt, wallet.CreatedAt)
	assert.Equal(t, updatedAt, wallet.UpdatedAt)
	assert.Equal(t, transactions, wallet.Transactions)

	// Reconstructed aggregate should not have any uncommitted events
	assert.Empty(t, wallet.GetUncommittedEvents())
}

func TestWalletAggregate_EventCommitting(t *testing.T) {
	wallet := CreateNewWallet("user-123")

	// Initially has creation event
	events := wallet.GetUncommittedEvents()
	assert.Len(t, events, 1)

	// Mark as committed
	wallet.MarkEventsAsCommitted()
	events = wallet.GetUncommittedEvents()
	assert.Empty(t, events)

	// Add fund creates new event
	_ = wallet.AddFund(100.0, "Test")
	events = wallet.GetUncommittedEvents()
	assert.Len(t, events, 1)

	// Mark as committed again
	wallet.MarkEventsAsCommitted()
	events = wallet.GetUncommittedEvents()
	assert.Empty(t, events)
}
