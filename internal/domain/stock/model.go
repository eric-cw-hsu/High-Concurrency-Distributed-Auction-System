package stock

type Stock struct {
	Id        string  `json:"id"`
	ProductId string  `json:"product_id"` // ID of the product this stock belongs to
	Price     float64 `json:"price"`      // Current price of the stock item
	Quantity  int     `json:"quantity"`   // Current stock quantity
	SellerId  string  `json:"seller_id"`  // ID of the seller who owns the stock
	CreatedAt int64   `json:"created_at"` // Unix timestamp in seconds
	UpdatedAt int64   `json:"updated_at"` // Unix timestamp in seconds
}
