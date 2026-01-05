package grpc

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/domain/product"
	productv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/product/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// DomainToProto converts domain Product to proto Product
func domainToProto(p *product.Product) *productv1.Product {
	pricing := &productv1.Pricing{
		RegularPrice: &productv1.Money{
			Amount:   p.Pricing().RegularPrice().Amount(),
			Currency: p.Pricing().RegularPrice().Currency(),
		},
	}

	if p.Pricing().HasFlashSale() {
		flashPrice := p.Pricing().FlashSalePrice()
		pricing.FlashSalePrice = &productv1.Money{
			Amount:   flashPrice.Amount(),
			Currency: flashPrice.Currency(),
		}
	}

	return &productv1.Product{
		Id:          p.ID().String(),
		SellerId:    p.SellerID().String(),
		Name:        p.Name(),
		Description: p.Description(),
		Pricing:     pricing,
		Status:      string(p.Status()),
		StockStatus: string(p.StockStatus()),
		CreatedAt:   timestamppb.New(p.CreatedAt()),
		UpdatedAt:   timestamppb.New(p.UpdatedAt()),
	}
}
