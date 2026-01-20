package postgres

import (
	"context"
	"database/sql"
	"fmt"

	productprice "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/product_price"
	"github.com/jmoiron/sqlx"
)

type ProductPriceRepository struct {
	db sqlx.ExtContext
}

func NewProductPriceRepository(db *sqlx.DB) *ProductPriceRepository {
	return &ProductPriceRepository{db: db}
}

// GetByID retrieves the price for a specific product from local cache.
func (r *ProductPriceRepository) GetByID(ctx context.Context, productID string) (*productprice.ProductPrice, error) {
	query := `SELECT product_id, unit_price, currency, updated_at FROM product_prices WHERE product_id = $1`

	var p productprice.ProductPrice
	err := sqlx.GetContext(ctx, r.db, &p, query, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product price not found in local cache: %s", productID)
		}
		return nil, err
	}
	return &p, nil
}

// Upsert updates or inserts product pricing when receiving events from Product Service.
func (r *ProductPriceRepository) Upsert(ctx context.Context, p *productprice.ProductPrice) error {
	query := `
		INSERT INTO product_prices (product_id, unit_price, currency, updated_at)
		VALUES (:product_id, :unit_price, :currency, :updated_at)
		ON CONFLICT (product_id) DO UPDATE SET
			unit_price = EXCLUDED.unit_price,
			currency = EXCLUDED.currency,
			updated_at = EXCLUDED.updated_at
	`
	_, err := sqlx.NamedExecContext(ctx, r.db, query, p)
	return err
}
