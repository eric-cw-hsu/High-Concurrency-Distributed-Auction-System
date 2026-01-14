package reservation

import "time"

// DomainEvent interface
type DomainEvent interface {
	OccurredAt() time.Time
	EventType() string
}

// ReservationCreatedEvent is emitted when stock is reserved
type ReservationCreatedEvent struct {
	ReservationID ReservationID
	ProductID     ProductID
	UserID        UserID
	Quantity      int
	occurredAt    time.Time
}

func NewReservationCreatedEvent(
	reservationID ReservationID,
	productID ProductID,
	userID UserID,
	quantity int,
	occurredAt time.Time,
) ReservationCreatedEvent {
	return ReservationCreatedEvent{
		ReservationID: reservationID,
		ProductID:     productID,
		UserID:        userID,
		Quantity:      quantity,
		occurredAt:    occurredAt,
	}
}

func (e ReservationCreatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ReservationCreatedEvent) EventType() string {
	return "stock.reserved"
}

// ReservationConsumedEvent is emitted when reservation is consumed
type ReservationConsumedEvent struct {
	ReservationID ReservationID
	ProductID     ProductID
	OrderID       string
	occurredAt    time.Time
}

func NewReservationConsumedEvent(
	reservationID ReservationID,
	productID ProductID,
	orderID string,
	occurredAt time.Time,
) ReservationConsumedEvent {
	return ReservationConsumedEvent{
		ReservationID: reservationID,
		ProductID:     productID,
		OrderID:       orderID,
		occurredAt:    occurredAt,
	}
}

func (e ReservationConsumedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ReservationConsumedEvent) EventType() string {
	return "stock.consumed"
}

// ReservationReleasedEvent is emitted when reservation is released
type ReservationReleasedEvent struct {
	ReservationID ReservationID
	ProductID     ProductID
	Quantity      int
	occurredAt    time.Time
}

func NewReservationReleasedEvent(
	reservationID ReservationID,
	productID ProductID,
	quantity int,
	occurredAt time.Time,
) ReservationReleasedEvent {
	return ReservationReleasedEvent{
		ReservationID: reservationID,
		ProductID:     productID,
		Quantity:      quantity,
		occurredAt:    occurredAt,
	}
}

func (e ReservationReleasedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ReservationReleasedEvent) EventType() string {
	return "stock.released"
}
