package stock

import "context"

type StockRepository interface {
	GetStocksByProductId(ctx context.Context, productId string) ([]*Stock, error)
	DecreaseStock(ctx context.Context, stockId string, quantity int) (int64, error)
	GetStockById(ctx context.Context, stockId string) (*Stock, error)
	SaveStock(ctx context.Context, stock *Stock) (*Stock, error)
	GetAllStocks(ctx context.Context) ([]*Stock, error)
}
