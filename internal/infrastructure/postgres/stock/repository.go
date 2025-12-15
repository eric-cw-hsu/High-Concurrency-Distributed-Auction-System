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

func (r *PostgresStockRepository) GetStocksByProductID(ctx context.Context, productID string) ([]*stock.Stock, error) {
	// Implementation for fetching stocks by product ID from PostgreSQL database
	// This is a placeholder implementation and should be replaced with actual SQL queries
	return nil, nil
}

func (r *PostgresStockRepository) DecreaseStock(ctx context.Context, stockID string, quantity int) (int, error) {
	query := `
		UPDATE stocks
		SET quantity = quantity - $1
		WHERE id = $2 AND quantity >= $1
		RETURNING quantity
	`

	row := r.db.QueryRowContext(ctx, query, quantity, stockID)
	var remainingQuantity int
	if err := row.Scan(&remainingQuantity); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Stock not found or insufficient quantity
		}
		return 0, err // Other error
	}

	return remainingQuantity, nil
}

func (r *PostgresStockRepository) GetStockByID(ctx context.Context, stockID string) (*stock.Stock, error) {
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

	row := r.db.QueryRowContext(ctx, query, stock.ID, stock.ProductID, stock.Quantity, stock.Price, stock.SellerID)
	if err := row.Scan(&stock.ID); err != nil {
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
		if err := rows.Scan(&s.ID, &s.ProductID, &s.Quantity, &s.Price); err != nil {
			return nil, err
		}
		stocks = append(stocks, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stocks, nil
}

func (r *PostgresStockRepository) UpdateStockQuantity(ctx context.Context, stock *stock.Stock) (*stock.Stock, error) {
	query := `
		UPDATE stocks
		SET quantity = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, product_id, quantity, price, seller_id, created_at, updated_at
	`
	row := r.db.QueryRowContext(ctx, query, stock.Quantity, stock.UpdatedAt, stock.ID)
	if err := row.Scan(&stock.ID, &stock.ProductID, &stock.Quantity, &stock.Price, &stock.SellerID, &stock.CreatedAt, &stock.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Stock not found, return empty stock
		}
		return nil, err // Other error
	}

	return stock, nil
}

// Ensure PostgresStockRepository implements the StockRepository interfaces
var _ stock.StockRepository = (*PostgresStockRepository)(nil)
