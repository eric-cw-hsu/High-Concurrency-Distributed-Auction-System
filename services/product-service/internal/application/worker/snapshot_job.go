package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/domain/product"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/infrastructure/persistence/postgres"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// SnapshotJob generates periodic snapshots of active products
type SnapshotJob struct {
	productRepo  product.Repository
	outboxRepo   *postgres.OutboxRepository
	kafkaBrokers []string
	kafkaTopic   string
	interval     time.Duration
}

// NewSnapshotJob creates a new snapshot job
func NewSnapshotJob(
	productRepo product.Repository,
	outboxRepo *postgres.OutboxRepository,
	kafkaBrokers []string,
	kafkaTopic string,
	cfg *config.SnapshotConfig,
) *SnapshotJob {
	return &SnapshotJob{
		productRepo:  productRepo,
		outboxRepo:   outboxRepo,
		kafkaBrokers: kafkaBrokers,
		kafkaTopic:   kafkaTopic,
		interval:     cfg.Interval,
	}
}

// Start starts the snapshot job
func (j *SnapshotJob) Start(ctx context.Context) error {
	zap.L().Info("starting product snapshot job",
		zap.Duration("interval", j.interval),
	)

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	// Generate initial snapshot on startup
	if err := j.generateSnapshot(ctx); err != nil {
		zap.L().Error("failed to generate initial snapshot", zap.Error(err))
	}

	for {
		select {
		case <-ticker.C:
			if err := j.generateSnapshot(ctx); err != nil {
				zap.L().Error("failed to generate snapshot", zap.Error(err))
			}

		case <-ctx.Done():
			zap.L().Info("snapshot job stopping")
			return nil
		}
	}
}

// generateSnapshot generates and publishes a snapshot
func (j *SnapshotJob) generateSnapshot(ctx context.Context) error {
	zap.L().Info("generating product snapshot")

	// Step 1: Get all active products
	products, err := j.productRepo.FindAllActiveProducts(ctx)
	if err != nil {
		return fmt.Errorf("failed to find active products: %w", err)
	}

	// Extract product IDs
	activeProductIDs := make([]string, 0, len(products))
	for _, p := range products {
		activeProductIDs = append(activeProductIDs, p.ID().String())
	}

	zap.L().Info("active products collected",
		zap.Int("count", len(activeProductIDs)),
	)

	// Step 2: Get current offsets of all partitions
	partitionOffsets, err := j.getCurrentPartitionOffsets(ctx)
	if err != nil {
		zap.L().Error("failed to get partition offsets", zap.Error(err))
		// Continue without offsets (fallback to timestamp-based)
		partitionOffsets = make(map[int]int64)
	}

	zap.L().Info("partition offsets captured",
		zap.Int("partitions", len(partitionOffsets)),
	)

	// Step 3: Create snapshot domain event
	now := time.Now()
	snapshotEvent := product.NewProductSnapshotEvent(
		activeProductIDs,
		partitionOffsets,
		now,
	)

	// Step 4: Convert partition offsets to interface map for JSON
	offsetsMap := make(map[string]interface{})
	for partID, offset := range partitionOffsets {
		offsetsMap[fmt.Sprintf("%d", partID)] = offset
	}

	// Step 5: Create outbox event
	outboxEvent := postgres.NewOutboxEvent(
		"product",
		"__snapshot__",
		snapshotEvent.EventType(),
		map[string]interface{}{
			"active_products":   snapshotEvent.ActiveProducts,
			"partition_offsets": offsetsMap,
			"total":             snapshotEvent.Total,
			"occurred_at":       snapshotEvent.OccurredAt().Format(time.RFC3339),
		},
	)

	// Step 6: Insert to outbox
	if err := j.outboxRepo.Insert(ctx, outboxEvent); err != nil {
		zap.L().Error("failed to insert snapshot to outbox",
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("product snapshot generated and queued",
		zap.Int("active_products", len(activeProductIDs)),
		zap.Int("partitions", len(partitionOffsets)),
	)

	return nil
}

// getCurrentPartitionOffsets gets current end offsets of all partitions
func (j *SnapshotJob) getCurrentPartitionOffsets(ctx context.Context) (map[int]int64, error) {
	zap.L().Debug("reading kafka partition offsets",
		zap.String("topic", j.kafkaTopic),
	)

	// Connect to Kafka
	conn, err := kafka.Dial("tcp", j.kafkaBrokers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	// Get partitions
	partitions, err := conn.ReadPartitions(j.kafkaTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	offsets := make(map[int]int64)

	// Read last offset of each partition
	for _, partition := range partitions {
		partConn, err := kafka.DialLeader(ctx, "tcp", j.kafkaBrokers[0], j.kafkaTopic, partition.ID)
		if err != nil {
			zap.L().Warn("failed to dial partition leader",
				zap.Int("partition", partition.ID),
				zap.Error(err),
			)
			continue
		}

		lastOffset, err := partConn.ReadLastOffset()
		partConn.Close()

		if err != nil {
			zap.L().Warn("failed to read last offset",
				zap.Int("partition", partition.ID),
				zap.Error(err),
			)
			continue
		}

		offsets[partition.ID] = lastOffset

		zap.L().Debug("partition offset captured",
			zap.Int("partition", partition.ID),
			zap.Int64("offset", lastOffset),
		)
	}

	return offsets, nil
}
