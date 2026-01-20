package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/order"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// OrderRepository implements order persistence in PostgreSQL
type OrderRepository struct {
	// db uses sqlx.ExtContext to allow both *sqlx.DB and *sqlx.Tx
	db sqlx.ExtContext
}

// NewOrderRepository creates a new order repository with a standard connection pool
func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// NewOrderRepositoryWithTx creates a decorated repository bound to a specific transaction
func NewOrderRepositoryWithTx(tx *sqlx.Tx) *OrderRepository {
	return &OrderRepository{db: tx}
}

// Save handles creating or updating an order aggregate within the current db/tx context
func (r *OrderRepository) Save(ctx context.Context, o *order.Order) error {
	model := DomainToModel(o)

	query := `
		INSERT INTO orders (
			order_id, reservation_id, user_id, product_id, quantity,
			unit_price, total_price, currency, status,
			payment_id, payment_method, payment_status,
			payment_transaction_id, payment_processed_at, payment_failure_reason,
			created_at, expires_at, paid_at, cancelled_at, cancel_reason, updated_at
		) VALUES (
			:order_id, :reservation_id, :user_id, :product_id, :quantity,
			:unit_price, :total_price, :currency, :status,
			:payment_id, :payment_method, :payment_status,
			:payment_transaction_id, :payment_processed_at, :payment_failure_reason,
			:created_at, :expires_at, :paid_at, :cancelled_at, :cancel_reason, :updated_at
		)
		ON CONFLICT (order_id) DO UPDATE SET
			status = EXCLUDED.status,
			payment_id = EXCLUDED.payment_id,
			payment_method = EXCLUDED.payment_method,
			payment_status = EXCLUDED.payment_status,
			payment_transaction_id = EXCLUDED.payment_transaction_id,
			payment_processed_at = EXCLUDED.payment_processed_at,
			payment_failure_reason = EXCLUDED.payment_failure_reason,
			paid_at = EXCLUDED.paid_at,
			cancelled_at = EXCLUDED.cancelled_at,
			cancel_reason = EXCLUDED.cancel_reason,
			updated_at = EXCLUDED.updated_at
	`

	// NamedExecContext automatically uses the underlying transaction if available
	_, err := sqlx.NamedExecContext(ctx, r.db, query, model)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	return nil
}

// FindByID retrieves a single order by its domain ID
func (r *OrderRepository) FindByID(ctx context.Context, id order.OrderID) (*order.Order, error) {
	query := `
		SELECT id, order_id, reservation_id, user_id, product_id, quantity,
			   unit_price, total_price, currency, status,
			   payment_id, payment_method, payment_status,
			   payment_transaction_id, payment_processed_at, payment_failure_reason,
			   created_at, expires_at, paid_at, cancelled_at, cancel_reason, updated_at
		FROM orders
		WHERE order_id = $1
	`

	var model OrderModel
	// Use GetContext from sqlx.ExtContext
	err := sqlx.GetContext(ctx, r.db, &model, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, order.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to find order by id: %w", err)
	}

	return ModelToDomain(&model)
}

// FindByReservationID retrieves an order linked to a specific reservation
func (r *OrderRepository) FindByReservationID(ctx context.Context, resID order.ReservationID) (*order.Order, error) {
	query := `
		SELECT id, order_id, reservation_id, user_id, product_id, quantity,
			   unit_price, total_price, currency, status,
			   payment_id, payment_method, payment_status,
			   payment_transaction_id, payment_processed_at, payment_failure_reason,
			   created_at, expires_at, paid_at, cancelled_at, cancel_reason, updated_at
		FROM orders
		WHERE reservation_id = $1
	`

	var model OrderModel
	err := sqlx.GetContext(ctx, r.db, &model, query, resID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, order.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to find order by reservation id: %w", err)
	}

	return ModelToDomain(&model)
}

// FindByUserID retrieves orders for a specific user with pagination support
func (r *OrderRepository) FindByUserID(ctx context.Context, uID order.UserID, limit, offset int) ([]*order.Order, error) {
	query := `
		SELECT id, order_id, reservation_id, user_id, product_id, quantity,
			   unit_price, total_price, currency, status,
			   payment_id, payment_method, payment_status,
			   payment_transaction_id, payment_processed_at, payment_failure_reason,
			   created_at, expires_at, paid_at, cancelled_at, cancel_reason, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var models []OrderModel
	// Use SelectContext from sqlx.ExtContext for multiple rows
	err := sqlx.SelectContext(ctx, r.db, &models, query, uID.String(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user orders: %w", err)
	}

	orders := make([]*order.Order, 0, len(models))
	for _, m := range models {
		o, err := ModelToDomain(&m)
		if err != nil {
			zap.L().Error("data corruption: failed to map order model to domain",
				zap.String("order_id", m.OrderID),
				zap.Error(err))
			continue
		}
		orders = append(orders, o)
	}

	return orders, nil
}

// FindExpired identifies orders in PENDING_PAYMENT state that passed their expiration time
func (r *OrderRepository) FindExpired(ctx context.Context, now time.Time, limit int) ([]*order.Order, error) {
	query := `
		SELECT id, order_id, reservation_id, user_id, product_id, quantity,
			   unit_price, total_price, currency, status,
			   payment_id, payment_method, payment_status,
			   payment_transaction_id, payment_processed_at, payment_failure_reason,
			   created_at, expires_at, paid_at, cancelled_at, cancel_reason, updated_at
		FROM orders
		WHERE status = 'PENDING_PAYMENT'
		  AND expires_at < $1
		ORDER BY expires_at ASC
		LIMIT $2
	`

	var models []OrderModel
	err := sqlx.SelectContext(ctx, r.db, &models, query, now, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired orders: %w", err)
	}

	orders := make([]*order.Order, 0, len(models))
	for _, m := range models {
		o, err := ModelToDomain(&m)
		if err != nil {
			zap.L().Error("data corruption: failed to map expired order model to domain",
				zap.String("order_id", m.OrderID),
				zap.Error(err))
			continue
		}
		orders = append(orders, o)
	}

	return orders, nil
}

// UpdateStatus performs a targeted update of an order's status
func (r *OrderRepository) UpdateStatus(ctx context.Context, id order.OrderID, status order.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE order_id = $2
	`

	// Use ExecContext from sqlx.ExtContext
	result, err := r.db.ExecContext(ctx, query, string(status), id.String())
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return order.ErrOrderNotFound
	}

	return nil
}
