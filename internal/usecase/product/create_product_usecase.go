package product

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/product"
	"github.com/samborkent/uuidv7"
)

type CreateProductUsecase struct {
	productRepository product.ProductRepository
}

func NewCreateProductUsecase(productRepository product.ProductRepository) *CreateProductUsecase {
	return &CreateProductUsecase{
		productRepository: productRepository,
	}
}

// Execute creates a new product with the provided command.
// It generates a UUID for the product, creates the product entity, and saves it to the repository.
func (uc *CreateProductUsecase) Execute(ctx context.Context, command product.CreateProductCommand) (*product.Product, error) {
	productEntity := product.Product{
		ID:          uuidv7.New().String(),
		Name:        command.Name,
		Description: command.Description,
	}

	savedProduct, err := uc.productRepository.SaveProduct(ctx, &productEntity)
	if err != nil {
		return nil, err
	}

	return savedProduct, nil
}
