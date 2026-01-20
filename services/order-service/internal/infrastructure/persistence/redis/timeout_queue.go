package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/order"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	timeoutQueueKey = "order:timeouts"
)

// TimeoutQueue manages order timeout using Redis Sorted Set
// Score = expiration timestamp (Unix seconds)
// Member = order_id
type TimeoutQueue struct {
	client *redis.Client
}

// NewTimeoutQueue creates a new timeout queue
func NewTimeoutQueue(client *redis.Client) *TimeoutQueue {
	return &TimeoutQueue{
		client: client,
	}
}

// Add adds an order to the timeout queue
func (q *TimeoutQueue) Add(ctx context.Context, orderID order.OrderID, expiresAt time.Time) error {
	score := float64(expiresAt.Unix())
	member := orderID.String()

	zap.L().Debug("adding order to timeout queue",
		zap.String("order_id", member),
		zap.Time("expires_at", expiresAt),
		zap.Float64("score", score),
	)

	err := q.client.ZAdd(ctx, timeoutQueueKey, redis.Z{
		Score:  score,
		Member: member,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to add to timeout queue: %w", err)
	}

	return nil
}

// Remove removes an order from the timeout queue
func (q *TimeoutQueue) Remove(ctx context.Context, orderID order.OrderID) error {
	member := orderID.String()

	zap.L().Debug("removing order from timeout queue",
		zap.String("order_id", member),
	)

	err := q.client.ZRem(ctx, timeoutQueueKey, member).Err()
	if err != nil {
		return fmt.Errorf("failed to remove from timeout queue: %w", err)
	}

	return nil
}

// GetExpired gets all expired orders
func (q *TimeoutQueue) GetExpired(ctx context.Context, now time.Time) ([]string, error) {
	maxScore := float64(now.Unix())

	zap.L().Debug("querying expired orders",
		zap.Time("now", now),
		zap.Float64("max_score", maxScore),
	)

	// Get all orders with score <= now (expired)
	orderIDs, err := q.client.ZRangeByScore(ctx, timeoutQueueKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", maxScore),
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to get expired orders: %w", err)
	}

	zap.L().Debug("expired orders found",
		zap.Int("count", len(orderIDs)),
	)

	return orderIDs, nil
}

// RemoveBatch removes multiple orders from the timeout queue
func (q *TimeoutQueue) RemoveBatch(ctx context.Context, orderIDs []string) error {
	if len(orderIDs) == 0 {
		return nil
	}

	zap.L().Debug("removing batch from timeout queue",
		zap.Int("count", len(orderIDs)),
	)

	// Convert to []interface{} for ZRem
	members := make([]interface{}, len(orderIDs))
	for i, id := range orderIDs {
		members[i] = id
	}

	err := q.client.ZRem(ctx, timeoutQueueKey, members...).Err()
	if err != nil {
		return fmt.Errorf("failed to remove batch from timeout queue: %w", err)
	}

	return nil
}

// Count returns the total number of orders in the queue
func (q *TimeoutQueue) Count(ctx context.Context) (int64, error) {
	count, err := q.client.ZCard(ctx, timeoutQueueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count timeout queue: %w", err)
	}
	return count, nil
}

// GetPending gets pending orders (not yet expired)
func (q *TimeoutQueue) GetPending(ctx context.Context, now time.Time, limit int) ([]string, error) {
	minScore := float64(now.Unix())

	// Get orders with score > now (still pending)
	orderIDs, err := q.client.ZRangeByScore(ctx, timeoutQueueKey, &redis.ZRangeBy{
		Min:   fmt.Sprintf("(%f", minScore), // Exclusive
		Max:   "+inf",
		Count: int64(limit),
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}

	return orderIDs, nil
}
