package wallet

import (
	"time"
)

// WalletCreatedPayload represents a wallet creation event payload
type WalletCreatedPayload struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// FundAddedPayload represents a fund addition event payload
type FundAddedPayload struct {
	UserID          string    `json:"user_id"`
	Amount          float64   `json:"amount"`
	PreviousBalance float64   `json:"previous_balance"`
	NewBalance      float64   `json:"new_balance"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

// FundSubtractedPayload represents a fund subtraction event payload
type FundSubtractedPayload struct {
	UserID          string    `json:"user_id"`
	Amount          float64   `json:"amount"`
	PreviousBalance float64   `json:"previous_balance"`
	NewBalance      float64   `json:"new_balance"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

// RefundProcessedPayload represents a refund processing event payload
type RefundProcessedPayload struct {
	UserID          string    `json:"user_id"`
	OrderID         string    `json:"order_id"`
	Amount          float64   `json:"amount"`
	PreviousBalance float64   `json:"previous_balance"`
	NewBalance      float64   `json:"new_balance"`
	CreatedAt       time.Time `json:"created_at"`
}

// WalletSuspendedPayload represents a wallet suspension event payload
type WalletSuspendedPayload struct {
	UserID         string       `json:"user_id"`
	PreviousStatus WalletStatus `json:"previous_status"`
	Reason         string       `json:"reason"`
	SuspendedAt    time.Time    `json:"suspended_at"`
}

// WalletActivatedPayload represents a wallet activation event payload
type WalletActivatedPayload struct {
	UserID         string       `json:"user_id"`
	PreviousStatus WalletStatus `json:"previous_status"`
	ActivatedAt    time.Time    `json:"activated_at"`
}
