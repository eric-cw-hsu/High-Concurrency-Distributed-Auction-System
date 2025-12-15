package order

import "time"

type Order struct {
	OrderID    string // UUID
	StockID    string // UUID
	BuyerID    string // UUID
	Quantity   int
	TotalPrice float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
