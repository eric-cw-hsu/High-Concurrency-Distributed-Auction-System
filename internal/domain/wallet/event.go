package wallet

import (
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
)

// WalletCreatedEvent represents a wallet creation event
type WalletCreatedEvent struct {
	UserId    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (e *WalletCreatedEvent) EventType() string {
	return "wallet.created"
}

func (e *WalletCreatedEvent) AggregateId() string {
	return e.UserId
}

func (e *WalletCreatedEvent) OccurredOn() time.Time {
	return e.CreatedAt
}

// FundAddedEvent represents a fund addition event
type FundAddedEvent struct {
	UserId          string    `json:"user_id"`
	Amount          float64   `json:"amount"`
	PreviousBalance float64   `json:"previous_balance"`
	NewBalance      float64   `json:"new_balance"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

func (e *FundAddedEvent) EventType() string {
	return "wallet.fund_added"
}

func (e *FundAddedEvent) AggregateId() string {
	return e.UserId
}

func (e *FundAddedEvent) OccurredOn() time.Time {
	return e.CreatedAt
}

// FundSubtractedEvent represents a fund subtraction event
type FundSubtractedEvent struct {
	UserId          string    `json:"user_id"`
	Amount          float64   `json:"amount"`
	PreviousBalance float64   `json:"previous_balance"`
	NewBalance      float64   `json:"new_balance"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

func (e *FundSubtractedEvent) EventType() string {
	return "wallet.fund_subtracted"
}

func (e *FundSubtractedEvent) AggregateId() string {
	return e.UserId
}

func (e *FundSubtractedEvent) OccurredOn() time.Time {
	return e.CreatedAt
}

// RefundProcessedEvent represents a refund processing event
type RefundProcessedEvent struct {
	UserId          string    `json:"user_id"`
	OrderId         string    `json:"order_id"`
	Amount          float64   `json:"amount"`
	PreviousBalance float64   `json:"previous_balance"`
	NewBalance      float64   `json:"new_balance"`
	CreatedAt       time.Time `json:"created_at"`
}

func (e *RefundProcessedEvent) EventType() string {
	return "wallet.refund_processed"
}

func (e *RefundProcessedEvent) AggregateId() string {
	return e.UserId
}

func (e *RefundProcessedEvent) OccurredOn() time.Time {
	return e.CreatedAt
}

// WalletSuspendedEvent represents a wallet suspension event
type WalletSuspendedEvent struct {
	UserId         string       `json:"user_id"`
	PreviousStatus WalletStatus `json:"previous_status"`
	Reason         string       `json:"reason"`
	SuspendedAt    time.Time    `json:"suspended_at"`
}

func (e *WalletSuspendedEvent) EventType() string {
	return "wallet.suspended"
}

func (e *WalletSuspendedEvent) AggregateId() string {
	return e.UserId
}

func (e *WalletSuspendedEvent) OccurredOn() time.Time {
	return e.SuspendedAt
}

// WalletActivatedEvent represents a wallet activation event
type WalletActivatedEvent struct {
	UserId         string       `json:"user_id"`
	PreviousStatus WalletStatus `json:"previous_status"`
	ActivatedAt    time.Time    `json:"activated_at"`
}

func (e *WalletActivatedEvent) EventType() string {
	return "wallet.activated"
}

func (e *WalletActivatedEvent) AggregateId() string {
	return e.UserId
}

func (e *WalletActivatedEvent) OccurredOn() time.Time {
	return e.ActivatedAt
}

// Ensure all events implement DomainEvent interface
var _ domain.DomainEvent = (*WalletCreatedEvent)(nil)
var _ domain.DomainEvent = (*FundAddedEvent)(nil)
var _ domain.DomainEvent = (*FundSubtractedEvent)(nil)
var _ domain.DomainEvent = (*RefundProcessedEvent)(nil)
var _ domain.DomainEvent = (*WalletSuspendedEvent)(nil)
var _ domain.DomainEvent = (*WalletActivatedEvent)(nil)
