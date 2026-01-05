package clients

import (
	"context"

	productv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/product/v1"

	"google.golang.org/grpc"
)

// ProductClient wraps Product Service gRPC client
type ProductClient struct {
	client productv1.ProductServiceClient
}

// NewProductClient creates a new ProductClient
func NewProductClient(conn *grpc.ClientConn) *ProductClient {
	return &ProductClient{
		client: productv1.NewProductServiceClient(conn),
	}
}

// CreateProduct creates a new product
func (c *ProductClient) CreateProduct(
	ctx context.Context,
	req *productv1.CreateProductRequest,
) (*productv1.CreateProductResponse, error) {
	return c.client.CreateProduct(ctx, req)
}

// UpdateProductInfo updates product information
func (c *ProductClient) UpdateProductInfo(
	ctx context.Context,
	req *productv1.UpdateProductInfoRequest,
) (*productv1.UpdateProductInfoResponse, error) {
	return c.client.UpdateProductInfo(ctx, req)
}

// UpdateProductPricing updates product pricing
func (c *ProductClient) UpdateProductPricing(
	ctx context.Context,
	req *productv1.UpdateProductPricingRequest,
) (*productv1.UpdateProductPricingResponse, error) {
	return c.client.UpdateProductPricing(ctx, req)
}

// PublishProduct publishes a product
func (c *ProductClient) PublishProduct(
	ctx context.Context,
	req *productv1.PublishProductRequest,
) (*productv1.PublishProductResponse, error) {
	return c.client.PublishProduct(ctx, req)
}

// DeactivateProduct deactivates a product
func (c *ProductClient) DeactivateProduct(
	ctx context.Context,
	req *productv1.DeactivateProductRequest,
) (*productv1.DeactivateProductResponse, error) {
	return c.client.DeactivateProduct(ctx, req)
}

// DeleteProduct deletes a product
func (c *ProductClient) DeleteProduct(
	ctx context.Context,
	req *productv1.DeleteProductRequest,
) (*productv1.DeleteProductResponse, error) {
	return c.client.DeleteProduct(ctx, req)
}

// GetProduct gets a product by ID
func (c *ProductClient) GetProduct(
	ctx context.Context,
	req *productv1.GetProductRequest,
) (*productv1.GetProductResponse, error) {
	return c.client.GetProduct(ctx, req)
}

// GetProductsBySeller gets products by seller
func (c *ProductClient) GetProductsBySeller(
	ctx context.Context,
	req *productv1.GetProductsBySellerRequest,
) (*productv1.GetProductsBySellerResponse, error) {
	return c.client.GetProductsBySeller(ctx, req)
}

// GetActiveProducts gets all active products
func (c *ProductClient) GetActiveProducts(
	ctx context.Context,
	req *productv1.GetActiveProductsRequest,
) (*productv1.GetActiveProductsResponse, error) {
	return c.client.GetActiveProducts(ctx, req)
}
