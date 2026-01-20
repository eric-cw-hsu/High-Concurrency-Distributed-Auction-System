package productprice

import "context"

type ProductClient interface {
	FetchProductDetail(ctx context.Context, productID string) (*ProductPrice, error)
}
