package stock

import "context"

type StockRepository interface {
	GetStocksByProductID(ctx context.Context, productID string) ([]*Stock, error)
	DecreaseStock(ctx context.Context, stockID string, quantity int) (int, error)
	GetStockByID(ctx context.Context, stockID string) (*Stock, error)
	SaveStock(ctx context.Context, stock *Stock) (*Stock, error)
	GetAllStocks(ctx context.Context) ([]*Stock, error)
	UpdateStockQuantity(ctx context.Context, stock *Stock) (*Stock, error)
}
