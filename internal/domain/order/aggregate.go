package order

import (
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain"
	"github.com/samborkent/uuidv7"
)

type OrderStatus int

const (
	OrderStatusPending OrderStatus = iota
	OrderStatusConfirmed
	OrderStatusCancelled
)

// Order Aggregate Root
type OrderAggregate struct {
	Id        string
	BuyerId   string
	StockId   string
	Quantity  int
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time

	status OrderStatus
	events []domain.DomainEvent
}

func (o *OrderAggregate) Status() OrderStatus {
	return o.status
}

func (o *OrderAggregate) DomainEvents() []domain.DomainEvent {
	return o.events
}

func NewOrderAggregate(buyerId, stockId string, price float64, requestQuantity int) (*OrderAggregate, error) {
	if buyerId == "" {
		return nil, ErrEmptyBuyerId
	}
	if stockId == "" {
		return nil, ErrEmptyStockId
	}
	if requestQuantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	totalPrice := price * float64(requestQuantity)

	return &OrderAggregate{
		Id:        uuidv7.New().String(),
		BuyerId:   buyerId,
		StockId:   stockId,
		Quantity:  requestQuantity,
		Price:     totalPrice,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),

		status: OrderStatusPending,
		events: make([]domain.DomainEvent, 0),
	}, nil
}

func (o *OrderAggregate) ConfirmAfterStockDeduction(timestamp time.Time) error {
	if o.status != OrderStatusPending {
		return ErrOrderNotConfirmed
	}
	o.status = OrderStatusConfirmed
	o.UpdatedAt = time.Now()

	event := OrderPlacedEvent{
		OrderId:    o.Id,
		BuyerId:    o.BuyerId,
		StockId:    o.StockId,
		Quantity:   o.Quantity,
		TotalPrice: o.Price,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,

		Timestamp: timestamp,
	}

	o.events = append(o.events, &event)

	return nil
}

func (o *OrderAggregate) Cancel() error {
	if o.status != OrderStatusPending {
		return ErrOrderNotCancelled
	}
	o.status = OrderStatusCancelled
	o.UpdatedAt = time.Now()

	return nil
}
