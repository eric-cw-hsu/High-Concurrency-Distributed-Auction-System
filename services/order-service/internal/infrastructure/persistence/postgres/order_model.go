package postgres

import (
	"database/sql"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/order"
)

// OrderModel represents the database model for orders
type OrderModel struct {
	ID            int64  `db:"id"`
	OrderID       string `db:"order_id"`
	ReservationID string `db:"reservation_id"`
	UserID        string `db:"user_id"`
	ProductID     string `db:"product_id"`
	Quantity      int    `db:"quantity"`
	UnitPrice     int64  `db:"unit_price"`
	TotalPrice    int64  `db:"total_price"`
	Currency      string `db:"currency"`
	Status        string `db:"status"`

	// Payment info (embedded)
	PaymentID            sql.NullString `db:"payment_id"`
	PaymentMethod        sql.NullString `db:"payment_method"`
	PaymentStatus        sql.NullString `db:"payment_status"`
	PaymentTransactionID sql.NullString `db:"payment_transaction_id"`
	PaymentProcessedAt   sql.NullTime   `db:"payment_processed_at"`
	PaymentFailureReason sql.NullString `db:"payment_failure_reason"`

	CreatedAt    time.Time      `db:"created_at"`
	ExpiresAt    time.Time      `db:"expires_at"`
	PaidAt       sql.NullTime   `db:"paid_at"`
	CancelledAt  sql.NullTime   `db:"cancelled_at"`
	CancelReason sql.NullString `db:"cancel_reason"`
	UpdatedAt    time.Time      `db:"updated_at"`
}

// DomainToModel converts domain order to database model
func DomainToModel(o *order.Order) *OrderModel {
	model := &OrderModel{
		OrderID:       o.ID().String(),
		ReservationID: o.ReservationID().String(),
		UserID:        o.UserID().String(),
		ProductID:     o.ProductID().String(),
		Quantity:      o.Quantity(),
		UnitPrice:     o.Pricing().UnitPrice().Amount(),
		TotalPrice:    o.Pricing().TotalPrice().Amount(),
		Currency:      o.Pricing().TotalPrice().Currency(),
		Status:        string(o.Status()),
		CreatedAt:     o.CreatedAt(),
		ExpiresAt:     o.ExpiresAt(),
		UpdatedAt:     o.UpdatedAt(),
	}

	if o.PaidAt() != nil {
		model.PaidAt = sql.NullTime{Time: *o.PaidAt(), Valid: true}
	}

	if o.CancelledAt() != nil {
		model.CancelledAt = sql.NullTime{Time: *o.CancelledAt(), Valid: true}
	}

	if o.CancelReason() != nil {
		model.CancelReason = sql.NullString{String: *o.CancelReason(), Valid: true}
	}

	// Payment info
	if payment := o.Payment(); payment != nil {
		model.PaymentID = sql.NullString{String: payment.ID().String(), Valid: true}
		model.PaymentMethod = sql.NullString{String: string(payment.Method()), Valid: true}
		model.PaymentStatus = sql.NullString{String: string(payment.Status()), Valid: true}

		if payment.TransactionID() != nil {
			model.PaymentTransactionID = sql.NullString{String: *payment.TransactionID(), Valid: true}
		}

		if payment.ProcessedAt() != nil {
			model.PaymentProcessedAt = sql.NullTime{Time: *payment.ProcessedAt(), Valid: true}
		}

		if payment.FailureReason() != nil {
			model.PaymentFailureReason = sql.NullString{String: *payment.FailureReason(), Valid: true}
		}
	}

	return model
}

// ModelToDomain converts database model to domain order
func ModelToDomain(m *OrderModel) (*order.Order, error) {
	orderID, err := order.ParseOrderID(m.OrderID)
	if err != nil {
		return nil, err
	}

	reservationID, err := order.ParseReservationID(m.ReservationID)
	if err != nil {
		return nil, err
	}

	userID, err := order.ParseUserID(m.UserID)
	if err != nil {
		return nil, err
	}

	productID, err := order.ParseProductID(m.ProductID)
	if err != nil {
		return nil, err
	}

	unitPrice, err := order.NewMoney(m.UnitPrice, m.Currency)
	if err != nil {
		return nil, err
	}

	pricing, err := order.NewPricing(unitPrice, m.Quantity)
	if err != nil {
		return nil, err
	}

	// Reconstruct payment if exists
	var payment *order.Payment
	if m.PaymentID.Valid {
		paymentID, _ := order.ParsePaymentID(m.PaymentID.String)
		payment = order.ReconstructPayment(
			paymentID,
			orderID,
			pricing.TotalPrice(),
			order.PaymentMethod(m.PaymentMethod.String),
			order.PaymentStatus(m.PaymentStatus.String),
			nullStringToPtr(m.PaymentTransactionID),
			nullTimeToPtr(m.PaymentProcessedAt),
			nullStringToPtr(m.PaymentFailureReason),
			m.CreatedAt,
		)
	}

	return order.ReconstructOrder(
		orderID,
		reservationID,
		userID,
		productID,
		m.Quantity,
		pricing,
		payment,
		order.OrderStatus(m.Status),
		m.CreatedAt,
		m.UpdatedAt,
		m.ExpiresAt,
		nullTimeToPtr(m.PaidAt),
		nullTimeToPtr(m.CancelledAt),
		nullStringToPtr(m.CancelReason),
	), nil
}

// Helper functions
func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
