package order

import "time"

type OrderReservedPayload struct {
	OrderID    string    `json:"order_id"`
	StockID    string    `json:"stock_id"`
	BuyerID    string    `json:"buyer_id"`
	Quantity   int       `json:"quantity"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"` // Unix timestamp in seconds
	UpdatedAt  time.Time `json:"updated_at"` // Unix timestamp in seconds
	Timestamp  time.Time `json:"timestamp"`  // Unix timestamp in seconds
}
