package stock

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/stock"
	"github.com/samborkent/uuidv7"
)

type PutOnMarketUsecase struct {
	stockRepository stock.StockRepository
	stockCache      stock.StockCache
}

func NewPutOnMarketUsecase(stockRepository stock.StockRepository, stockCache stock.StockCache) *PutOnMarketUsecase {
	return &PutOnMarketUsecase{
		stockRepository: stockRepository,
		stockCache:      stockCache,
	}
}

// Execute puts a stock item on the market with the provided command.
// It creates a new stock item, saves it to the repository, and updates the cache.
func (uc *PutOnMarketUsecase) Execute(ctx context.Context, command stock.PutOnMarketCommand) error {
	stockItem := stock.Stock{
		ID:        uuidv7.New().String(),
		ProductID: command.ProductID,
		Quantity:  command.Quantity,
		Price:     command.Price,
		SellerID:  command.SellerID,
	}

	_, err := uc.stockRepository.SaveStock(ctx, &stockItem)
	if err != nil {
		return err
	}

	// Update the stock cache
	if err := uc.stockCache.SetStock(ctx, stockItem.ID, stockItem.Quantity); err != nil {
		return err
	}

	if err := uc.stockCache.SetPrice(ctx, stockItem.ID, stockItem.Price); err != nil {
		return err
	}

	return nil
}
