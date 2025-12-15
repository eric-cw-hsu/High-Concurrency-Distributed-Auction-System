package stock

import "time"

// StockUpdatedPayload is the event payload for stock.updated
type StockUpdatedPayload struct {
	StockID   string
	Quantity  int
	UpdatedAt time.Time
}
