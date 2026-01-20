package service

import (
	"context"
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/domain/product"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/infrastructure/persistence/postgres"
)

// ProductService handles product use cases
type ProductService struct {
	productRepo         product.Repository
	productTxRepository *postgres.ProductTxRepository
}

// NewProductService creates a new ProductService
func NewProductService(
	productRepo product.Repository,
	productTxRepository *postgres.ProductTxRepository,
) *ProductService {
	return &ProductService{
		productRepo:         productRepo,
		productTxRepository: productTxRepository,
	}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(
	ctx context.Context,
	sellerID string,
	name string,
	description string,
	regularPrice int64,
	flashSalePrice *int64,
	currency string,
) (*product.Product, error) {
	sid, err := product.ParseSellerID(sellerID)
	if err != nil {
		return nil, fmt.Errorf("invalid seller id: %w", err)
	}

	regularMoney, err := product.NewMoney(regularPrice, currency)
	if err != nil {
		return nil, fmt.Errorf("invalid regular price: %w", err)
	}

	var pricing product.Pricing
	if flashSalePrice != nil {
		flashMoney, err := product.NewMoney(*flashSalePrice, currency)
		if err != nil {
			return nil, fmt.Errorf("invalid flash sale price: %w", err)
		}
		pricing, err = product.NewPricingWithFlashSale(regularMoney, flashMoney)
		if err != nil {
			return nil, fmt.Errorf("invalid pricing: %w", err)
		}
	} else {
		pricing, err = product.NewPricing(regularMoney)
		if err != nil {
			return nil, fmt.Errorf("invalid pricing: %w", err)
		}
	}

	p, err := product.NewProduct(sid, name, description, pricing)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Use productTxRepository (handles transaction + events)
	if err := s.productTxRepository.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to save product: %w", err)
	}

	return p, nil
}

// PublishProduct publishes a product
func (s *ProductService) PublishProduct(ctx context.Context, productID string) error {
	pid, err := product.ParseProductID(productID)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	p, err := s.productRepo.FindByID(ctx, pid)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	if err := p.Publish(); err != nil {
		return fmt.Errorf("cannot publish product: %w", err)
	}

	// Use productTxRepository (handles transaction + events)
	if err := s.productTxRepository.Save(ctx, p); err != nil {
		return fmt.Errorf("failed to save product: %w", err)
	}

	return nil
}

// DeactivateProduct deactivates a product
func (s *ProductService) DeactivateProduct(ctx context.Context, productID string) error {
	pid, err := product.ParseProductID(productID)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	p, err := s.productRepo.FindByID(ctx, pid)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	if err := p.Deactivate(); err != nil {
		return fmt.Errorf("cannot deactivate product: %w", err)
	}

	// Use productTxRepository
	if err := s.productTxRepository.Save(ctx, p); err != nil {
		return fmt.Errorf("failed to save product: %w", err)
	}

	return nil
}

// MarkProductAsSoldOut marks a product as sold out
func (s *ProductService) MarkProductAsSoldOut(ctx context.Context, productID string) error {
	pid, err := product.ParseProductID(productID)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	p, err := s.productRepo.FindByID(ctx, pid)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	if err := p.MarkAsSoldOut(); err != nil {
		return fmt.Errorf("cannot mark as sold out: %w", err)
	}

	// Use productTxRepository
	if err := s.productTxRepository.Save(ctx, p); err != nil {
		return fmt.Errorf("failed to save product: %w", err)
	}

	return nil
}

// UpdateProductInfo updates product information
func (s *ProductService) UpdateProductInfo(
	ctx context.Context,
	productID string,
	name string,
	description string,
) error {
	pid, err := product.ParseProductID(productID)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	p, err := s.productRepo.FindByID(ctx, pid)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	if err := p.UpdateInfo(name, description); err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	// Use ProductRepository (no events, no transaction needed)
	if err := s.productRepo.Save(ctx, p); err != nil {
		return fmt.Errorf("failed to save product: %w", err)
	}

	return nil
}

// UpdateProductPricing updates product pricing
func (s *ProductService) UpdateProductPricing(
	ctx context.Context,
	productID string,
	regularPrice int64,
	flashSalePrice *int64,
	currency string,
) error {
	pid, err := product.ParseProductID(productID)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	p, err := s.productRepo.FindByID(ctx, pid)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	regularMoney, err := product.NewMoney(regularPrice, currency)
	if err != nil {
		return fmt.Errorf("invalid regular price: %w", err)
	}

	var pricing product.Pricing
	if flashSalePrice != nil {
		flashMoney, err := product.NewMoney(*flashSalePrice, currency)
		if err != nil {
			return fmt.Errorf("invalid flash sale price: %w", err)
		}
		pricing, err = product.NewPricingWithFlashSale(regularMoney, flashMoney)
		if err != nil {
			return fmt.Errorf("invalid pricing: %w", err)
		}
	} else {
		pricing, err = product.NewPricing(regularMoney)
		if err != nil {
			return fmt.Errorf("invalid pricing: %w", err)
		}
	}

	if err := p.UpdatePricing(pricing); err != nil {
		return fmt.Errorf("failed to update pricing: %w", err)
	}

	// Use ProductRepository (no events)
	if err := s.productRepo.Save(ctx, p); err != nil {
		return fmt.Errorf("failed to save product: %w", err)
	}

	return nil
}

// DeleteProduct deletes a product
func (s *ProductService) DeleteProduct(ctx context.Context, productID string, sellerID string) error {
	pid, err := product.ParseProductID(productID)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	sid, err := product.ParseSellerID(sellerID)
	if err != nil {
		return fmt.Errorf("invalid seller id: %w", err)
	}

	p, err := s.productRepo.FindByID(ctx, pid)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	if !p.CanBeDeletedBy(sid) {
		return product.ErrUnauthorizedDelete
	}

	if err := s.productRepo.Delete(ctx, pid); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, productID string) (*product.Product, error) {
	pid, err := product.ParseProductID(productID)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}

	p, err := s.productRepo.FindByID(ctx, pid)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	return p, nil
}

// GetProductsBySeller retrieves products by seller
func (s *ProductService) GetProductsBySeller(
	ctx context.Context,
	sellerID string,
	limit, offset int,
) ([]*product.Product, error) {
	sid, err := product.ParseSellerID(sellerID)
	if err != nil {
		return nil, fmt.Errorf("invalid seller id: %w", err)
	}

	products, err := s.productRepo.FindBySeller(ctx, sid, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}

	return products, nil
}

// GetActiveProducts retrieves all active products
func (s *ProductService) GetActiveProducts(
	ctx context.Context,
	limit, offset int,
) ([]*product.Product, error) {
	products, err := s.productRepo.FindByStatus(ctx, product.ProductStatusActive, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}

	return products, nil
}
