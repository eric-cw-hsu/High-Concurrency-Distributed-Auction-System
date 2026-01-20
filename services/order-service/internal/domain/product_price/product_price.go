package productprice

import "time"

type ProductPrice struct {
	ProductID string
	UnitPrice int64
	Currency  string
	UpdatedAt time.Time
}

func NewProductPrice(id string, price int64, currency string) *ProductPrice {
	return &ProductPrice{
		ProductID: id,
		UnitPrice: price,
		Currency:  currency,
		UpdatedAt: time.Now(),
	}
}
