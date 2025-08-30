package stock

import "context"

type StockCache interface {
	// DecreaseStock decreases the stock of an item by the specified quantity.
	// It returns the timestamp of the operation if successful, or an error if the item is not found or out of stock.
	DecreaseStock(ctx context.Context, stockId string, quantity int) (int64, error)
	RestoreStock(ctx context.Context, stockId string, quantity int) error
	GetStock(ctx context.Context, stockId string) (int, error)
	SetStock(ctx context.Context, stockId string, quantity int) error
	SetPrice(ctx context.Context, stockId string, price float64) error
	GetPrice(ctx context.Context, stockId string) (float64, error)
	RemoveAll(ctx context.Context) error
}
