package order

import (
	"time"
)

// Payment represents a payment entity within Order aggregate
type Payment struct {
	id            PaymentID
	orderID       OrderID
	amount        Money
	method        PaymentMethod
	status        PaymentStatus
	transactionID *string
	processedAt   *time.Time
	failureReason *string
	createdAt     time.Time
}

// NewPayment creates a new payment
func NewPayment(
	orderID OrderID,
	amount Money,
	method PaymentMethod,
) *Payment {
	return &Payment{
		id:        NewPaymentID(),
		orderID:   orderID,
		amount:    amount,
		method:    method,
		status:    PaymentStatusPending,
		createdAt: time.Now(),
	}
}

// MarkAsCompleted marks the payment as completed
func (p *Payment) MarkAsCompleted(transactionID string) error {
	if p.status == PaymentStatusCompleted {
		return ErrPaymentAlreadyCompleted
	}

	if p.status == PaymentStatusFailed {
		return ErrCannotCompleteFailedPayment
	}

	now := time.Now()
	p.status = PaymentStatusCompleted
	p.transactionID = &transactionID
	p.processedAt = &now

	return nil
}

// MarkAsFailed marks the payment as failed
func (p *Payment) MarkAsFailed(reason string) error {
	if p.status == PaymentStatusCompleted {
		return ErrCannotFailCompletedPayment
	}

	p.status = PaymentStatusFailed
	p.failureReason = &reason

	return nil
}

// Getters
func (p *Payment) ID() PaymentID {
	return p.id
}

func (p *Payment) OrderID() OrderID {
	return p.orderID
}

func (p *Payment) Amount() Money {
	return p.amount
}

func (p *Payment) Method() PaymentMethod {
	return p.method
}

func (p *Payment) Status() PaymentStatus {
	return p.status
}

func (p *Payment) TransactionID() *string {
	return p.transactionID
}

func (p *Payment) ProcessedAt() *time.Time {
	return p.processedAt
}

func (p *Payment) FailureReason() *string {
	return p.failureReason
}

func (p *Payment) CreatedAt() time.Time {
	return p.createdAt
}

// IsCompleted checks if payment is completed
func (p *Payment) IsCompleted() bool {
	return p.status == PaymentStatusCompleted
}

// IsFailed checks if payment failed
func (p *Payment) IsFailed() bool {
	return p.status == PaymentStatusFailed
}

// IsPending checks if payment is pending
func (p *Payment) IsPending() bool {
	return p.status == PaymentStatusPending
}

func ReconstructPayment(
	id PaymentID,
	orderID OrderID,
	amount Money,
	method PaymentMethod,
	status PaymentStatus,
	transactionID *string,
	processedAt *time.Time,
	failureReason *string,
	createdAt time.Time,
) *Payment {
	return &Payment{
		id:            id,
		orderID:       orderID,
		amount:        amount,
		method:        method,
		status:        status,
		transactionID: transactionID,
		processedAt:   processedAt,
		failureReason: failureReason,
		createdAt:     createdAt,
	}
}
