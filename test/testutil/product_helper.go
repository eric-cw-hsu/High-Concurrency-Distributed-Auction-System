package testutil

import (
	"context"
	"database/sql"
	"fmt"
)

type ProductTestHelper struct {
	db *sql.DB
}

func NewProductTestHelper(db *sql.DB) *ProductTestHelper {
	return &ProductTestHelper{db: db}
}

// CreateTestProduct creates a test product in the database
func (h *ProductTestHelper) CreateTestProduct(ctx context.Context, productID, name, description string) error {
	query := `
		INSERT INTO products (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`

	_, err := h.db.ExecContext(ctx, query, productID, name, description)
	if err != nil {
		return fmt.Errorf("failed to create test product: %w", err)
	}

	return nil
} // CreateTestProducts creates multiple test products
func (h *ProductTestHelper) CreateTestProducts(ctx context.Context, count int) ([]string, error) {
	var productIDs []string

	for i := 0; i < count; i++ {
		productID := fmt.Sprintf("550e8400-e29b-41d4-a716-4466554400%02d", i)
		name := fmt.Sprintf("Test Product %d", i+1)
		description := fmt.Sprintf("Description for test product %d", i+1)

		err := h.CreateTestProduct(ctx, productID, name, description)
		if err != nil {
			return nil, err
		}

		productIDs = append(productIDs, productID)
	}

	return productIDs, nil
}

// GetProductByID retrieves a product by ID for testing
func (h *ProductTestHelper) GetProductByID(ctx context.Context, productID string) (*TestProduct, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM products WHERE id = $1`

	var product TestProduct
	err := h.db.QueryRowContext(ctx, query, productID).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
} // DeleteTestProduct deletes a test product
func (h *ProductTestHelper) DeleteTestProduct(ctx context.Context, productID string) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := h.db.ExecContext(ctx, query, productID)
	if err != nil {
		return fmt.Errorf("failed to delete test product: %w", err)
	}
	return nil
}

// TestProduct represents a product for testing
type TestProduct struct {
	ID          string
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
}
