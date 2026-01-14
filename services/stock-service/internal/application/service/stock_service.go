package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/stock"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/persistence/postgres"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/persistence/redis"
	"go.uber.org/zap"
)

// StockService handles stock use cases
type StockService struct {
	cfg                         *config.ServiceConfig
	stockRepo                   stock.Repository
	cacheReservationRepo        reservation.CacheRepository
	persistentReservationRepo   reservation.PersistentRepository
	stockReservationCoordinator *redis.StockReservationCoordinator
	outboxRepo                  *postgres.OutboxRepository
	persistQueue                chan *reservation.Reservation
	productStateRepo            *redis.ProductStateRepository
}

// NewStockService creates a new StockService
func NewStockService(
	cfg *config.ServiceConfig,
	stockRepo stock.Repository,
	cacheReservationRepo reservation.CacheRepository,
	persistentReservationRepo reservation.PersistentRepository,
	stockReservationCoordinator *redis.StockReservationCoordinator,
	outboxRepo *postgres.OutboxRepository,
	persistQueue chan *reservation.Reservation,
	productStateRepo *redis.ProductStateRepository,
) *StockService {
	s := &StockService{
		cfg:                         cfg,
		stockRepo:                   stockRepo,
		cacheReservationRepo:        cacheReservationRepo,
		persistentReservationRepo:   persistentReservationRepo,
		stockReservationCoordinator: stockReservationCoordinator,
		outboxRepo:                  outboxRepo,
		persistQueue:                persistQueue,
		productStateRepo:            productStateRepo,
	}

	return s
}

// SetStock sets initial stock for a product
func (s *StockService) SetStock(
	ctx context.Context,
	productID string,
	quantity int,
) error {
	logger.InfoContext(ctx, "setting stock",
		zap.String("product_id", productID),
		zap.Int("quantity", quantity),
	)

	pid, err := stock.ParseProductID(productID)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	stk, err := stock.NewStock(pid, quantity)
	if err != nil {
		return fmt.Errorf("failed to create stock: %w", err)
	}

	if err := s.stockRepo.Save(ctx, stk); err != nil {
		return fmt.Errorf("failed to save stock: %w", err)
	}

	logger.InfoContext(ctx, "stock set successfully",
		zap.String("product_id", productID),
		zap.Int("quantity", quantity),
	)

	return nil
}

// Reserve reserves stock for a user
func (s *StockService) Reserve(
	ctx context.Context,
	productID string,
	userID string,
	quantity int,
) (*reservation.Reservation, int, error) {
	logger.InfoContext(ctx, "reserving stock",
		zap.String("product_id", productID),
		zap.String("user_id", userID),
		zap.Int("quantity", quantity),
	)

	// Parse IDs
	reservationProductID, err := reservation.ParseProductID(productID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid reservation product id: %w", err)
	}

	uid, err := reservation.ParseUserID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user id: %w", err)
	}

	// Validate quantity
	if quantity <= 0 || quantity > reservation.MaxReservationQuantity {
		return nil, 0, reservation.ErrInvalidQuantity
	}

	// Check if product is active
	isActive, err := s.productStateRepo.IsActive(ctx, productID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to check product state",
			zap.String("product_id", productID),
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("failed to check product state: %w", err)
	}

	if !isActive {
		logger.WarnContext(ctx, "product is not active",
			zap.String("product_id", productID),
		)
		return nil, 0, errors.New("product is not active")
	}

	// Create reservation
	res, err := reservation.NewReservation(reservationProductID, uid, quantity)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Execute Redis Lua script to reserve stock atomically
	stockProductID, err := stock.ParseProductID(productID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid stock product id: %w", err)
	}
	newQty, err := s.stockReservationCoordinator.Reserve(ctx, stockProductID, res)
	if err != nil {
		logger.WarnContext(ctx, "failed to reserve stock in redis",
			zap.String("product_id", productID),
			zap.Int("quantity", quantity),
			zap.Error(err),
		)
		return nil, 0, err
	}

	logger.InfoContext(ctx, "stock reserved in redis",
		zap.String("product_id", productID),
		zap.String("reservation_id", res.ID().String()),
		zap.Int("remaining", newQty),
	)

	// Write to outbox (synchronous, ensures event will be sent)
	if err := s.publishReservedEvent(ctx, res); err != nil {
		// Outbox failed, rollback Redis
		logger.ErrorContext(ctx, "outbox insert failed, rolling back redis",
			zap.String("product_id", productID),
			zap.String("reservation_id", res.ID().String()),
			zap.Error(err),
		)

		if _, rollbackErr := s.stockReservationCoordinator.Release(ctx, stockProductID, res.ID(), quantity); rollbackErr != nil {
			logger.ErrorContext(ctx, "CRITICAL: failed to rollback redis after outbox failure",
				zap.String("product_id", productID),
				zap.String("reservation_id", res.ID().String()),
				zap.Error(err),
				zap.NamedError("rollback_error", rollbackErr),
			)
		}

		return nil, 0, fmt.Errorf("failed to publish event: %w", err)
	}

	// Async write to PostgreSQL (send to queue, don't wait)
	s.persistQueue <- res

	// Check if stock is depleted or low
	if newQty == 0 {
		if err := s.publishDepletedEvent(ctx, stockProductID); err != nil {
			logger.ErrorContext(ctx, "failed to publish depleted event",
				zap.String("product_id", productID),
				zap.Error(err),
			)
		}
	} else {
		// Check low stock (asynchronously, don't block user)
		go s.checkAndPublishLowStock(context.Background(), stockProductID, newQty)
	}

	logger.InfoContext(ctx, "stock reserved successfully",
		zap.String("product_id", productID),
		zap.String("reservation_id", res.ID().String()),
		zap.String("user_id", userID),
		zap.Int("quantity", quantity),
		zap.Int("remaining", newQty),
	)

	return res, newQty, nil
}

