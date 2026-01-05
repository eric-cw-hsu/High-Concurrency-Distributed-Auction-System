package postgres

import (
	"database/sql"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/domain/product"
)

// DomainToModel converts domain Product to database ProductModel
func DomainToModel(p *product.Product) *ProductModel {
	model := &ProductModel{
		ID:           p.ID().String(),
		SellerID:     p.SellerID().String(),
		Name:         p.Name(),
		Description:  p.Description(),
		RegularPrice: p.Pricing().RegularPrice().Amount(),
		Currency:     p.Pricing().RegularPrice().Currency(),
		Status:       string(p.Status()),
		StockStatus:  string(p.StockStatus()),
		CreatedAt:    p.CreatedAt(),
		UpdatedAt:    p.UpdatedAt(),
	}

	// Handle optional flash sale price
	if p.Pricing().HasFlashSale() {
		flashPrice := p.Pricing().FlashSalePrice()
		model.FlashSalePrice = sql.NullInt64{
			Int64: flashPrice.Amount(),
			Valid: true,
		}
	}

	return model
}

// ModelToDomain converts database ProductModel to domain Product
func ModelToDomain(model *ProductModel) (*product.Product, error) {
	// Parse IDs
	productID, err := product.ParseProductID(model.ID)
	if err != nil {
		return nil, err
	}

	sellerID, err := product.ParseSellerID(model.SellerID)
	if err != nil {
		return nil, err
	}

	// Create pricing
	regularPrice, err := product.NewMoney(model.RegularPrice, model.Currency)
	if err != nil {
		return nil, err
	}

	var pricing product.Pricing
	if model.FlashSalePrice.Valid {
		flashPrice, err := product.NewMoney(model.FlashSalePrice.Int64, model.Currency)
		if err != nil {
			return nil, err
		}
		pricing, err = product.NewPricingWithFlashSale(regularPrice, flashPrice)
		if err != nil {
			return nil, err
		}
	} else {
		pricing, err = product.NewPricing(regularPrice)
		if err != nil {
			return nil, err
		}
	}

	// Reconstruct product
	return product.ReconstructProduct(
		productID,
		sellerID,
		model.Name,
		model.Description,
		pricing,
		product.ProductStatus(model.Status),
		product.StockStatus(model.StockStatus),
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}
