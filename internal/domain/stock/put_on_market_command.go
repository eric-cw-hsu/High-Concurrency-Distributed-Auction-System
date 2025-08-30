package stock

type PutOnMarketCommand struct {
	ProductId string  `json:"product_id"`
	Quantity  int     `json:"quantity" binding:"required,min=1"` // Quantity must be greater than 0
	Price     float64 `json:"price" binding:"required,min=0.01"` // Price must be greater than 0
	SellerId  string  `json:"seller_id"`                         // Seller ID must be provided
}