// Release releases a reservation
func (s *StockService) Release(
	ctx context.Context,
	reservationID string,
) (int, error) {
	logger.InfoContext(ctx, "releasing reservation",
		zap.String("reservation_id", reservationID),
	)

	rid, err := reservation.ParseReservationID(reservationID)
	if err != nil {
		return 0, fmt.Errorf("invalid reservation id: %w", err)
	}

	// Find reservation
	res, err := s.cacheReservationRepo.FindByID(ctx, rid)
	if err != nil {
		return 0, fmt.Errorf("reservation not found: %w", err)
	}

	// Release (domain logic)
	if err := res.Release(); err != nil {
		return 0, fmt.Errorf("cannot release reservation: %w", err)
	}

	var newQty int
	newQty, err = s.stockReservationCoordinator.Release(ctx, stock.ProductID(res.ProductID()), res.ID(), res.Quantity())
	if err == reservation.ErrReservationNotFound {
		// cache reservation already expired, just restore stock
		newQty, err = s.stockRepo.Release(ctx, stock.ProductID(res.ProductID()), res.Quantity())
		if err != nil {
			logger.ErrorContext(ctx, "failed to release stock in redis",
				zap.String("reservation_id", reservationID),
				zap.String("product_id", res.ProductID().String()),
				zap.Error(err),
			)
			return 0, fmt.Errorf("failed to release stock: %w", err)
		}
	} else if err != nil {
		logger.ErrorContext(ctx, "failed to release stock and delete reservation in redis",
			zap.String("reservation_id", reservationID),
			zap.String("product_id", res.ProductID().String()),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to release: %w", err)
	}

	// Update PostgreSQL status (async)
	if err := s.updateReservationStatus(ctx, rid, reservation.ReservationStatusReleased); err != nil {
		logger.ErrorContext(ctx, "failed to update reservation status",
			zap.String("reservation_id", reservationID),
			zap.Error(err),
		)
	}

	// Publish event
	if err := s.publishReleasedEvent(ctx, res); err != nil {
		logger.ErrorContext(ctx, "failed to publish released event",
			zap.String("reservation_id", reservationID),
			zap.Error(err),
		)
	}

	logger.InfoContext(ctx, "reservation released successfully",
		zap.String("reservation_id", reservationID),
		zap.String("product_id", res.ProductID().String()),
		zap.Int("quantity", res.Quantity()),
		zap.Int("new_stock", newQty),
	)

	return newQty, nil
}

// GetStock gets current stock for a product
func (s *StockService) GetStock(
	ctx context.Context,
	productID string,
) (*stock.Stock, error) {
	pid, err := stock.ParseProductID(productID)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}

	stk, err := s.stockRepo.FindByProductID(ctx, pid)
	if err != nil {
		return nil, fmt.Errorf("stock not found: %w", err)
	}

	return stk, nil
}

// GetReservation gets reservation by ID
func (s *StockService) GetReservation(
	ctx context.Context,
	reservationID string,
) (*reservation.Reservation, error) {
	rid, err := reservation.ParseReservationID(reservationID)
	if err != nil {
		return nil, fmt.Errorf("invalid reservation id: %w", err)
	}

	res, err := s.cacheReservationRepo.FindByID(ctx, rid)
	if err != nil {
		return nil, fmt.Errorf("reservation not found: %w", err)
	}

	return res, nil
}

