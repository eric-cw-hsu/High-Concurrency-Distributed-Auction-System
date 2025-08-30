package wallet

import (
	"fmt"
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
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
	Id           string
	UserId       string
	Balance      float64
	Status       WalletStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Transactions []Transaction
	events       []domain.DomainEvent
}

// NewWalletAggregate creates a new wallet aggregate instance (without events)
// This is used for reconstituting aggregates from storage
func NewWalletAggregate(userId string) *WalletAggregate {
	now := time.Now()
	return &WalletAggregate{
		UserId:       userId,
		Balance:      0.0,
		Status:       WalletStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		Transactions: make([]Transaction, 0),
		events:       make([]domain.DomainEvent, 0),
	}
}

// ReconstructWalletAggregate reconstructs a wallet aggregate from stored data
// This is used when loading from database and should not trigger any events
func ReconstructWalletAggregate(id, userId string, balance float64, status WalletStatus,
	createdAt, updatedAt time.Time, transactions []Transaction) *WalletAggregate {
	return &WalletAggregate{
		Id:           id,
		UserId:       userId,
		Balance:      balance,
		Status:       status,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		Transactions: transactions,
		events:       make([]domain.DomainEvent, 0), // No events when reconstructing
	}
}

// CreateNewWallet creates a new wallet for business operations (with events)
// This should be used when actually creating a new wallet in the business context
func CreateNewWallet(userId string) *WalletAggregate {
	now := time.Now()
	aggregate := &WalletAggregate{
		UserId:       userId,
		Balance:      0.0,
		Status:       WalletStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		Transactions: make([]Transaction, 0),
		events:       make([]domain.DomainEvent, 0),
	}

	// Add wallet created event for business wallet creation
	aggregate.addEvent(&WalletCreatedEvent{
		UserId:    userId,
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
	w.addEvent(&FundAddedEvent{
		UserId:          w.UserId,
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
	w.addEvent(&FundSubtractedEvent{
		UserId:          w.UserId,
		Amount:          amount,
		PreviousBalance: previousBalance,
		NewBalance:      w.Balance,
		Description:     description,
		CreatedAt:       w.UpdatedAt,
	})

	return nil
}

// ProcessPayment processes a payment from the wallet
func (w *WalletAggregate) ProcessPayment(amount float64, orderId string) error {
	return w.SubtractFund(amount, fmt.Sprintf("Payment for order: %s", orderId))
}

// ProcessRefund processes a refund to the wallet
func (w *WalletAggregate) ProcessRefund(amount float64, orderId string) error {
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
		Description: fmt.Sprintf("Refund for order: %s", orderId),
		CreatedAt:   w.UpdatedAt,
	}
	w.Transactions = append(w.Transactions, transaction)

	// Add refund processed event
	w.addEvent(&RefundProcessedEvent{
		UserId:          w.UserId,
		OrderId:         orderId,
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
			UserId:    w.UserId,
			Status:    int(w.Status),
			Operation: "suspend",
		}
	}

	previousStatus := w.Status
	w.Status = WalletStatusSuspended
	w.UpdatedAt = time.Now()

	w.addEvent(&WalletSuspendedEvent{
		UserId:         w.UserId,
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
			UserId:    w.UserId,
			Status:    int(w.Status),
			Operation: "activate",
		}
	}

	previousStatus := w.Status
	w.Status = WalletStatusActive
	w.UpdatedAt = time.Now()

	w.addEvent(&WalletActivatedEvent{
		UserId:         w.UserId,
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

// GetEvents returns and clears all domain events
func (w *WalletAggregate) GetEvents() []domain.DomainEvent {
	events := make([]domain.DomainEvent, len(w.events))
	copy(events, w.events)
	w.events = w.events[:0] // Clear events
	return events
}

// addEvent adds a domain event to the aggregate
func (w *WalletAggregate) addEvent(event domain.DomainEvent) {
	w.events = append(w.events, event)
}

// verifyWalletStatus checks if the wallet is in an active state
func (w *WalletAggregate) verifyWalletStatus() error {
	if w.Status != WalletStatusActive {
		return &WalletStatusError{
			UserId:    w.UserId,
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
			UserId:    w.UserId,
			Current:   w.Balance,
			Required:  amount,
			Operation: "validate_balance",
		}
	}
	return nil
}

// GetUncommittedEvents returns uncommitted domain events
func (w *WalletAggregate) GetUncommittedEvents() []domain.DomainEvent {
	return w.events
}

// MarkEventsAsCommitted clears the uncommitted events
func (w *WalletAggregate) MarkEventsAsCommitted() {
	w.events = []domain.DomainEvent{}
}
