package order

import "time"

// DomainEvent is the interface for domain events
type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

// OrderCreatedEvent is emitted when an order is created
type OrderCreatedEvent struct {
	OrderID       OrderID
	ReservationID ReservationID
	UserID        UserID
	ProductID     ProductID
	Quantity      int
	Pricing       Pricing
	occurredAt    time.Time
}

func NewOrderCreatedEvent(
	orderID OrderID,
	reservationID ReservationID,
	userID UserID,
	productID ProductID,
	quantity int,
	pricing Pricing,
	occurredAt time.Time,
) OrderCreatedEvent {
	return OrderCreatedEvent{
		OrderID:       orderID,
		ReservationID: reservationID,
		UserID:        userID,
		ProductID:     productID,
		Quantity:      quantity,
		Pricing:       pricing,
		occurredAt:    occurredAt,
	}
}

func (e OrderCreatedEvent) EventType() string {
	return "order.created"
}

func (e OrderCreatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// OrderPaidEvent is emitted when an order is paid
type OrderPaidEvent struct {
	OrderID       OrderID
	ReservationID ReservationID
	PaymentID     PaymentID
	TransactionID string
	occurredAt    time.Time
}

func NewOrderPaidEvent(
	orderID OrderID,
	reservationID ReservationID,
	paymentID PaymentID,
	transactionID string,
	occurredAt time.Time,
) OrderPaidEvent {
	return OrderPaidEvent{
		OrderID:       orderID,
		ReservationID: reservationID,
		PaymentID:     paymentID,
		TransactionID: transactionID,
		occurredAt:    occurredAt,
	}
}

func (e OrderPaidEvent) EventType() string {
	return "order.paid"
}

func (e OrderPaidEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// OrderCancelledEvent is emitted when an order is cancelled
type OrderCancelledEvent struct {
	OrderID       OrderID
	ReservationID ReservationID
	Reason        string
	occurredAt    time.Time
}

func NewOrderCancelledEvent(
	orderID OrderID,
	reservationID ReservationID,
	reason string,
	occurredAt time.Time,
) OrderCancelledEvent {
	return OrderCancelledEvent{
		OrderID:       orderID,
		ReservationID: reservationID,
		Reason:        reason,
		occurredAt:    occurredAt,
	}
}

func (e OrderCancelledEvent) EventType() string {
	return "order.cancelled"
}

func (e OrderCancelledEvent) OccurredAt() time.Time {
	return e.occurredAt
}
