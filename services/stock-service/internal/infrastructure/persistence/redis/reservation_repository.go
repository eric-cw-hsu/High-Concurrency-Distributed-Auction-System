package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ReservationRepository implements reservation persistence in Redis
type ReservationRepository struct {
	client *redis.Client
}

var _ reservation.CacheRepository = (*ReservationRepository)(nil)

// NewReservationRepository creates a new ReservationRepository
func NewReservationRepository(client *redis.Client) *ReservationRepository {
	return &ReservationRepository{client: client}
}

// Save saves reservation to Redis with TTL
// Note: This is usually called by Lua script in StockRepository.ReserveWithReservation
// This method is for manual saves if needed
func (r *ReservationRepository) Save(ctx context.Context, res *reservation.Reservation) error {
	key := reservationKey(res.ID())

	logger.DebugContext(ctx, "saving reservation to redis",
		zap.String("reservation_id", res.ID().String()),
		zap.String("product_id", res.ProductID().String()),
	)

	data := map[string]interface{}{
		"id":          res.ID().String(),
		"product_id":  res.ProductID().String(),
		"user_id":     res.UserID().String(),
		"quantity":    res.Quantity(),
		"status":      string(res.Status()),
		"reserved_at": res.ReservedAt().Format(time.RFC3339),
		"expired_at":  res.ExpiredAt().Format(time.RFC3339),
	}

	if orderID := res.OrderID(); orderID != nil {
		data["order_id"] = *orderID
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.ErrorContext(ctx, "failed to marshal reservation",
			zap.String("reservation_id", res.ID().String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to marshal reservation: %w", err)
	}

	ttl := time.Until(res.ExpiredAt())
	if ttl <= 0 {
		logger.WarnContext(ctx, "reservation already expired",
			zap.String("reservation_id", res.ID().String()),
			zap.Time("expired_at", res.ExpiredAt()),
		)
		return reservation.ErrReservationExpired
	}

	if err := r.client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		logger.ErrorContext(ctx, "failed to save reservation to redis",
			zap.String("reservation_id", res.ID().String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to save reservation: %w", err)
	}

	logger.DebugContext(ctx, "reservation saved to redis",
		zap.String("reservation_id", res.ID().String()),
		zap.Duration("ttl", ttl),
	)

	return nil
}

// FindByID finds reservation by ID from Redis
func (r *ReservationRepository) FindByID(ctx context.Context, id reservation.ReservationID) (*reservation.Reservation, error) {
	key := reservationKey(id)

	logger.DebugContext(ctx, "finding reservation in redis",
		zap.String("reservation_id", id.String()),
	)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.DebugContext(ctx, "reservation not found in redis",
				zap.String("reservation_id", id.String()),
			)
			return nil, reservation.ErrReservationNotFound
		}

		logger.ErrorContext(ctx, "failed to get reservation from redis",
			zap.String("reservation_id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to find reservation: %w", err)
	}

	var resData map[string]interface{}
	if err := json.Unmarshal(data, &resData); err != nil {
		logger.ErrorContext(ctx, "failed to unmarshal reservation",
			zap.String("reservation_id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to unmarshal reservation: %w", err)
	}

	// Parse fields
	productID, _ := reservation.ParseProductID(resData["product_id"].(string))
	userID, _ := reservation.ParseUserID(resData["user_id"].(string))
	quantity := int(resData["quantity"].(float64))
	status := reservation.ReservationStatus(resData["status"].(string))

	reservedAt, _ := time.Parse(time.RFC3339, resData["reserved_at"].(string))
	expiredAt, _ := time.Parse(time.RFC3339, resData["expired_at"].(string))

	var orderID *string
	if oid, ok := resData["order_id"].(string); ok {
		orderID = &oid
	}

	res := reservation.ReconstructReservation(
		id,
		productID,
		userID,
		quantity,
		status,
		reservedAt,
		expiredAt,
		nil, nil,
		orderID,
	)

	return res, nil
}

// Delete deletes reservation from Redis
func (r *ReservationRepository) Delete(ctx context.Context, id reservation.ReservationID) error {
	key := reservationKey(id)

	logger.DebugContext(ctx, "deleting reservation from redis",
		zap.String("reservation_id", id.String()),
	)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		logger.ErrorContext(ctx, "failed to delete reservation from redis",
			zap.String("reservation_id", id.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete reservation: %w", err)
	}

	return nil
}

// FindActiveByProductID finds all active reservations for a product
func (r *ReservationRepository) FindActiveByProductID(
	ctx context.Context,
	productID reservation.ProductID,
) ([]*reservation.Reservation, error) {
	pattern := "reservation:*"

	logger.DebugContext(ctx, "scanning active reservations",
		zap.String("product_id", productID.String()),
	)

	var cursor uint64
	var reservations []*reservation.Reservation

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			logger.ErrorContext(ctx, "failed to scan reservations",
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to scan reservations: %w", err)
		}

		for _, key := range keys {
			data, err := r.client.Get(ctx, key).Bytes()
			if err != nil {
				continue
			}

			var resData map[string]interface{}
			if err := json.Unmarshal(data, &resData); err != nil {
				continue
			}

			// Filter by product_id
			if resData["product_id"].(string) != productID.String() {
				continue
			}

			// Parse reservation
			resID, _ := reservation.ParseReservationID(resData["id"].(string))
			res, err := r.FindByID(ctx, resID)
			if err != nil {
				continue
			}

			if res.IsActive() {
				reservations = append(reservations, res)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	logger.DebugContext(ctx, "active reservations found",
		zap.String("product_id", productID.String()),
		zap.Int("count", len(reservations)),
	)

	return reservations, nil
}
