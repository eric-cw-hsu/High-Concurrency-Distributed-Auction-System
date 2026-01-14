package service

import (
	"context"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"go.uber.org/zap"
)

type ReservationPersistWorker struct {
	cfg                       *config.ServiceConfig
	persistentReservationRepo reservation.PersistentRepository
	persistQueue              chan *reservation.Reservation
}

func NewReservationPersistQueue(cfg *config.ServiceConfig) chan *reservation.Reservation {
	return make(chan *reservation.Reservation, cfg.PersistBatchSize*10)
}

func NewReservationPersistWorker(
	cfg *config.ServiceConfig,
	persistentReservationRepo reservation.PersistentRepository,
	persistQueue chan *reservation.Reservation,
) *ReservationPersistWorker {
	return &ReservationPersistWorker{
		cfg:                       cfg,
		persistentReservationRepo: persistentReservationRepo,
		persistQueue:              persistQueue,
	}
}

// startPersistWorker starts background worker for async PostgreSQL writes
func (s *ReservationPersistWorker) Start(ctx context.Context) {
	batch := make([]*reservation.Reservation, 0, s.cfg.PersistBatchSize)
	ticker := time.NewTicker(s.cfg.PersistFlushWindow)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.flushBatch(ctx, batch)
			return
		case res := <-s.persistQueue:
			batch = append(batch, res)

			if len(batch) >= s.cfg.PersistBatchSize {
				s.flushBatch(ctx, batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.flushBatch(ctx, batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch flushes a batch of reservations to PostgreSQL
func (s *ReservationPersistWorker) flushBatch(ctx context.Context, batch []*reservation.Reservation) {
	logger.DebugContext(ctx, "flushing reservation batch to postgresql",
		zap.Int("count", len(batch)),
	)

	for _, res := range batch {
		if err := s.persistentReservationRepo.Save(ctx, res); err != nil {
			logger.ErrorContext(ctx, "failed to persist reservation",
				zap.String("reservation_id", res.ID().String()),
				zap.String("product_id", res.ProductID().String()),
				zap.Error(err),
			)
			// Continue with other reservations
		}
	}

	logger.InfoContext(ctx, "reservation batch persisted",
		zap.Int("count", len(batch)),
	)
}
