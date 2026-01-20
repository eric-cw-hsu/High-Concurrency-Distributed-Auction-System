package service

import (
	"context"

	productprice "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/product_price"
)

type ProductAppService struct {
	priceRepo productprice.Repository
}

var _ productprice.ProductPriceSyncer = (*ProductAppService)(nil)

func NewProductAppService(repo productprice.Repository) *ProductAppService {
	return &ProductAppService{priceRepo: repo}
}

// SyncProductPrice encapsulate
func (s *ProductAppService) SyncProductPrice(ctx context.Context, productID string, price int64, currency string) error {
	p := productprice.NewProductPrice(productID, price, currency)
	return s.priceRepo.Upsert(ctx, p)
}
