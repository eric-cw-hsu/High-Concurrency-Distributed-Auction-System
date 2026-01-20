package product

import "context"

// Repository defines the interface for product persistence
// Defined in domain layer, implemented in infrastructure layer (Dependency Inversion)
type Repository interface {
	// Save saves a product (insert or update)
	Save(ctx context.Context, product *Product) error

	// FindByID finds a product by ID
	FindByID(ctx context.Context, id ProductID) (*Product, error)

	// FindBySeller finds all products by seller with pagination
	FindBySeller(ctx context.Context, sellerID SellerID, limit, offset int) ([]*Product, error)

	// FindByStatus finds products by status with pagination
	FindByStatus(ctx context.Context, status ProductStatus, limit, offset int) ([]*Product, error)

	// FindActiveProducts finds all active products
	FindAllActiveProducts(ctx context.Context) ([]*Product, error)

	// Delete deletes a product
	Delete(ctx context.Context, id ProductID) error

	// Exists checks if a product exists
	Exists(ctx context.Context, id ProductID) (bool, error)

	// CountBySeller counts products by seller
	CountBySeller(ctx context.Context, sellerID SellerID) (int, error)

	// CountByStatus counts products by status
	CountByStatus(ctx context.Context, status ProductStatus) (int, error)
}
