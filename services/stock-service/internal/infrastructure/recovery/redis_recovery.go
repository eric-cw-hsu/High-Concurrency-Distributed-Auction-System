package recovery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"github.com/segmentio/kafka-go"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisRecovery handles Redis failure recovery
type RedisRecovery struct {
	redisClient               *redis.Client
	persistentReservationRepo reservation.PersistentRepository
	cacheReservationRepo      reservation.CacheRepository
}

// NewRedisRecovery creates a new RedisRecovery
func NewRedisRecovery(
	redisClient *redis.Client,
	persistentReservationRepo reservation.PersistentRepository,
	cacheReservationRepo reservation.CacheRepository,
) *RedisRecovery {
	return &RedisRecovery{
		redisClient:               redisClient,
		persistentReservationRepo: persistentReservationRepo,
		cacheReservationRepo:      cacheReservationRepo,
	}
}

// RecoverActiveReservations recovers active reservations from PostgreSQL to Redis
func (r *RedisRecovery) RecoverActiveReservations(ctx context.Context) error {
	zap.L().Info("starting recovery of active reservations from postgresql")

	reservations, err := r.persistentReservationRepo.FindAllActive(ctx)
	if err != nil {
		zap.L().Error("failed to query active reservations",
			zap.Error(err),
		)
		return fmt.Errorf("failed to query reservations: %w", err)
	}

	zap.L().Info("active reservations found in postgresql",
		zap.Int("count", len(reservations)),
	)

	// Restore to Redis
	successCount := 0
	for _, res := range reservations {
		// Save to Redis with remaining TTL
		if err := r.cacheReservationRepo.Save(ctx, res); err != nil {
			zap.L().Error("failed to restore reservation to redis",
				zap.String("reservation_id", res.ID().String()),
				zap.Error(err),
			)
			continue
		}

		successCount++
	}

	zap.L().Info("active reservations recovered",
		zap.Int("success", successCount),
		zap.Int("total", len(reservations)),
	)

	return nil
}

// RecoverStockFromKafka recovers stock by replaying Kafka events
func (r *RedisRecovery) RecoverStockFromKafka(
	ctx context.Context,
	kafkaBrokers []string,
	topic string,
	lookbackDuration time.Duration,
) error {
	zap.L().Info("recovering stock from kafka",
		zap.String("topic", topic),
		zap.Duration("lookback", lookbackDuration),
	)

	// Create Kafka reader from earliest offset
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: kafkaBrokers,
		Topic:   topic,
		GroupID: "stock-recovery-" + time.Now().Format("20060102150405"),
		// Start from beginning
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	// Track stock changes per product
	stockChanges := make(map[string]int) // product_id -> net change

	startTime := time.Now().Add(-lookbackDuration)
	eventsProcessed := 0

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				break
			}
			zap.L().Error("failed to read kafka message",
				zap.Error(err),
			)
			continue
		}

		// Parse event
		var event map[string]interface{}
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			continue
		}

		// Check event time
		occurredAt, _ := time.Parse(time.RFC3339, event["occurred_at"].(string))
		if occurredAt.Before(startTime) {
			continue // Skip old events
		}

		eventType := event["event_type"].(string)
		productID := event["data"].(map[string]interface{})["product_id"].(string)

		// Apply event
		switch eventType {
		case "stock.set":
			quantity := int(event["data"].(map[string]interface{})["quantity"].(float64))
			stockChanges[productID] = quantity

		case "stock.reserved":
			quantity := int(event["data"].(map[string]interface{})["quantity"].(float64))
			stockChanges[productID] -= quantity

		case "stock.released":
			quantity := int(event["data"].(map[string]interface{})["quantity"].(float64))
			stockChanges[productID] += quantity
		}

		eventsProcessed++
	}

	// Apply changes to Redis
	zap.L().Info("applying stock changes to redis",
		zap.Int("products", len(stockChanges)),
		zap.Int("events_processed", eventsProcessed),
	)

	for productID, quantity := range stockChanges {
		key := fmt.Sprintf("stock:product:%s", productID)
		if err := r.redisClient.Set(ctx, key, quantity, 0).Err(); err != nil {
			zap.L().Error("failed to set stock during recovery",
				zap.String("product_id", productID),
				zap.Error(err),
			)
			continue
		}
	}

	zap.L().Info("stock recovery from kafka completed",
		zap.Int("products_recovered", len(stockChanges)),
	)

	return nil
}

// FullRecovery performs full recovery (reservations + stock)
func (r *RedisRecovery) FullRecovery(ctx context.Context) error {
	zap.L().Info("starting full redis recovery")

	// Step 1: Recover reservations from PostgreSQL
	if err := r.RecoverActiveReservations(ctx); err != nil {
		zap.L().Error("failed to recover reservations",
			zap.Error(err),
		)
		return err
	}

	// Step 2: Stock recovery requires manual intervention or Kafka replay
	// zap.L().Warn("stock quantities need manual recovery",
	// 	zap.String("action", "use admin API to set stock or replay kafka events"),
	// )

	zap.L().Info("redis recovery completed")

	return nil
}

// CheckRedisHealth checks if Redis needs recovery
func (r *RedisRecovery) CheckRedisHealth(ctx context.Context) (bool, error) {
	// Check if Redis has any stock keys
	cursor := uint64(0)
	keys, _, err := r.redisClient.Scan(ctx, cursor, "stock:product:*", 10).Result()
	if err != nil {
		return false, err
	}

	// If no stock keys found, might need recovery
	needsRecovery := len(keys) == 0

	zap.L().Info("redis health check",
		zap.Bool("needs_recovery", needsRecovery),
		zap.Int("stock_keys_found", len(keys)),
	)

	return needsRecovery, nil
}
