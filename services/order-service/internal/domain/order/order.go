package order

import (
	"time"
)

const (
	OrderExpirationDuration = 15 * time.Minute
)

// Order represents an order aggregate
type Order struct {
	id            OrderID
	reservationID ReservationID
	userID        UserID
	productID     ProductID
	quantity      int
	pricing       Pricing
	payment       *Payment
	status        OrderStatus
	createdAt     time.Time
	updatedAt     time.Time
	expiresAt     time.Time
	paidAt        *time.Time
	cancelledAt   *time.Time
	cancelReason  *string
	events        []DomainEvent
}

// NewOrder creates a new order from a reservation
func NewOrder(
	reservationID ReservationID,
	userID UserID,
	productID ProductID,
	quantity int,
	unitPrice Money,
) (*Order, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	pricing, err := NewPricing(unitPrice, quantity)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	order := &Order{
		id:            NewOrderID(),
		reservationID: reservationID,
		userID:        userID,
		productID:     productID,
		quantity:      quantity,
		pricing:       pricing,
		status:        OrderStatusPendingPayment,
		createdAt:     now,
		updatedAt:     now,
		expiresAt:     now.Add(OrderExpirationDuration),
		events:        []DomainEvent{},
	}

	// Record domain event
	order.recordEvent(NewOrderCreatedEvent(
		order.id,
		order.reservationID,
		order.userID,
		order.productID,
		order.quantity,
		order.pricing,
		now,
	))

	return order, nil
}

// ProcessPayment processes payment for the order
func (o *Order) ProcessPayment(method PaymentMethod, transactionID string) error {
	// Validate can pay
	if err := o.ValidateCanPay(); err != nil {
		return err
	}

	// Create payment if not exists
	if o.payment == nil {
		o.payment = NewPayment(o.id, o.pricing.TotalPrice(), method)
	}

	// Mark payment as completed
	if err := o.payment.MarkAsCompleted(transactionID); err != nil {
		return err
	}

	// Update order status
	now := time.Now()
	o.status = OrderStatusPaid
	o.paidAt = &now

	// Record domain event
	o.recordEvent(NewOrderPaidEvent(
		o.id,
		o.reservationID,
		o.payment.ID(),
		transactionID,
		now,
	))

	return nil
}

// RecordPaymentFailure records a payment failure
func (o *Order) RecordPaymentFailure(reason string) error {
	if o.payment == nil {
		o.payment = NewPayment(o.id, o.pricing.TotalPrice(), PaymentMethodMock)
	}

	return o.payment.MarkAsFailed(reason)
}

// Cancel cancels the order
func (o *Order) Cancel(reason string) error {
	if o.status == OrderStatusPaid {
		return ErrOrderAlreadyPaid
	}

	if o.status == OrderStatusCancelled || o.status == OrderStatusExpired {
		return ErrOrderAlreadyCancelled
	}

	now := time.Now()
	o.status = OrderStatusCancelled
	o.cancelledAt = &now
	o.cancelReason = &reason

	// Record domain event
	o.recordEvent(NewOrderCancelledEvent(
		o.id,
		o.reservationID,
		reason,
		now,
	))

	return nil
}

// MarkAsExpired marks the order as expired
func (o *Order) MarkAsExpired() error {
	if o.status != OrderStatusPendingPayment {
		return ErrInvalidOrderStatus
	}

	now := time.Now()
	o.status = OrderStatusExpired
	o.cancelledAt = &now
	reason := "payment timeout"
	o.cancelReason = &reason

	// Record domain event
	o.recordEvent(NewOrderCancelledEvent(
		o.id,
		o.reservationID,
		reason,
		now,
	))

	return nil
}

// ValidateCanPay validates if order can be paid
func (o *Order) ValidateCanPay() error {
	if o.status != OrderStatusPendingPayment {
		return ErrInvalidOrderStatus
	}

	if time.Now().After(o.expiresAt) {
		return ErrOrderExpired
	}

	return nil
}

// IsExpired checks if order has expired
func (o *Order) IsExpired() bool {
	return time.Now().After(o.expiresAt) && o.status == OrderStatusPendingPayment
}

// Getters
func (o *Order) ID() OrderID {
	return o.id
}

func (o *Order) ReservationID() ReservationID {
	return o.reservationID
}

func (o *Order) UserID() UserID {
	return o.userID
}

func (o *Order) ProductID() ProductID {
	return o.productID
}

func (o *Order) Quantity() int {
	return o.quantity
}

func (o *Order) Pricing() Pricing {
	return o.pricing
}

func (o *Order) Payment() *Payment {
	return o.payment
}

func (o *Order) Status() OrderStatus {
	return o.status
}

func (o *Order) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Order) ExpiresAt() time.Time {
	return o.expiresAt
}

func (o *Order) UpdatedAt() time.Time {
	return o.updatedAt
}

func (o *Order) PaidAt() *time.Time {
	return o.paidAt
}

func (o *Order) CancelledAt() *time.Time {
	return o.cancelledAt
}

func (o *Order) CancelReason() *string {
	return o.cancelReason
}

// Domain Events
func (o *Order) DomainEvents() []DomainEvent {
	return o.events
}

func (o *Order) ClearEvents() {
	o.events = []DomainEvent{}
}

func (o *Order) recordEvent(event DomainEvent) {
	o.events = append(o.events, event)
}

// ReconstructOrder reconstructs an order from persistence (for repository)
func ReconstructOrder(
	id OrderID,
	reservationID ReservationID,
	userID UserID,
	productID ProductID,
	quantity int,
	pricing Pricing,
	payment *Payment,
	status OrderStatus,
	createdAt time.Time,
	updatedAt time.Time,
	expiresAt time.Time,
	paidAt *time.Time,
	cancelledAt *time.Time,
	cancelReason *string,
) *Order {
	return &Order{
		id:            id,
		reservationID: reservationID,
		userID:        userID,
		productID:     productID,
		quantity:      quantity,
		pricing:       pricing,
		payment:       payment,
		status:        status,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		expiresAt:     expiresAt,
		paidAt:        paidAt,
		cancelledAt:   cancelledAt,
		cancelReason:  cancelReason,
		events:        []DomainEvent{},
	}
}
