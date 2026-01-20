package service

import (
	"context"
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/order"
	productprice "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/product_price"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/infrastructure/persistence/postgres"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/infrastructure/persistence/redis"
)

type OrderAppService struct {
	txManager    *postgres.TxManager
	timeoutQueue *redis.TimeoutQueue
	// Use a read-only repository for queries outside transactions
	orderRepo          order.Repository
	productPriceRepo   productprice.Repository
	productPriceClient productprice.ProductClient
}

func NewOrderAppService(
	tm *postgres.TxManager,
	tq *redis.TimeoutQueue,
	orderRepo order.Repository,
	productPriceRepo productprice.Repository,
	productPriceClient productprice.ProductClient,
) *OrderAppService {
	return &OrderAppService{
		txManager:          tm,
		timeoutQueue:       tq,
		orderRepo:          orderRepo,
		productPriceRepo:   productPriceRepo,
		productPriceClient: productPriceClient,
	}
}

var _ order.Service = (*OrderAppService)(nil)
var _ order.Creator = (*OrderAppService)(nil)

// CreateOrder coordinates the atomic creation of an order and its outbox events
func (s *OrderAppService) CreateOrder(
	ctx context.Context,
	reservationIDStr string,
	userIDStr string,
	productIDStr string,
	quantity int,
	unitPriceAmount int64, // Input: unit price in cents
	currency string, // Input: e.g., "USD"
) (*order.Order, error) {
	// 1. Parse string inputs into Domain Value Objects
	resID, err := order.ParseReservationID(reservationIDStr)
	if err != nil {
		return nil, err
	}
	uID, err := order.ParseUserID(userIDStr)
	if err != nil {
		return nil, err
	}
	pID, err := order.ParseProductID(productIDStr)
	if err != nil {
		return nil, err
	}
	// Create Money for unit price
	unitPrice, err := order.NewMoney(unitPriceAmount, currency)
	if err != nil {
		return nil, err
	}

	// 2. Create Aggregate - Pricing & TotalPrice are calculated INSIDE NewOrder
	o, err := order.NewOrder(resID, uID, pID, quantity, unitPrice)
	if err != nil {
		return nil, err
	}

	// 3. Persist via TxManager
	// Repo handles converting o.Pricing() into OrderModel.UnitPrice and OrderModel.TotalPrice
	err = s.txManager.Execute(ctx, func(p postgres.RepositoryProvider) error {
		if err := p.Orders().Save(ctx, o); err != nil {
			return err
		}

		for _, event := range o.DomainEvents() {
			if err := p.Outbox().SaveEvent(ctx, o.ID().String(), event); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	_ = s.timeoutQueue.Add(ctx, o.ID(), o.ExpiresAt())
	return o, nil
}

// CancelExpiredOrder marks an order as expired and stages a cancellation event atomically
func (s *OrderAppService) CancelExpiredOrder(ctx context.Context, orderIDStr string) error {
	orderID, err := order.ParseOrderID(orderIDStr)
	if err != nil {
		return err
	}

	// Execute within a transaction to ensure Outbox consistency
	return s.txManager.Execute(ctx, func(p postgres.RepositoryProvider) error {
		// 1. Fetch aggregate using the transactional repository
		o, err := p.Orders().FindByID(ctx, orderID)
		if err != nil {
			return err
		}

		// 2. Domain logic: Perform state transition
		if err := o.MarkAsExpired(); err != nil {
			// If state is already terminal (Paid/Cancelled), we skip
			return nil
		}

		// 3. Save updated state and stage domain events
		if err := p.Orders().Save(ctx, o); err != nil {
			return err
		}

		for _, event := range o.DomainEvents() {
			if err := p.Outbox().SaveEvent(ctx, o.ID().String(), event); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetOrder performs a simple read operation
func (s *OrderAppService) GetOrder(ctx context.Context, orderIDStr string) (*order.Order, error) {
	orderID, err := order.ParseOrderID(orderIDStr)
	if err != nil {
		return nil, err
	}

	return s.orderRepo.FindByID(ctx, orderID)
}

// CreateOrderFromReservation implements kafka.OrderCreator interface
func (s *OrderAppService) CreateOrderFromReservation(
	ctx context.Context,
	reservationID, userID, productID string,
	quantity int,
) error {
	priceInfo, err := s.productPriceRepo.GetByID(ctx, productID)
	// 2. Fallback: If not found locally, fetch via gRPC from Product Service
	if err != nil {
		// Use a gRPC client (which you'll need to inject into this service)
		priceInfo, err = s.productPriceClient.FetchProductDetail(ctx, productID)
		if err != nil {
			return fmt.Errorf("failed to fetch price from product service after local miss: %w", err)
		}

		// 3. (Optional but recommended) Update local cache for next time
		_ = s.productPriceRepo.Upsert(ctx, priceInfo)
	}

	_, err = s.CreateOrder(ctx, reservationID, userID, productID, quantity, priceInfo.UnitPrice, priceInfo.Currency)
	if err != nil {
		return fmt.Errorf("failed to create order for reservation %s: %w", reservationID, err)
	}

	return nil
}

// ListUserOrders retrieves a paginated list of orders for a specific user.
// This is a read-only operation and doesn't require a transaction.
func (s *OrderAppService) ListUserOrders(
	ctx context.Context,
	userIDStr string,
	limit, offset int,
) ([]*order.Order, error) {
	// 1. Parse and validate the user ID
	uID, err := order.ParseUserID(userIDStr)
	if err != nil {
		return nil, err
	}

	// 2. Fetch from repository (Read-only path)
	orders, err := s.orderRepo.FindByUserID(ctx, uID, limit, offset)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
