package postgres

import (
	"database/sql"
	"time"
)

// ProductModel represents the database model for products
type ProductModel struct {
	ID             string        `db:"id"`
	SellerID       string        `db:"seller_id"`
	Name           string        `db:"name"`
	Description    string        `db:"description"`
	RegularPrice   int64         `db:"regular_price"`
	FlashSalePrice sql.NullInt64 `db:"flash_sale_price"`
	Currency       string        `db:"currency"`
	Status         string        `db:"status"`
	StockStatus    string        `db:"stock_status"`
	CreatedAt      time.Time     `db:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at"`
}
