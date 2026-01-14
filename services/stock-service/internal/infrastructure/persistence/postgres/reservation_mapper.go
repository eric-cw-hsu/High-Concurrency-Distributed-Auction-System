package postgres

import (
	"database/sql"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"github.com/samborkent/uuidv7"
)

// DomainToModel converts domain Reservation to database model
func DomainToModel(r *reservation.Reservation) *ReservationModel {
	model := &ReservationModel{
		ID:            uuidv7.New().String(),
		ReservationID: r.ID().String(),
		ProductID:     r.ProductID().String(),
		UserID:        r.UserID().String(),
		Quantity:      r.Quantity(),
		Status:        string(r.Status()),
		ReservedAt:    r.ReservedAt(),
		ExpiredAt:     r.ExpiredAt(),
		CreatedAt:     r.ReservedAt(),
		UpdatedAt:     r.ReservedAt(),
	}

	// Handle optional order_id
	if orderID := r.OrderID(); orderID != nil {
		model.OrderID = sql.NullString{String: *orderID, Valid: true}
	}

	return model
}

// ModelToDomain converts database model to domain Reservation
func ModelToDomain(model *ReservationModel) (*reservation.Reservation, error) {
	rid, err := reservation.ParseReservationID(model.ReservationID)
	if err != nil {
		return nil, err
	}

	pid, err := reservation.ParseProductID(model.ProductID)
	if err != nil {
		return nil, err
	}

	uid, err := reservation.ParseUserID(model.UserID)
	if err != nil {
		return nil, err
	}

	var consumedAt *time.Time
	if model.ConsumedAt.Valid {
		consumedAt = &model.ConsumedAt.Time
	}

	var releasedAt *time.Time
	if model.ReleasedAt.Valid {
		releasedAt = &model.ReleasedAt.Time
	}

	var orderID *string
	if model.OrderID.Valid {
		orderID = &model.OrderID.String
	}

	return reservation.ReconstructReservation(
		rid,
		pid,
		uid,
		model.Quantity,
		reservation.ReservationStatus(model.Status),
		model.ReservedAt,
		model.ExpiredAt,
		consumedAt,
		releasedAt,
		orderID,
	), nil
}
