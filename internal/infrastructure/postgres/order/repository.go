package order

import (
	"context"
	"database/sql"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) order.OrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

func (r *PostgresOrderRepository) SaveOrder(ctx context.Context, order *order.Order) error {
	query := `INSERT INTO orders (order_id, buyer_id, stock_id, total_price, quantity, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		order.OrderID,
		order.BuyerID,
		order.StockID,
		order.TotalPrice,
		order.Quantity,
		order.CreatedAt,
		order.UpdatedAt)
	if err != nil {
		return WrapRepositoryError("save_placed_order", order.OrderID, err)
	}
	return nil
}
