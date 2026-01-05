package dto

// CreateProductRequest represents HTTP request to create product
type CreateProductRequest struct {
	Name           string `json:"name" binding:"required,max=200"`
	Description    string `json:"description" binding:"max=5000"`
	RegularPrice   int64  `json:"regular_price" binding:"required,min=1"`
	FlashSalePrice *int64 `json:"flash_sale_price,omitempty" binding:"omitempty,min=1"`
	Currency       string `json:"currency" binding:"required,len=3"`
}

// UpdateProductInfoRequest represents HTTP request to update product info
type UpdateProductInfoRequest struct {
	Name        string `json:"name" binding:"required,max=200"`
	Description string `json:"description" binding:"max=5000"`
}

// UpdateProductPricingRequest represents HTTP request to update pricing
type UpdateProductPricingRequest struct {
	RegularPrice   int64  `json:"regular_price" binding:"required,min=1"`
	FlashSalePrice *int64 `json:"flash_sale_price,omitempty" binding:"omitempty,min=1"`
	Currency       string `json:"currency" binding:"required,len=3"`
}

// ProductResponse represents product data in HTTP response
type ProductResponse struct {
	ID          string     `json:"id"`
	SellerID    string     `json:"seller_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Pricing     PricingDTO `json:"pricing"`
	Status      string     `json:"status"`
	StockStatus string     `json:"stock_status"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

// PricingDTO represents pricing information
type PricingDTO struct {
	RegularPrice   MoneyDTO  `json:"regular_price"`
	FlashSalePrice *MoneyDTO `json:"flash_sale_price,omitempty"`
}

// MoneyDTO represents monetary amount
type MoneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// ProductListResponse represents paginated product list
type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int32             `json:"total"`
	Page     int32             `json:"page"`
	PageSize int32             `json:"page_size"`
}
