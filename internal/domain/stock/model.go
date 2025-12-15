package stock

import "time"

type Stock struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"` // ID of the product this stock belongs to
	Price     float64   `json:"price"`      // Current price of the stock item
	Quantity  int       `json:"quantity"`   // Current stock quantity
	SellerID  string    `json:"seller_id"`  // ID of the seller who owns the stock
	CreatedAt time.Time `json:"created_at"` // Unix timestamp in seconds
	UpdatedAt time.Time `json:"updated_at"` // Unix timestamp in seconds
}
