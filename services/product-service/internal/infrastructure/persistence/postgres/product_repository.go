package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/domain/product"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// ProductRepository implements product.Repository interface (read + simple write)
type ProductRepository struct {
	db *sqlx.DB
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Save saves a product (for simple updates without events)
func (r *ProductRepository) Save(ctx context.Context, p *product.Product) error {
	model := DomainToModel(p)

	query := `
		INSERT INTO products (
			id, seller_id, name, description,
			regular_price, flash_sale_price, currency,
			status, stock_status, created_at, updated_at
		) VALUES (
			:id, :seller_id, :name, :description,
			:regular_price, :flash_sale_price, :currency,
			:status, :stock_status, :created_at, :updated_at
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			regular_price = EXCLUDED.regular_price,
			flash_sale_price = EXCLUDED.flash_sale_price,
			currency = EXCLUDED.currency,
			status = EXCLUDED.status,
			stock_status = EXCLUDED.stock_status,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to save product: %w", err)
	}

	return nil
}

// FindByID finds a product by ID
func (r *ProductRepository) FindByID(ctx context.Context, id product.ProductID) (*product.Product, error) {
	query := `
		SELECT id, seller_id, name, description,
			   regular_price, flash_sale_price, currency,
			   status, stock_status, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var model ProductModel
	err := r.db.GetContext(ctx, &model, query, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Not found is expected, debug level
			logger.DebugContext(ctx, "product not found in database",
				zap.String("product_id", id.String()),
			)
			return nil, product.ErrProductNotFound
		}

		// Database error, error level
		logger.ErrorContext(ctx, "database query failed",
			zap.String("operation", "FindByID"),
			zap.String("product_id", id.String()),
			zap.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	return ModelToDomain(&model)
}

// FindBySeller finds all products by seller with pagination
func (r *ProductRepository) FindBySeller(
	ctx context.Context,
	sellerID product.SellerID,
	limit, offset int,
) ([]*product.Product, error) {
	query := `
		SELECT id, seller_id, name, description,
			   regular_price, flash_sale_price, currency,
			   status, stock_status, created_at, updated_at
		FROM products
		WHERE seller_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var models []ProductModel
	err := r.db.SelectContext(ctx, &models, query, sellerID.String(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find products by seller: %w", err)
	}

	products := make([]*product.Product, 0, len(models))
	for _, model := range models {
		p, err := ModelToDomain(&model)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// FindByStatus finds products by status with pagination
func (r *ProductRepository) FindByStatus(
	ctx context.Context,
	status product.ProductStatus,
	limit, offset int,
) ([]*product.Product, error) {
	query := `
		SELECT id, seller_id, name, description,
			   regular_price, flash_sale_price, currency,
			   status, stock_status, created_at, updated_at
		FROM products
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var models []ProductModel
	err := r.db.SelectContext(ctx, &models, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find products by status: %w", err)
	}

	products := make([]*product.Product, 0, len(models))
	for _, model := range models {
		p, err := ModelToDomain(&model)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// FindActiveProducts finds all active products with pagination
func (r *ProductRepository) FindActiveProducts(
	ctx context.Context,
	limit, offset int,
) ([]*product.Product, error) {
	return r.FindByStatus(ctx, product.ProductStatusActive, limit, offset)
}

// Delete deletes a product
func (r *ProductRepository) Delete(ctx context.Context, id product.ProductID) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return product.ErrProductNotFound
	}

	return nil
}

// Exists checks if a product exists
func (r *ProductRepository) Exists(ctx context.Context, id product.ProductID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, id.String())
	if err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}

	return exists, nil
}

// CountBySeller counts products by seller
func (r *ProductRepository) CountBySeller(ctx context.Context, sellerID product.SellerID) (int, error) {
	query := `SELECT COUNT(*) FROM products WHERE seller_id = $1`

	var count int
	err := r.db.GetContext(ctx, &count, query, sellerID.String())
	if err != nil {
		return 0, fmt.Errorf("failed to count products by seller: %w", err)
	}

	return count, nil
}

// CountByStatus counts products by status
func (r *ProductRepository) CountByStatus(ctx context.Context, status product.ProductStatus) (int, error) {
	query := `SELECT COUNT(*) FROM products WHERE status = $1`

	var count int
	err := r.db.GetContext(ctx, &count, query, string(status))
	if err != nil {
		return 0, fmt.Errorf("failed to count products by status: %w", err)
	}

	return count, nil
}
