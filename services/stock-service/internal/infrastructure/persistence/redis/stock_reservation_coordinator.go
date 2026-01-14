package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/stock"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// StockReservationCoordinator handles atomic operations across Stock and Reservation aggregates.
// This is an infrastructure-level component that exists purely to satisfy Redis technical
// constraints (Lua script atomicity), not a domain service.
// It coordinates operations that must be atomic due to business requirements (prevent overselling).
type StockReservationCoordinator struct {
	client *redis.Client
}

// NewStockReservationCoordinator creates a new coordinator
func NewStockReservationCoordinator(client *redis.Client) *StockReservationCoordinator {
	return &StockReservationCoordinator{
		client: client,
	}
}

// Reserve reserves stock and creates reservation atomically using Lua script
func (c *StockReservationCoordinator) Reserve(
	ctx context.Context,
	productID stock.ProductID,
	res *reservation.Reservation,
) (int, error) {
	sKey := stockKey(productID)
	rKey := reservationKey(res.ID())

	logger.InfoContext(ctx, "reserving stock with lua script",
		zap.String("product_id", productID.String()),
		zap.String("reservation_id", res.ID().String()),
		zap.Int("quantity", res.Quantity()),
	)

	// Serialize reservation data
	resData := map[string]interface{}{
		"id":          res.ID().String(),
		"product_id":  res.ProductID().String(),
		"user_id":     res.UserID().String(),
		"quantity":    res.Quantity(),
		"status":      string(res.Status()),
		"reserved_at": res.ReservedAt().Format(time.RFC3339),
		"expired_at":  res.ExpiredAt().Format(time.RFC3339),
	}

	resJSON, err := json.Marshal(resData)
	if err != nil {
		logger.ErrorContext(ctx, "failed to marshal reservation data",
			zap.String("reservation_id", res.ID().String()),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to marshal reservation: %w", err)
	}

	// Calculate TTL in seconds
	ttl := int(time.Until(res.ExpiredAt()).Seconds())
	if ttl <= 0 {
		return 0, reservation.ErrReservationExpired
	}

	// Execute Lua script
	result, err := c.client.Eval(ctx, ReserveStockScript,
		[]string{sKey, rKey},
		res.Quantity(),
		string(resJSON),
		ttl,
	).Result()

	if err != nil {
		logger.ErrorContext(ctx, "lua script execution failed",
			zap.String("product_id", productID.String()),
			zap.String("reservation_id", res.ID().String()),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to execute reserve script: %w", err)
	}

	// Parse result: {success, new_quantity}
	values, ok := result.([]interface{})
	if !ok || len(values) != 2 {
		logger.ErrorContext(ctx, "invalid lua script result",
			zap.String("product_id", productID.String()),
			zap.Any("result", result),
		)
		return 0, fmt.Errorf("invalid script result")
	}

	success, ok1 := values[0].(int64)
	newQty, ok2 := values[1].(int64)
	if !ok1 || !ok2 {
		logger.ErrorContext(ctx, "failed to parse lua script result",
			zap.Any("values", values),
		)
		return 0, fmt.Errorf("failed to parse result")
	}

	// Check success
	if success == 0 {
		logger.WarnContext(ctx, "insufficient stock",
			zap.String("product_id", productID.String()),
			zap.Int("requested", res.Quantity()),
			zap.Int64("available", newQty),
		)
		return int(newQty), stock.ErrInsufficientStock
	}

	logger.InfoContext(ctx, "stock reserved successfully with lua script",
		zap.String("product_id", productID.String()),
		zap.String("reservation_id", res.ID().String()),
		zap.Int("quantity", res.Quantity()),
		zap.Int64("remaining", newQty),
	)

	return int(newQty), nil
}

// Release releases stock and deletes reservation atomically using Lua script
func (c *StockReservationCoordinator) Release(
	ctx context.Context,
	productID stock.ProductID,
	reservationID reservation.ReservationID,
	quantity int,
) (int, error) {
	sKey := stockKey(productID)
	rKey := reservationKey(reservationID)

	logger.InfoContext(ctx, "releasing stock with lua script",
		zap.String("product_id", productID.String()),
		zap.String("reservation_id", reservationID.String()),
		zap.Int("quantity", quantity),
	)

	// Execute Lua script
	result, err := c.client.Eval(ctx, ReleaseStockScript,
		[]string{sKey, rKey},
		quantity,
	).Result()

	if err != nil {
		logger.ErrorContext(ctx, "lua script execution failed",
			zap.String("product_id", productID.String()),
			zap.String("reservation_id", reservationID.String()),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to execute release script: %w", err)
	}

	// Parse result: {success, new_quantity}
	values, ok := result.([]interface{})
	if !ok || len(values) != 2 {
		logger.ErrorContext(ctx, "invalid lua script result",
			zap.Any("result", result),
		)
		return 0, fmt.Errorf("invalid script result")
	}

	success, ok1 := values[0].(int64)
	newQty, ok2 := values[1].(int64)
	if !ok1 || !ok2 {
		logger.ErrorContext(ctx, "failed to parse lua script result",
			zap.Any("values", values),
		)
		return 0, fmt.Errorf("failed to parse result")
	}

	if success == 0 {
		logger.WarnContext(ctx, "reservation not found in redis",
			zap.String("reservation_id", reservationID.String()),
		)
		return 0, reservation.ErrReservationNotFound
	}

	logger.InfoContext(ctx, "stock released successfully with lua script",
		zap.String("product_id", productID.String()),
		zap.String("reservation_id", reservationID.String()),
		zap.Int("quantity", quantity),
		zap.Int64("new_stock", newQty),
	)

	return int(newQty), nil
}
