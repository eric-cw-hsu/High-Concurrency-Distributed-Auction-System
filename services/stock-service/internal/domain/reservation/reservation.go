package reservation

import (
	"time"
)

const (
	// ReservationTTL is the time-to-live for reservations
	ReservationTTL = 15 * time.Minute

	// MaxReservationQuantity is the maximum quantity per reservation
	MaxReservationQuantity = 10
)

// ReservationStatus represents the status of a reservation
type ReservationStatus string

const (
	ReservationStatusReserved ReservationStatus = "RESERVED"
	ReservationStatusConsumed ReservationStatus = "CONSUMED"
	ReservationStatusReleased ReservationStatus = "RELEASED"
	ReservationStatusExpired  ReservationStatus = "EXPIRED"
)

// Reservation represents a stock reservation
type Reservation struct {
	id           ReservationID
	productID    ProductID
	userID       UserID
	quantity     int
	status       ReservationStatus
	reservedAt   time.Time
	expiredAt    time.Time
	consumedAt   *time.Time
	releasedAt   *time.Time
	orderID      *string
	domainEvents []DomainEvent
}

// NewReservation creates a new reservation
func NewReservation(
	productID ProductID,
	userID UserID,
	quantity int,
) (*Reservation, error) {
	if productID.IsEmpty() {
		return nil, ErrProductIDRequired
	}
	if userID.IsEmpty() {
		return nil, ErrUserIDRequired
	}
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if quantity > MaxReservationQuantity {
		return nil, ErrExceedsMaxQuantity
	}

	now := time.Now()
	reservationID := NewReservationID()

	r := &Reservation{
		id:         reservationID,
		productID:  productID,
		userID:     userID,
		quantity:   quantity,
		status:     ReservationStatusReserved,
		reservedAt: now,
		expiredAt:  now.Add(ReservationTTL),
	}

	r.recordEvent(NewReservationCreatedEvent(reservationID, productID, userID, quantity, now))

	return r, nil
}

// ReconstructReservation reconstructs from persistence
func ReconstructReservation(
	id ReservationID,
	productID ProductID,
	userID UserID,
	quantity int,
	status ReservationStatus,
	reservedAt time.Time,
	expiredAt time.Time,
	consumedAt *time.Time,
	releasedAt *time.Time,
	orderID *string,
) *Reservation {
	return &Reservation{
		id:         id,
		productID:  productID,
		userID:     userID,
		quantity:   quantity,
		status:     status,
		reservedAt: reservedAt,
		expiredAt:  expiredAt,
		consumedAt: consumedAt,
		releasedAt: releasedAt,
		orderID:    orderID,
	}
}

// Getters
func (r *Reservation) ID() ReservationID {
	return r.id
}

func (r *Reservation) ProductID() ProductID {
	return r.productID
}

func (r *Reservation) UserID() UserID {
	return r.userID
}

func (r *Reservation) Quantity() int {
	return r.quantity
}

func (r *Reservation) Status() ReservationStatus {
	return r.status
}

func (r *Reservation) ReservedAt() time.Time {
	return r.reservedAt
}

func (r *Reservation) ExpiredAt() time.Time {
	return r.expiredAt
}

func (r *Reservation) OrderID() *string {
	return r.orderID
}

// IsExpired checks if reservation has expired
func (r *Reservation) IsExpired() bool {
	return time.Now().After(r.expiredAt)
}

// IsActive checks if reservation is still active
func (r *Reservation) IsActive() bool {
	return r.status == ReservationStatusReserved && !r.IsExpired()
}

// Consume marks reservation as consumed (order created)
func (r *Reservation) Consume(orderID string) error {
	if r.status != ReservationStatusReserved {
		return ErrCanOnlyConsumeReserved
	}
	if r.IsExpired() {
		return ErrReservationExpired
	}

	now := time.Now()
	r.status = ReservationStatusConsumed
	r.consumedAt = &now
	r.orderID = &orderID

	r.recordEvent(NewReservationConsumedEvent(r.id, r.productID, orderID, now))

	return nil
}

// Release marks reservation as released (cancelled)
func (r *Reservation) Release() error {
	if r.status != ReservationStatusReserved {
		return ErrCanOnlyReleaseReserved
	}

	now := time.Now()
	r.status = ReservationStatusReleased
	r.releasedAt = &now

	r.recordEvent(NewReservationReleasedEvent(r.id, r.productID, r.quantity, now))

	return nil
}

// MarkAsExpired marks reservation as expired
func (r *Reservation) MarkAsExpired() error {
	if r.status != ReservationStatusReserved {
		return ErrCanOnlyExpireReserved
	}

	r.status = ReservationStatusExpired

	return nil
}

// Domain events
func (r *Reservation) recordEvent(event DomainEvent) {
	r.domainEvents = append(r.domainEvents, event)
}

func (r *Reservation) DomainEvents() []DomainEvent {
	events := make([]DomainEvent, len(r.domainEvents))
	copy(events, r.domainEvents)
	return events
}

func (r *Reservation) ClearEvents() {
	r.domainEvents = nil
}
