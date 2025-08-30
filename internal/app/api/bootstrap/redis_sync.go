package bootstrap

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/stock"
)

type SyncService struct {
	stockRepo  stock.StockRepository
	stockCache stock.StockCache
}

func NewSyncService(stockRepo stock.StockRepository, stockCache stock.StockCache) *SyncService {
	return &SyncService{
		stockRepo:  stockRepo,
		stockCache: stockCache,
	}
}

func (s *SyncService) SyncStockToRedis(ctx context.Context) error {
	s.stockCache.RemoveAll(ctx)

	stocks, err := s.stockRepo.GetAllStocks(ctx)
	if err != nil {
		return err
	}

	for _, stockItem := range stocks {
		if err := s.stockCache.SetStock(ctx, stockItem.Id, stockItem.Quantity); err != nil {
			return err
		}

		if err := s.stockCache.SetPrice(ctx, stockItem.Id, stockItem.Price); err != nil {
			return err
		}
	}

	return nil
}
