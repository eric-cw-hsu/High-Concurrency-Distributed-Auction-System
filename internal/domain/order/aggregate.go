package order

import (
	"time"

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
	ID        string
	BuyerID   string
	StockID   string
	Quantity  int
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time

	status        OrderStatus
	eventPayloads []interface{}
}

func (o *OrderAggregate) Status() OrderStatus {
	return o.status
}

func (o *OrderAggregate) PopEventPayloads() []interface{} {
	payloads := o.eventPayloads
	o.eventPayloads = make([]interface{}, 0)
	return payloads
}

func NewOrderAggregate(buyerID, stockID string, price float64, requestQuantity int) (*OrderAggregate, error) {
	if buyerID == "" {
		return nil, ErrEmptyBuyerID
	}
	if stockID == "" {
		return nil, ErrEmptyStockID
	}
	if requestQuantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	totalPrice := price * float64(requestQuantity)

	return &OrderAggregate{
		ID:        uuidv7.New().String(),
		BuyerID:   buyerID,
		StockID:   stockID,
		Quantity:  requestQuantity,
		Price:     totalPrice,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),

		status:        OrderStatusPending,
		eventPayloads: make([]interface{}, 0),
	}, nil
}

func (o *OrderAggregate) ConfirmAfterStockDeduction(timestamp time.Time) error {
	if o.status != OrderStatusPending {
		return ErrOrderNotConfirmed
	}
	o.status = OrderStatusConfirmed
	o.UpdatedAt = time.Now()

	event := OrderReservedPayload{
		OrderID:    o.ID,
		BuyerID:    o.BuyerID,
		StockID:    o.StockID,
		Quantity:   o.Quantity,
		TotalPrice: o.Price,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,

		Timestamp: timestamp,
	}

	o.eventPayloads = append(o.eventPayloads, &event)

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
