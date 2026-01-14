package stock

import "context"

// Repository defines the interface for stock persistence
type Repository interface {
	// Save saves stock to Redis
	Save(ctx context.Context, stock *Stock) error

	// FindByProductID finds stock by product ID
	FindByProductID(ctx context.Context, productID ProductID) (*Stock, error)

	// Exists checks if stock exists for a product
	Exists(ctx context.Context, productID ProductID) (bool, error)

	// Reserve reserves stock atomically (Lua script)
	// Returns new quantity after deduction
	Reserve(ctx context.Context, productID ProductID, quantity int) (newQuantity int, err error)

	// Release releases reserved stock
	// Returns new quantity after addition
	Release(ctx context.Context, productID ProductID, quantity int) (newQuantity int, err error)
}
