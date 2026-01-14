package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/stock"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// StockRepository implements stock.Repository using Redis
type StockRepository struct {
	client *redis.Client
}

// NewStockRepository creates a new StockRepository
func NewStockRepository(client *redis.Client) *StockRepository {
	return &StockRepository{client: client}
}

// Save saves stock to Redis
func (r *StockRepository) Save(ctx context.Context, s *stock.Stock) error {
	key := stockKey(s.ProductID())
	metaKey := stockMetadataKey(s.ProductID())

	logger.DebugContext(ctx, "saving stock to redis",
		zap.String("product_id", s.ProductID().String()),
		zap.Int("quantity", s.Quantity()),
	)

	// Save quantity
	if err := r.client.Set(ctx, key, s.Quantity(), 0).Err(); err != nil {
		logger.ErrorContext(ctx, "failed to save stock quantity",
			zap.String("product_id", s.ProductID().String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to save stock: %w", err)
	}

	// Save metadata (initial quantity for low stock calculation)
	metadata := map[string]interface{}{
		"initial_quantity": s.InitialQuantity(),
		"updated_at":       s.UpdatedAt().Unix(),
	}
	metaJSON, _ := json.Marshal(metadata)

	if err := r.client.Set(ctx, metaKey, metaJSON, 0).Err(); err != nil {
		logger.WarnContext(ctx, "failed to save stock metadata",
			zap.String("product_id", s.ProductID().String()),
			zap.Error(err),
		)
		// Don't fail if metadata save fails
	}

	logger.DebugContext(ctx, "stock saved successfully",
		zap.String("product_id", s.ProductID().String()),
	)

	return nil
}

// FindByProductID finds stock by product ID
func (r *StockRepository) FindByProductID(ctx context.Context, productID stock.ProductID) (*stock.Stock, error) {
	key := stockKey(productID)
	metaKey := stockMetadataKey(productID)

	logger.DebugContext(ctx, "finding stock in redis",
		zap.String("product_id", productID.String()),
	)

	// Get quantity
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			logger.DebugContext(ctx, "stock not found in redis",
				zap.String("product_id", productID.String()),
			)
			return nil, stock.ErrStockNotFound
		}

		logger.ErrorContext(ctx, "failed to get stock from redis",
			zap.String("product_id", productID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to find stock: %w", err)
	}

	quantity, err := strconv.Atoi(val)
	if err != nil {
		logger.ErrorContext(ctx, "invalid stock quantity in redis",
			zap.String("product_id", productID.String()),
			zap.String("value", val),
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid stock quantity: %w", err)
	}

	// Get metadata (initial quantity)
	initialQuantity := quantity // Default to current if metadata not found
	if metaData, err := r.client.Get(ctx, metaKey).Bytes(); err == nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal(metaData, &metadata); err == nil {
			if initial, ok := metadata["initial_quantity"].(float64); ok {
				initialQuantity = int(initial)
			}
		}
	}

	s := stock.ReconstructStock(productID, quantity, initialQuantity, time.Now())

	return s, nil
}

// Exists checks if stock exists
func (r *StockRepository) Exists(ctx context.Context, productID stock.ProductID) (bool, error) {
	key := stockKey(productID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		logger.ErrorContext(ctx, "failed to check stock existence",
			zap.String("product_id", productID.String()),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return exists > 0, nil
}

// Reserve reserves stock
func (r *StockRepository) Reserve(
	ctx context.Context,
	productID stock.ProductID,
	quantity int,
) (int, error) {
	// This is simplified version, full version with reservation is in ReserveWithReservation
	if quantity <= 0 {
		return 0, stock.ErrInvalidQuantity
	}
	if quantity > stock.MaxDeductQuantity {
		return 0, stock.ErrExceedsMaxQuantity
	}

	stockKey := stockKey(productID)

	logger.InfoContext(ctx, "reserving stock",
		zap.String("product_id", productID.String()),
		zap.Int("quantity", quantity),
	)

	result, err := r.client.DecrBy(ctx, stockKey, int64(quantity)).Result()
	if err != nil {
		logger.ErrorContext(ctx, "failed to deduct stock",
			zap.String("product_id", productID.String()),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to reserve stock: %w", err)
	}

	if result < 0 {
		// Rollback
		r.client.IncrBy(ctx, stockKey, int64(quantity))
		logger.WarnContext(ctx, "insufficient stock",
			zap.String("product_id", productID.String()),
			zap.Int("requested", quantity),
		)
		return 0, stock.ErrInsufficientStock
	}

	logger.InfoContext(ctx, "stock reserved successfully",
		zap.String("product_id", productID.String()),
		zap.Int("quantity", quantity),
		zap.Int64("remaining", result),
	)

	return int(result), nil
}

// Release releases reserved stock
func (r *StockRepository) Release(
	ctx context.Context,
	productID stock.ProductID,
	quantity int,
) (int, error) {
	stockKey := stockKey(productID)

	logger.InfoContext(ctx, "releasing stock",
		zap.String("product_id", productID.String()),
		zap.Int("quantity", quantity),
	)

	result, err := r.client.IncrBy(ctx, stockKey, int64(quantity)).Result()
	if err != nil {
		logger.ErrorContext(ctx, "failed to release stock",
			zap.String("product_id", productID.String()),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to release stock: %w", err)
	}

	logger.InfoContext(ctx, "stock released successfully",
		zap.String("product_id", productID.String()),
		zap.Int("quantity", quantity),
		zap.Int64("new_quantity", result),
	)

	return int(result), nil
}
