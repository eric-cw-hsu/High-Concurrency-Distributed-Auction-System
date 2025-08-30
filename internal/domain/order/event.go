package order

import "time"

type OrderPlacedEvent struct {
	OrderId    string    `json:"order_id"`
	StockId    string    `json:"stock_id"`
	BuyerId    string    `json:"buyer_id"`
	Quantity   int       `json:"quantity"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"` // Unix timestamp in seconds
	UpdatedAt  time.Time `json:"updated_at"` // Unix timestamp in seconds

	Timestamp time.Time `json:"timestamp"` // Unix timestamp in seconds
}

func (e *OrderPlacedEvent) OccurredOn() time.Time {
	return e.Timestamp
}

func (e *OrderPlacedEvent) EventType() string {
	return "order.placed"
}

func (e *OrderPlacedEvent) AggregateId() string {
	return e.OrderId
}

type OrderCancelledEvent struct {
	OrderId   string    `json:"order_id"`
	StockId   string    `json:"stock_id"`
	BuyerId   string    `json:"buyer_id"`
	Quantity  int       `json:"quantity"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *OrderCancelledEvent) OccurredOn() time.Time {
	return e.Timestamp
}

func (e *OrderCancelledEvent) EventType() string {
	return "order.cancelled"
}

func (e *OrderCancelledEvent) AggregateId() string {
	return e.OrderId
}
