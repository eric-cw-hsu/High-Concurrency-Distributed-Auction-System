package product

import "context"

type ProductRepository interface {
	GetAllProducts(ctx context.Context) ([]Product, error)
	SearchProductsByName(ctx context.Context, name string) ([]Product, error)
	GetProductByID(ctx context.Context, id string) (*Product, error)
	SaveProduct(ctx context.Context, product *Product) (*Product, error)
	UpdateProduct(ctx context.Context, product *Product) (*Product, error)
	DeleteProduct(ctx context.Context, id string) error
}
