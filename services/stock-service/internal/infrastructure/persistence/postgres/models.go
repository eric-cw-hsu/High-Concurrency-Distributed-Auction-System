package postgres

import (
	"database/sql"
	"time"
)

// ReservationModel represents the database model for reservations
type ReservationModel struct {
	ID            string         `db:"id"`
	ReservationID string         `db:"reservation_id"`
	ProductID     string         `db:"product_id"`
	UserID        string         `db:"user_id"`
	Quantity      int            `db:"quantity"`
	Status        string         `db:"status"`
	ReservedAt    time.Time      `db:"reserved_at"`
	ExpiredAt     time.Time      `db:"expired_at"`
	ConsumedAt    sql.NullTime   `db:"consumed_at"`
	ReleasedAt    sql.NullTime   `db:"released_at"`
	OrderID       sql.NullString `db:"order_id"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}

// OutboxEventModel represents the database model for outbox events
type OutboxEventModel struct {
	ID            string         `db:"id"`
	AggregateType string         `db:"aggregate_type"`
	AggregateID   string         `db:"aggregate_id"`
	EventType     string         `db:"event_type"`
	EventID       string         `db:"event_id"`
	Payload       []byte         `db:"payload"`
	Status        string         `db:"status"`
	CreatedAt     time.Time      `db:"created_at"`
	ProcessedAt   sql.NullTime   `db:"processed_at"`
	RetryCount    int            `db:"retry_count"`
	LastError     sql.NullString `db:"last_error"`
	NextRetryAt   sql.NullTime   `db:"next_retry_at"`
}
