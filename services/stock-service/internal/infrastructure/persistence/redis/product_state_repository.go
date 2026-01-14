package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const (
	activeProductsKey = "stock_service:active_products"
)

// ProductStateRepository manages product state in Redis
type ProductStateRepository struct {
	client *redis.Client
}

// NewProductStateRepository creates a new product state repository
func NewProductStateRepository(client *redis.Client) *ProductStateRepository {
	return &ProductStateRepository{
		client: client,
	}
}

// IsActive checks if a product is active
func (r *ProductStateRepository) IsActive(ctx context.Context, productID string) (bool, error) {
	return r.client.SIsMember(ctx, activeProductsKey, productID).Result()
}

// MarkActive marks a product as active
func (r *ProductStateRepository) MarkActive(ctx context.Context, productID string) error {
	return r.client.SAdd(ctx, activeProductsKey, productID).Err()
}

// MarkInactive marks a product as inactive
func (r *ProductStateRepository) MarkInactive(ctx context.Context, productID string) error {
	return r.client.SRem(ctx, activeProductsKey, productID).Err()
}

// Remove removes a product from the active set
func (r *ProductStateRepository) Remove(ctx context.Context, productID string) error {
	return r.MarkInactive(ctx, productID)
}

// GetAllActive returns all active product IDs
func (r *ProductStateRepository) GetAllActive(ctx context.Context) ([]string, error) {
	return r.client.SMembers(ctx, activeProductsKey).Result()
}

// Count returns the number of active products
func (r *ProductStateRepository) Count(ctx context.Context) (int64, error) {
	return r.client.SCard(ctx, activeProductsKey).Result()
}