// rollbackReservation rollbacks reservation in Redis
func (s *StockService) rollbackReservation(
	ctx context.Context,
	productID reservation.ProductID,
	quantity int,
	reservationID reservation.ReservationID,
) error {
	// Return stock
	if _, err := s.stockRepo.Release(ctx, stock.ProductID(productID), quantity); err != nil {
		return err
	}

	// Delete reservation
	if err := s.cacheReservationRepo.Delete(ctx, reservationID); err != nil {
		logger.WarnContext(ctx, "failed to delete reservation during rollback",
			zap.String("reservation_id", reservationID.String()),
			zap.Error(err),
		)
	}

	return nil
}

// publishReservedEvent publishes stock.reserved event to outbox
func (s *StockService) publishReservedEvent(ctx context.Context, res *reservation.Reservation) error {
	events := res.DomainEvents()
	for _, event := range events {
		outboxEvent := postgres.NewOutboxEvent(
			"reservation",
			res.ID().String(),
			event.EventType(),
			s.reservationEventToPayload(event),
		)

		if err := s.outboxRepo.Insert(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to insert outbox event: %w", err)
		}
	}

	res.ClearEvents()
	return nil
}

// publishDepletedEvent publishes stock.depleted event
func (s *StockService) publishDepletedEvent(ctx context.Context, productID stock.ProductID) error {
	event := stock.NewStockDepletedEvent(productID)

	outboxEvent := postgres.NewOutboxEvent(
		"stock",
		productID.String(),
		event.EventType(),
		map[string]interface{}{
			"product_id":  productID.String(),
			"occurred_at": event.OccurredAt().Format(time.RFC3339),
		},
	)

	if err := s.outboxRepo.Insert(ctx, outboxEvent); err != nil {
		return fmt.Errorf("failed to insert depleted event: %w", err)
	}

	return nil
}

// publishReleasedEvent publishes stock.released event
func (s *StockService) publishReleasedEvent(ctx context.Context, res *reservation.Reservation) error {
	events := res.DomainEvents()
	for _, event := range events {
		outboxEvent := postgres.NewOutboxEvent(
			"reservation",
			res.ID().String(),
			event.EventType(),
			s.reservationEventToPayload(event),
		)

		if err := s.outboxRepo.Insert(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to insert outbox event: %w", err)
		}
	}

	res.ClearEvents()
	return nil
}

// checkAndPublishLowStock checks if stock is low and publishes event
func (s *StockService) checkAndPublishLowStock(ctx context.Context, productID stock.ProductID, currentQty int) {
	// Get full stock info to check threshold
	stk, err := s.stockRepo.FindByProductID(ctx, productID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to check low stock",
			zap.String("product_id", productID.String()),
			zap.Error(err),
		)
		return
	}

	if stk.IsLowStock() {
		logger.InfoContext(ctx, "stock is low",
			zap.String("product_id", productID.String()),
			zap.Int("quantity", currentQty),
			zap.Int("threshold", stk.GetLowStockThreshold()),
		)

		event := stock.NewStockLowEvent(productID, currentQty, stk.GetLowStockThreshold())

		outboxEvent := postgres.NewOutboxEvent(
			"stock",
			productID.String(),
			event.EventType(),
			map[string]interface{}{
				"product_id":  productID.String(),
				"quantity":    currentQty,
				"threshold":   stk.GetLowStockThreshold(),
				"occurred_at": event.OccurredAt().Format(time.RFC3339),
			},
		)

		if err := s.outboxRepo.Insert(ctx, outboxEvent); err != nil {
			logger.ErrorContext(ctx, "failed to publish low stock event",
				zap.String("product_id", productID.String()),
				zap.Error(err),
			)
		}
	}
}

// updateReservationStatus updates reservation status in persistent Repo
func (s *StockService) updateReservationStatus(
	ctx context.Context,
	id reservation.ReservationID,
	status reservation.ReservationStatus,
) error {
	return s.persistentReservationRepo.UpdateStatus(ctx, id, status)
}

// reservationEventToPayload converts reservation event to payload
func (s *StockService) reservationEventToPayload(event reservation.DomainEvent) map[string]interface{} {
	payload := map[string]interface{}{
		"occurred_at": event.OccurredAt().Format(time.RFC3339),
	}

	switch e := event.(type) {
	case reservation.ReservationCreatedEvent:
		payload["reservation_id"] = e.ReservationID.String()
		payload["product_id"] = e.ProductID.String()
		payload["user_id"] = e.UserID.String()
		payload["quantity"] = e.Quantity

	case reservation.ReservationReleasedEvent:
		payload["reservation_id"] = e.ReservationID.String()
		payload["product_id"] = e.ProductID.String()
		payload["quantity"] = e.Quantity

	case reservation.ReservationConsumedEvent:
		payload["reservation_id"] = e.ReservationID.String()
		payload["product_id"] = e.ProductID.String()
		payload["order_id"] = e.OrderID
	}

	return payload
}
