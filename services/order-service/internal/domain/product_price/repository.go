package productprice

import "context"

// Repository defines the storage contract for product price snapshots.
// This interface lives in the Domain/Application layer to maintain decoupling.
type Repository interface {
	GetByID(ctx context.Context, productID string) (*ProductPrice, error)
	Upsert(ctx context.Context, price *ProductPrice) error
}
