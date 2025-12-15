package wallet

import (
	"fmt"
	"time"
)

type WalletStatus int

const (
	WalletStatusActive WalletStatus = iota
	WalletStatusInactive
	WalletStatusSuspended
)

// Transaction represents a wallet transaction
type Transaction struct {
	Type        TransactionType `json:"type"`
	Amount      float64         `json:"amount"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}

type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
	TransactionTypeRefund     TransactionType = "REFUND"
	TransactionTypePayment    TransactionType = "PAYMENT"
)

type WalletAggregate struct {
	ID            string
	UserID        string
	Balance       float64
	Status        WalletStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Transactions  []Transaction
	eventPayloads []interface{}
}

// NewWalletAggregate creates a new wallet aggregate instance (without events)
// This is used for reconstituting aggregates from storage
func NewWalletAggregate(userId string) *WalletAggregate {
	now := time.Now()
	return &WalletAggregate{
		UserID:        userId,
		Balance:       0.0,
		Status:        WalletStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
		Transactions:  make([]Transaction, 0),
		eventPayloads: make([]interface{}, 0),
	}
}

// ReconstructWalletAggregate reconstructs a wallet aggregate from stored data
// This is used when loading from database and should not trigger any events
func ReconstructWalletAggregate(id, userId string, balance float64, status WalletStatus,
	createdAt, updatedAt time.Time, transactions []Transaction) *WalletAggregate {
	return &WalletAggregate{
		ID:            id,
		UserID:        userId,
		Balance:       balance,
		Status:        status,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		Transactions:  transactions,
		eventPayloads: make([]interface{}, 0), // No events when reconstructing
	}
}

// CreateNewWallet creates a new wallet for business operations (with events)
// This should be used when actually creating a new wallet in the business context
func CreateNewWallet(userID string) *WalletAggregate {
	now := time.Now()
	aggregate := &WalletAggregate{
		UserID:        userID,
		Balance:       0.0,
		Status:        WalletStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
		Transactions:  make([]Transaction, 0),
		eventPayloads: make([]interface{}, 0),
	}

	// Add wallet created event for business wallet creation
	aggregate.addEvent(&WalletCreatedPayload{
		UserID:    userID,
		CreatedAt: now,
	})

	return aggregate
}

// AddFund adds funds to the wallet
func (w *WalletAggregate) AddFund(amount float64, description string) error {
	if err := w.verifyWalletStatus(); err != nil {
		return err
	}
	if err := w.validateAmount(amount); err != nil {
		return err
	}

	previousBalance := w.Balance
	w.Balance += amount
	w.UpdatedAt = time.Now()

	// Add transaction record
	transaction := Transaction{
		Type:        TransactionTypeDeposit,
		Amount:      amount,
		Description: description,
		CreatedAt:   w.UpdatedAt,
	}
	w.Transactions = append(w.Transactions, transaction)

	// Add fund added event
	w.addEvent(&FundAddedPayload{
		UserID:          w.UserID,
		Amount:          amount,
		PreviousBalance: previousBalance,
		NewBalance:      w.Balance,
		Description:     description,
		CreatedAt:       w.UpdatedAt,
	})

	return nil
}

// SubtractFund subtracts funds from the wallet
func (w *WalletAggregate) SubtractFund(amount float64, description string) error {
	if err := w.verifyWalletStatus(); err != nil {
		return err
	}
	if err := w.validateAmount(amount); err != nil {
		return err
	}
	if err := w.validateSufficientBalance(amount); err != nil {
		return err
	}

	previousBalance := w.Balance
	w.Balance -= amount
	w.UpdatedAt = time.Now()

	// Add transaction record
	transaction := Transaction{
		Type:        TransactionTypeWithdrawal,
		Amount:      amount,
		Description: description,
		CreatedAt:   w.UpdatedAt,
	}
	w.Transactions = append(w.Transactions, transaction)

	// Add fund subtracted event
	w.addEvent(&FundSubtractedPayload{
		UserID:          w.UserID,
		Amount:          amount,
		PreviousBalance: previousBalance,
		NewBalance:      w.Balance,
		Description:     description,
		CreatedAt:       w.UpdatedAt,
	})

	return nil
}

// ProcessPayment processes a payment from the wallet
func (w *WalletAggregate) ProcessPayment(amount float64, orderID string) error {
	return w.SubtractFund(amount, fmt.Sprintf("Payment for order: %s", orderID))
}

// ProcessRefund processes a refund to the wallet
func (w *WalletAggregate) ProcessRefund(amount float64, orderID string) error {
	if err := w.verifyWalletStatus(); err != nil {
		return err
	}
	if err := w.validateAmount(amount); err != nil {
		return err
	}

	previousBalance := w.Balance
	w.Balance += amount
	w.UpdatedAt = time.Now()

	// Add transaction record
	transaction := Transaction{
		Type:        TransactionTypeRefund,
		Amount:      amount,
		Description: fmt.Sprintf("Refund for order: %s", orderID),
		CreatedAt:   w.UpdatedAt,
	}
	w.Transactions = append(w.Transactions, transaction)

	// Add refund processed event
	w.addEvent(&RefundProcessedPayload{
		UserID:          w.UserID,
		OrderID:         orderID,
		Amount:          amount,
		PreviousBalance: previousBalance,
		NewBalance:      w.Balance,
		CreatedAt:       w.UpdatedAt,
	})

	return nil
}

// Suspend suspends the wallet
func (w *WalletAggregate) Suspend(reason string) error {
	if w.Status == WalletStatusSuspended {
		return &WalletStatusError{
			UserID:    w.UserID,
			Status:    int(w.Status),
			Operation: "suspend",
		}
	}

	previousStatus := w.Status
	w.Status = WalletStatusSuspended
	w.UpdatedAt = time.Now()

	w.addEvent(&WalletSuspendedPayload{
		UserID:         w.UserID,
		PreviousStatus: previousStatus,
		Reason:         reason,
		SuspendedAt:    w.UpdatedAt,
	})

	return nil
}

// Activate activates the wallet
func (w *WalletAggregate) Activate() error {
	if w.Status == WalletStatusActive {
		return &WalletStatusError{
			UserID:    w.UserID,
			Status:    int(w.Status),
			Operation: "activate",
		}
	}

	previousStatus := w.Status
	w.Status = WalletStatusActive
	w.UpdatedAt = time.Now()

	w.addEvent(&WalletActivatedPayload{
		UserID:         w.UserID,
		PreviousStatus: previousStatus,
		ActivatedAt:    w.UpdatedAt,
	})

	return nil
}

// GetBalance returns the current balance
func (w *WalletAggregate) GetBalance() float64 {
	return w.Balance
}

// GetStatus returns the current status
func (w *WalletAggregate) GetStatus() WalletStatus {
	return w.Status
}

// PopEventPayloads returns and clears all event payloads
func (w *WalletAggregate) PopEventPayloads() []interface{} {
	payloads := w.eventPayloads
	w.eventPayloads = make([]interface{}, 0)
	return payloads
}

// addEvent adds an event payload to the aggregate
func (w *WalletAggregate) addEvent(payload interface{}) {
	w.eventPayloads = append(w.eventPayloads, payload)
}

// verifyWalletStatus checks if the wallet is in an active state
func (w *WalletAggregate) verifyWalletStatus() error {
	if w.Status != WalletStatusActive {
		return &WalletStatusError{
			UserID:    w.UserID,
			Status:    int(w.Status),
			Operation: "verify_status",
		}
	}
	return nil
}

// validateAmount validates that the amount is positive
func (w *WalletAggregate) validateAmount(amount float64) error {
	if amount <= 0 {
		return &InvalidAmountError{
			Amount:    amount,
			Operation: "validate_amount",
		}
	}
	return nil
}

// validateSufficientBalance checks if the wallet has sufficient balance
func (w *WalletAggregate) validateSufficientBalance(amount float64) error {
	if w.Balance < amount {
		return &InsufficientBalanceError{
			UserID:    w.UserID,
			Current:   w.Balance,
			Required:  amount,
			Operation: "validate_balance",
		}
	}
	return nil
}
