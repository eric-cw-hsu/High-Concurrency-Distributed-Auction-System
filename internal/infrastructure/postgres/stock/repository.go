package stock

import (
	"context"
	"database/sql"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/stock"
)

type PostgresStockRepository struct {
	db *sql.DB
}

func NewPostgresStockRepository(db *sql.DB) *PostgresStockRepository {
	return &PostgresStockRepository{
		db: db,
	}
}

func (r *PostgresStockRepository) GetStocksByProductId(ctx context.Context, productId string) ([]*stock.Stock, error) {
	// Implementation for fetching stocks by product ID from PostgreSQL database
	// This is a placeholder implementation and should be replaced with actual SQL queries
	return nil, nil
}

func (r *PostgresStockRepository) DecreaseStock(ctx context.Context, stockId string, quantity int) (int64, error) {
	query := `
		UPDATE stocks
		SET quantity = quantity - $1
		WHERE id = $2 AND quantity >= $1
		RETURNING quantity
	`

	row := r.db.QueryRowContext(ctx, query, quantity, stockId)
	var remainingQuantity int64
	if err := row.Scan(&remainingQuantity); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Stock not found or insufficient quantity
		}
		return 0, err // Other error
	}

	return remainingQuantity, nil
}

func (r *PostgresStockRepository) GetStockById(ctx context.Context, stockId string) (*stock.Stock, error) {
	// Implementation for fetching stock by stock ID from PostgreSQL database
	// This is a placeholder implementation and should be replaced with actual SQL queries
	return nil, nil
}

func (r *PostgresStockRepository) SaveStock(ctx context.Context, stock *stock.Stock) (*stock.Stock, error) {
	query := `
		INSERT INTO stocks (id, product_id, quantity, price, seller_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	row := r.db.QueryRowContext(ctx, query, stock.Id, stock.ProductId, stock.Quantity, stock.Price, stock.SellerId)
	if err := row.Scan(&stock.Id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Stock not found, return empty stock
		}
		return nil, err // Other error
	}

	return stock, nil
}

func (r *PostgresStockRepository) GetAllStocks(ctx context.Context) ([]*stock.Stock, error) {
	query := `
		SELECT id, product_id, quantity, price 
		FROM stocks
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []*stock.Stock

	for rows.Next() {
		var s stock.Stock
		if err := rows.Scan(&s.Id, &s.ProductId, &s.Quantity, &s.Price); err != nil {
			return nil, err
		}
		stocks = append(stocks, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stocks, nil
}
