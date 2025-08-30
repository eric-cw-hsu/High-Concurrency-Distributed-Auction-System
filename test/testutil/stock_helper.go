package testutil

import (
	"context"
	"database/sql"
	"fmt"
)

type StockTestHelper struct {
	db *sql.DB
}

func NewStockTestHelper(db *sql.DB) *StockTestHelper {
	return &StockTestHelper{db: db}
}

// CreateTestStock creates a test stock record in the database
func (h *StockTestHelper) CreateTestStock(ctx context.Context, stockID, productID, sellerID string, price float64, quantity int) error {
	query := `
		INSERT INTO stocks (id, product_id, seller_id, price, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`

	_, err := h.db.ExecContext(ctx, query, stockID, productID, sellerID, price, quantity)
	if err != nil {
		return fmt.Errorf("failed to create test stock: %w", err)
	}

	return nil
}

// CreateTestStocks creates multiple test stock records
func (h *StockTestHelper) CreateTestStocks(ctx context.Context, productIDs []string, sellerIDs []string, quantity int) ([]string, error) {
	var stockIDs []string

	for i, productID := range productIDs {
		stockID := fmt.Sprintf("660e8400-e29b-41d4-a716-4466554400%02d", i)
		sellerID := sellerIDs[i%len(sellerIDs)] // Cycle through sellers if fewer than products
		price := float64((i + 1) * 100)         // Different prices for different stocks

		err := h.CreateTestStock(ctx, stockID, productID, sellerID, price, quantity)
		if err != nil {
			return nil, err
		}

		stockIDs = append(stockIDs, stockID)
	}

	return stockIDs, nil
}

// CreateTestStockWithQuantity creates a single test stock with specific quantity
func (h *StockTestHelper) CreateTestStockWithQuantity(ctx context.Context, stockID, productID, sellerID string, price float64, quantity int) error {
	return h.CreateTestStock(ctx, stockID, productID, sellerID, price, quantity)
}

// GetStockByID retrieves a stock record by ID for testing
func (h *StockTestHelper) GetStockByID(ctx context.Context, stockID string) (*TestStock, error) {
	query := `SELECT id, product_id, seller_id, price, quantity, created_at, updated_at FROM stocks WHERE id = $1`

	var stock TestStock
	err := h.db.QueryRowContext(ctx, query, stockID).Scan(
		&stock.ID,
		&stock.ProductID,
		&stock.SellerID,
		&stock.Price,
		&stock.Quantity,
		&stock.CreatedAt,
		&stock.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	return &stock, nil
}

// UpdateStockQuantity updates the quantity of a stock
func (h *StockTestHelper) UpdateStockQuantity(ctx context.Context, stockID string, quantity int) error {
	query := `UPDATE stocks SET quantity = $1, updated_at = NOW() WHERE id = $2`

	_, err := h.db.ExecContext(ctx, query, quantity, stockID)
	if err != nil {
		return fmt.Errorf("failed to update stock quantity: %w", err)
	}

	return nil
}

// DeleteTestStock deletes a test stock record
func (h *StockTestHelper) DeleteTestStock(ctx context.Context, stockID string) error {
	query := `DELETE FROM stocks WHERE id = $1`
	_, err := h.db.ExecContext(ctx, query, stockID)
	if err != nil {
		return fmt.Errorf("failed to delete test stock: %w", err)
	}
	return nil
}

// GetStocksByProductID retrieves all stocks for a product
func (h *StockTestHelper) GetStocksByProductID(ctx context.Context, productID string) ([]*TestStock, error) {
	query := `SELECT id, product_id, seller_id, price, quantity, created_at, updated_at FROM stocks WHERE product_id = $1`

	rows, err := h.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stocks by product ID: %w", err)
	}
	defer rows.Close()

	var stocks []*TestStock
	for rows.Next() {
		var stock TestStock
		err := rows.Scan(
			&stock.ID,
			&stock.ProductID,
			&stock.SellerID,
			&stock.Price,
			&stock.Quantity,
			&stock.CreatedAt,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock row: %w", err)
		}
		stocks = append(stocks, &stock)
	}

	return stocks, nil
}

// TestStock represents a stock record for testing
type TestStock struct {
	ID        string
	ProductID string
	SellerID  string
	Price     float64
	Quantity  int
	CreatedAt string
	UpdatedAt string
}
