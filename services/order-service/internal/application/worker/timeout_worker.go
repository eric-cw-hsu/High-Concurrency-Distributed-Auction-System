package worker

import (
	"context"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/infrastructure/persistence/redis"
	"go.uber.org/zap"
)

type OrderCanceller interface {
	CancelExpiredOrder(ctx context.Context, orderID string) error
}

// OrderTimeoutWorker monitors Redis sorted set for expired orders
type OrderTimeoutWorker struct {
	orderCanceller OrderCanceller
	timeoutQueue   *redis.TimeoutQueue
	config         *config.OrderTimeoutWorkerConfig
}

// NewOrderTimeoutWorker creates a new order timeout worker
func NewOrderTimeoutWorker(
	orderCanceller OrderCanceller,
	timeoutQueue *redis.TimeoutQueue,
	config *config.OrderTimeoutWorkerConfig,
) *OrderTimeoutWorker {
	return &OrderTimeoutWorker{
		orderCanceller: orderCanceller,
		timeoutQueue:   timeoutQueue,
		config:         config,
	}
}

// Start starts the timeout worker
func (w *OrderTimeoutWorker) Start(ctx context.Context) error {
	zap.L().Info("starting order timeout worker",
		zap.Duration("check_interval", w.config.CheckInterval),
	)

	ticker := time.NewTicker(w.config.CheckInterval)
	defer ticker.Stop()

	// Run once immediately
	if err := w.processExpiredOrders(ctx); err != nil {
		zap.L().Error("initial timeout check failed", zap.Error(err))
	}

	for {
		select {
		case <-ticker.C:
			if err := w.processExpiredOrders(ctx); err != nil {
				zap.L().Error("timeout check failed", zap.Error(err))
			}

		case <-ctx.Done():
			zap.L().Info("order timeout worker stopping")
			return nil
		}
	}
}

// processExpiredOrders processes all expired orders
func (w *OrderTimeoutWorker) processExpiredOrders(ctx context.Context) error {
	now := time.Now()

	zap.L().Debug("checking for expired orders",
		zap.Time("now", now),
	)

	// Get expired order IDs from Redis sorted set
	orderIDs, err := w.timeoutQueue.GetExpired(ctx, now)
	if err != nil {
		zap.L().Error("failed to get expired orders from queue",
			zap.Error(err),
		)
		return err
	}

	if len(orderIDs) == 0 {
		zap.L().Debug("no expired orders found")
		return nil
	}

	zap.L().Info("found expired orders in timeout queue",
		zap.Int("count", len(orderIDs)),
	)

	processedIDs := make([]string, 0, len(orderIDs))
	successCount := 0
	failCount := 0

	for _, orderID := range orderIDs {
		// Cancel expired order
		err := w.orderCanceller.CancelExpiredOrder(ctx, orderID)
		if err != nil {
			zap.L().Error("failed to cancel expired order",
				zap.String("order_id", orderID),
				zap.Error(err),
			)
			failCount++

			// Still mark as processed to avoid retry storm
			// Background scanner will catch it if needed
			processedIDs = append(processedIDs, orderID)
			continue
		}

		processedIDs = append(processedIDs, orderID)
		successCount++
	}

	// Remove processed orders from queue
	if len(processedIDs) > 0 {
		if err := w.timeoutQueue.RemoveBatch(ctx, processedIDs); err != nil {
			zap.L().Error("failed to remove processed orders from queue",
				zap.Int("count", len(processedIDs)),
				zap.Error(err),
			)
		}
	}

	zap.L().Info("expired orders processed",
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
		zap.Int("total", len(orderIDs)),
	)

	return nil
}
