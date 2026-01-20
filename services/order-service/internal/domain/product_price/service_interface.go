package productprice

import "context"

// ProductPriceSyncer defines the contract for synchronizing product information.
// Placing this in the domain layer prevents infra from depending on application services.
type ProductPriceSyncer interface {
	SyncProductPrice(ctx context.Context, productID string, price int64, currency string) error
}
