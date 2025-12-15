package order

import (
	"context"
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/stock"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/producer"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	walletUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/wallet"
)

type PlaceOrderUsecase struct {
	producer      producer.EventProducer
	stockCache    stock.StockCache
	walletService walletUsecase.WalletService
}

func NewPlaceOrderUsecase(
	producer producer.EventProducer,
	stockCache stock.StockCache,
	walletService walletUsecase.WalletService,
) *PlaceOrderUsecase {
	return &PlaceOrderUsecase{
		producer:      producer,
		stockCache:    stockCache,
		walletService: walletService,
	}
}

// Execute handles placing an order with immediate payment processing and stock reservation.
// This ensures transaction consistency and real-time balance deduction.
func (uc *PlaceOrderUsecase) Execute(ctx context.Context, command order.PlaceOrderCommand) error {
	// 1. Get stock price
	price, err := uc.stockCache.GetPrice(ctx, command.StockID)
	if err != nil {
		return &order.StockError{
			StockID:   command.StockID,
			Operation: "get_price",
			Wrapped:   err,
		}
	}

	// 2. Calculate total amount
	totalAmount := price * float64(command.Quantity)

	// 3. Check stock availability
	availableQty, err := uc.stockCache.GetStock(ctx, command.StockID)
	if err != nil {
		return &order.StockError{
			StockID:   command.StockID,
			Operation: "get_quantity",
			Wrapped:   err,
		}
	}

	if availableQty < command.Quantity {
		return &order.InsufficientStockError{
			StockID:   command.StockID,
			Available: availableQty,
			Requested: command.Quantity,
		}
	}

	// 4. CRITICAL: Process payment IMMEDIATELY (real-time balance deduction)
	// This ensures the user's balance is updated before the API call returns
	if _, err := uc.walletService.ProcessPaymentWithSufficientFunds(ctx, command.BuyerID, "", totalAmount); err != nil {
		return &order.PaymentError{
			UserId:  command.BuyerID,
			Amount:  totalAmount,
			Reason:  "insufficient funds or payment processing failed",
			Wrapped: err,
		}
	}

	// 5. Reduce stock in cache (also immediate)
	occurredOn, err := uc.stockCache.DecreaseStock(ctx, command.StockID, command.Quantity)
	if err != nil {
		// CRITICAL: Rollback payment immediately if stock update fails
		uc.walletService.ProcessRefundSafely(ctx, command.BuyerID, "", totalAmount)
		return &order.StockError{
			StockID:   command.StockID,
			Operation: "update",
			Wrapped:   err,
		}
	}

	// 6. Create and publish order event (asynchronous for performance)
	// This is for audit trail, notifications, etc. - NOT for payment processing
	go uc.publishOrderEventAsync(ctx, command, totalAmount, time.Unix(occurredOn, 0))

	return nil
}

// publishOrderEventAsync publishes order events asynchronously for audit and notifications
func (uc *PlaceOrderUsecase) publishOrderEventAsync(ctx context.Context, command order.PlaceOrderCommand, totalAmount float64, occurredOn time.Time) {
	// Create order aggregate for event
	orderAggregate, err := order.NewOrderAggregate(command.BuyerID, command.StockID, totalAmount, command.Quantity)
	if err != nil {
		// Log error but don't fail the main transaction
		logger.Errorf("Failed to create order aggregate for event: %v\n", err)
		return
	}

	if err := orderAggregate.ConfirmAfterStockDeduction(occurredOn); err != nil {
		// Log error but don't fail the main transaction
		logger.Errorf("Failed to confirm order aggregate for event: %v\n", err)
		return
	}

	// Prepare event payload
	for _, payload := range orderAggregate.PopEventPayloads() {
		domainEvent := message.DomainEvent{
			Name:        "order.reserved",
			AggregateID: orderAggregate.ID,
			OccurredAt:  occurredOn,
			Payload:     payload,
			Version:     1,
		}

		// Publish event for audit trail, notifications, etc.
		if err := uc.producer.PublishEvent(ctx, &domainEvent); err != nil {
			// Log error but don't fail - this is for audit/notification only
			logger.Errorf("Failed to publish order event: %v\n", err)
		}
	}
}
