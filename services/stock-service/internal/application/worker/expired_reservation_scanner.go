package worker

import (
	"context"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"go.uber.org/zap"
)

// ExpiredReservationScanner scans for expired reservations and releases them
// This is a safety net mechanism - primary expiry is handled by Order Service
type ExpiredReservationScanner struct {
	stockService   *service.StockService
	persistentRepo reservation.PersistentRepository
	scanInterval   time.Duration
	timeWindow     time.Duration
	batchSize      int
}

// NewExpiredReservationScanner creates a new reservation scanner
func NewExpiredReservationScanner(
	stockService *service.StockService,
	persistentRepo reservation.PersistentRepository,
	cfg *config.ExpiredReservationScannerConfig,
) *ExpiredReservationScanner {
	return &ExpiredReservationScanner{
		stockService:   stockService,
		persistentRepo: persistentRepo,
		scanInterval:   cfg.ScanInterval,
		timeWindow:     cfg.TimeWindow,
		batchSize:      cfg.BatchSize,
	}
}

// Start starts the scanner worker
func (s *ExpiredReservationScanner) Start(ctx context.Context) error {
	zap.L().Info("starting reservation scanner",
		zap.Duration("scan_interval", s.scanInterval),
		zap.Duration("time_window", s.timeWindow),
		zap.Int("batch_size", s.batchSize),
	)

	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	// Run once immediately on startup
	if err := s.scanAndReleaseExpired(ctx); err != nil {
		zap.L().Error("initial scan failed", zap.Error(err))
	}

	for {
		select {
		case <-ticker.C:
			if err := s.scanAndReleaseExpired(ctx); err != nil {
				zap.L().Error("scan failed", zap.Error(err))
			}

		case <-ctx.Done():
			zap.L().Info("reservation scanner stopping")
			return nil
		}
	}
}

// scanAndReleaseExpired scans for expired reservations and releases them
func (s *ExpiredReservationScanner) scanAndReleaseExpired(ctx context.Context) error {
	zap.L().Debug("scanning for expired reservations")

	// Calculate time window (only scan recent expirations)
	now := time.Now()
	windowStart := now.Add(-s.timeWindow)

	// Find expired reservations within the time window
	expired, err := s.persistentRepo.FindExpiredWithinWindow(ctx, windowStart, now, s.batchSize)
	if err != nil {
		zap.L().Error("failed to find expired reservations", zap.Error(err))
		return err
	}

	if len(expired) == 0 {
		zap.L().Debug("no expired reservations found")
		return nil
	}

	zap.L().Info("found expired reservations",
		zap.Int("count", len(expired)),
	)

	successCount := 0
	failCount := 0

	for _, res := range expired {
		// Call Release (idempotent - safe even if already released)
		_, err := s.stockService.Release(ctx, res.ID().String())
		if err != nil {
			// Log but continue with other reservations
			zap.L().Error("failed to release expired reservation",
				zap.String("reservation_id", res.ID().String()),
				zap.String("product_id", res.ProductID().String()),
				zap.Time("expired_at", res.ExpiredAt()),
				zap.Error(err),
			)
			failCount++
			continue
		}

		successCount++
	}

	zap.L().Info("expired reservations processed",
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
		zap.Int("total", len(expired)),
	)

	return nil
}
