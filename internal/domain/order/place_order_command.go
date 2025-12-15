package order

type PlaceOrderCommand struct {
	StockID  string `json:"stock_id" binding:"required"`
	BuyerID  string `json:"buyer_id"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}
