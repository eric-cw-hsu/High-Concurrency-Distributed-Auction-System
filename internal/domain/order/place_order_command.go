package order

type PlaceOrderCommand struct {
	StockId  string `json:"stock_id" binding:"required"`
	BuyerId  string `json:"buyer_id"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}
