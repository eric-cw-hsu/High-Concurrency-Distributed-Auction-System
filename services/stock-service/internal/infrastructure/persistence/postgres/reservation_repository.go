package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// ReservationRepository implements reservation persistence in PostgreSQL
type ReservationRepository struct {
	db *sqlx.DB
}

var _ reservation.PersistentRepository = (*ReservationRepository)(nil)

// NewReservationRepository creates a new ReservationRepository
func NewReservationRepository(db *sqlx.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// SaveToPostgres saves reservation to PostgreSQL
func (r *ReservationRepository) Save(ctx context.Context, res *reservation.Reservation) error {
	model := DomainToModel(res)

	query := `
		INSERT INTO stock_reservations (
			id, reservation_id, product_id, user_id,
			quantity, status, reserved_at, expired_at,
			consumed_at, released_at, order_id, created_at, updated_at
		) VALUES (
			:id, :reservation_id, :product_id, :user_id,
			:quantity, :status, :reserved_at, :expired_at,
			:consumed_at, :released_at, :order_id, :created_at, :updated_at
		)
		ON CONFLICT (reservation_id) DO UPDATE SET
			status = EXCLUDED.status,
			consumed_at = EXCLUDED.consumed_at,
			released_at = EXCLUDED.released_at,
			order_id = EXCLUDED.order_id,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		logger.ErrorContext(ctx, "failed to save reservation to postgresql",
			zap.String("reservation_id", res.ID().String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to save reservation: %w", err)
	}

	return nil
}

// FindByID finds reservation by ID from PostgreSQL
func (r *ReservationRepository) FindByID(ctx context.Context, id reservation.ReservationID) (*reservation.Reservation, error) {
	query := `
		SELECT id, reservation_id, product_id, user_id,
			   quantity, status, reserved_at, expired_at,
			   consumed_at, released_at, order_id, created_at, updated_at
		FROM stock_reservations
		WHERE reservation_id = $1
	`

	var model ReservationModel
	err := r.db.GetContext(ctx, &model, query, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.DebugContext(ctx, "reservation not found in postgresql",
				zap.String("reservation_id", id.String()),
			)
			return nil, reservation.ErrReservationNotFound
		}

		logger.ErrorContext(ctx, "database query failed",
			zap.String("operation", "FindByID"),
			zap.String("reservation_id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to find reservation: %w", err)
	}

	return ModelToDomain(&model)
}

// FindActiveByProductID finds all active reservations for a product
func (r *ReservationRepository) FindActiveByProductID(
	ctx context.Context,
	productID reservation.ProductID,
) ([]*reservation.Reservation, error) {
	query := `
		SELECT id, reservation_id, product_id, user_id,
			   quantity, status, reserved_at, expired_at,
			   consumed_at, released_at, order_id, created_at, updated_at
		FROM stock_reservations
		WHERE product_id = $1
		  AND status = 'RESERVED'
		  AND expired_at > NOW()
		ORDER BY reserved_at DESC
	`

	var models []ReservationModel
	err := r.db.SelectContext(ctx, &models, query, productID.String())
	if err != nil {
		logger.ErrorContext(ctx, "failed to query active reservations",
			zap.String("product_id", productID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to find active reservations: %w", err)
	}

	reservations := make([]*reservation.Reservation, 0, len(models))
	for _, model := range models {
		res, err := ModelToDomain(&model)
		if err != nil {
			logger.ErrorContext(ctx, "failed to convert model to domain",
				zap.String("reservation_id", model.ReservationID),
				zap.Error(err),
			)
			continue
		}
		reservations = append(reservations, res)
	}

	return reservations, nil
}

// UpdateStatus updates reservation status
func (r *ReservationRepository) UpdateStatus(
	ctx context.Context,
	id reservation.ReservationID,
	status reservation.ReservationStatus,
) error {
	query := `
		UPDATE stock_reservations
		SET status = $1, updated_at = NOW()
		WHERE reservation_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, string(status), id.String())
	if err != nil {
		logger.ErrorContext(ctx, "failed to update reservation status",
			zap.String("reservation_id", id.String()),
			zap.String("status", string(status)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return reservation.ErrReservationNotFound
	}

	return nil
}

// FindAllActive finds all active reservations
func (r *ReservationRepository) FindAllActive(ctx context.Context) ([]*reservation.Reservation, error) {
	logger.InfoContext(ctx, "querying all active reservations from postgresql")

	query := `
		SELECT id, reservation_id, product_id, user_id,
			   quantity, status, reserved_at, expired_at,
			   consumed_at, released_at, order_id, created_at, updated_at
		FROM stock_reservations
		WHERE status = 'RESERVED'
		  AND expired_at > NOW()
		ORDER BY reserved_at ASC
	`

	var models []ReservationModel
	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		logger.ErrorContext(ctx, "failed to query active reservations",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to query active reservations: %w", err)
	}

	reservations := make([]*reservation.Reservation, 0, len(models))
	for _, model := range models {
		res, err := ModelToDomain(&model)
		if err != nil {
			logger.ErrorContext(ctx, "failed to convert model to domain",
				zap.String("reservation_id", model.ReservationID),
				zap.Error(err),
			)
			continue
		}
		reservations = append(reservations, res)
	}

	logger.InfoContext(ctx, "active reservations queried",
		zap.Int("count", len(reservations)),
	)

	return reservations, nil
}
