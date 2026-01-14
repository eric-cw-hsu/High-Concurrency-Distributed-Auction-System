package stock

import "time"

// DomainEvent interface
type DomainEvent interface {
	OccurredAt() time.Time
	EventType() string
}

// StockSetEvent is emitted when stock is initially set
type StockSetEvent struct {
	ProductID  ProductID
	Quantity   int
	occurredAt time.Time
}

func NewStockSetEvent(productID ProductID, quantity int) StockSetEvent {
	return StockSetEvent{
		ProductID:  productID,
		Quantity:   quantity,
		occurredAt: time.Now(),
	}
}

func (e StockSetEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e StockSetEvent) EventType() string {
	return "stock.set"
}

// StockDepletedEvent is emitted when stock reaches zero
type StockDepletedEvent struct {
	ProductID  ProductID
	occurredAt time.Time
}

func NewStockDepletedEvent(productID ProductID) StockDepletedEvent {
	return StockDepletedEvent{
		ProductID:  productID,
		occurredAt: time.Now(),
	}
}

func (e StockDepletedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e StockDepletedEvent) EventType() string {
	return "stock.depleted"
}

// StockLowEvent is emitted when stock falls below threshold
type StockLowEvent struct {
	ProductID  ProductID
	Quantity   int
	Threshold  int
	occurredAt time.Time
}

func NewStockLowEvent(productID ProductID, quantity int, threshold int) StockLowEvent {
	return StockLowEvent{
		ProductID:  productID,
		Quantity:   quantity,
		Threshold:  threshold,
		occurredAt: time.Now(),
	}
}

func (e StockLowEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e StockLowEvent) EventType() string {
	return "stock.low"
}
