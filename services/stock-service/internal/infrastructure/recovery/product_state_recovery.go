package recovery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/messaging/kafka"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/persistence/redis"
	goredis "github.com/redis/go-redis/v9"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type ProductStateRecovery struct {
	redisClient      *goredis.Client
	productStateRepo *redis.ProductStateRepository
	kafkaBrokers     []string
	productTopic     string
}

type SnapshotData struct {
	ActiveProducts   []string         `json:"active_products"`
	PartitionOffsets map[string]int64 `json:"partition_offsets"` // "0" -> offset
	Total            int              `json:"total"`
	OccurredAt       string           `json:"occurred_at"`
}

type SnapshotInfo struct {
	Data      *SnapshotData
	Offset    int64
	Partition int
}

type PartitionInfo struct {
	ID          int
	LastOffset  int64
	FirstOffset int64
}

func NewProductStateRecovery(
	redisClient *goredis.Client,
	productStateRepo *redis.ProductStateRepository,
	kafkaBrokers []string,
	productTopic string,
) *ProductStateRecovery {
	return &ProductStateRecovery{
		redisClient:      redisClient,
		productStateRepo: productStateRepo,
		kafkaBrokers:     kafkaBrokers,
		productTopic:     productTopic,
	}
}

func (r *ProductStateRecovery) CheckAndRecover(ctx context.Context) error {
	zap.L().Info("checking product state cache health")

	count, err := r.productStateRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to check cache: %w", err)
	}

	zap.L().Info("product state cache status", zap.Int64("count", count))

	if count == 0 {
		zap.L().Warn("product state cache is empty, triggering recovery")
		return r.RecoverWithSnapshotAndReplay(ctx)
	}

	zap.L().Info("product state cache is healthy, skipping recovery")
	return nil
}

func (r *ProductStateRecovery) FullRecovery(ctx context.Context) error {
	zap.L().Info("starting full recovery (clear + rebuild)")

	if err := r.clearCache(ctx); err != nil {
		zap.L().Error("failed to clear cache", zap.Error(err))
	}

	return r.RecoverWithSnapshotAndReplay(ctx)
}

func (r *ProductStateRecovery) RecoverWithSnapshotAndReplay(ctx context.Context) error {
	zap.L().Info("starting recovery with snapshot + replay strategy")

	// Step 1: Get all partitions info
	partitions, err := r.getAllPartitionsInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get partitions info: %w", err)
	}

	zap.L().Info("topic partitions discovered",
		zap.Int("count", len(partitions)),
	)

	// Step 2: Find latest snapshot across all partitions
	snapshot, err := r.findLatestSnapshotAcrossPartitions(ctx, partitions)
	if err != nil {
		return fmt.Errorf("failed to find snapshot: %w", err)
	}

	if snapshot == nil {
		zap.L().Warn("no snapshot found")
		return fmt.Errorf("no snapshot available")
	}

	zap.L().Info("latest snapshot found",
		zap.Int("partition", snapshot.Partition),
		zap.Int64("offset", snapshot.Offset),
		zap.Int("active_products", snapshot.Data.Total),
		zap.Int("partition_offsets_count", len(snapshot.Data.PartitionOffsets)),
	)

	// Step 3: Load snapshot to Redis
	if err := r.loadSnapshot(ctx, snapshot.Data); err != nil {
		return fmt.Errorf("failed to load snapshot: %w", err)
	}

	// Step 4: Replay events after snapshot from ALL partitions using recorded offsets
	if err := r.replayAfterSnapshotWithOffsets(ctx, snapshot, partitions); err != nil {
		zap.L().Warn("replay had issues, but snapshot is loaded", zap.Error(err))
	}

	zap.L().Info("recovery completed successfully")
	return nil
}

func (r *ProductStateRecovery) getAllPartitionsInfo(ctx context.Context) ([]PartitionInfo, error) {
	conn, err := kafkago.Dial("tcp", r.kafkaBrokers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(r.productTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	partitionInfos := make([]PartitionInfo, 0, len(partitions))

	for _, p := range partitions {
		partConn, err := kafkago.DialLeader(ctx, "tcp", r.kafkaBrokers[0], r.productTopic, p.ID)
		if err != nil {
			zap.L().Warn("failed to dial partition leader",
				zap.Int("partition", p.ID),
				zap.Error(err),
			)
			continue
		}

		firstOffset, err := partConn.ReadFirstOffset()
		if err != nil {
			partConn.Close()
			continue
		}

		lastOffset, err := partConn.ReadLastOffset()
		partConn.Close()
		if err != nil {
			continue
		}

		partitionInfos = append(partitionInfos, PartitionInfo{
			ID:          p.ID,
			FirstOffset: firstOffset,
			LastOffset:  lastOffset,
		})

		zap.L().Debug("partition info",
			zap.Int("partition", p.ID),
			zap.Int64("first_offset", firstOffset),
			zap.Int64("last_offset", lastOffset),
		)
	}

	return partitionInfos, nil
}

func (r *ProductStateRecovery) findLatestSnapshotAcrossPartitions(ctx context.Context, partitions []PartitionInfo) (*SnapshotInfo, error) {
	zap.L().Info("scanning all partitions for latest snapshot")

	var latestSnapshot *SnapshotInfo

	for _, partition := range partitions {
		if partition.LastOffset == 0 {
			zap.L().Debug("partition is empty, skipping",
				zap.Int("partition", partition.ID),
			)
			continue
		}

		snapshot, err := r.findLatestSnapshotInPartition(ctx, partition)
		if err != nil {
			zap.L().Warn("failed to scan partition for snapshot",
				zap.Int("partition", partition.ID),
				zap.Error(err),
			)
			continue
		}

		if snapshot != nil {
			if latestSnapshot == nil {
				latestSnapshot = snapshot
			} else {
				currentTime, _ := time.Parse(time.RFC3339, snapshot.Data.OccurredAt)
				latestTime, _ := time.Parse(time.RFC3339, latestSnapshot.Data.OccurredAt)

				if currentTime.After(latestTime) {
					latestSnapshot = snapshot
					zap.L().Info("found newer snapshot",
						zap.Int("partition", snapshot.Partition),
						zap.Int64("offset", snapshot.Offset),
					)
				}
			}
		}
	}

	return latestSnapshot, nil
}

func (r *ProductStateRecovery) findLatestSnapshotInPartition(ctx context.Context, partition PartitionInfo) (*SnapshotInfo, error) {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:   r.kafkaBrokers,
		Topic:     r.productTopic,
		Partition: partition.ID,
	})
	defer reader.Close()

	if err := reader.SetOffset(partition.FirstOffset); err != nil {
		return nil, fmt.Errorf("failed to set offset: %w", err)
	}

	scanCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	var latestSnapshot *SnapshotInfo

	for {
		msg, err := reader.FetchMessage(scanCtx)
		if err != nil {
			if err == context.DeadlineExceeded || err == context.Canceled {
				break
			}
			continue
		}

		var event kafka.EventMessage
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			continue
		}

		if event.EventType == "product.snapshot" {
			var snapshotData SnapshotData
			dataBytes, _ := json.Marshal(event.Data)
			if err := json.Unmarshal(dataBytes, &snapshotData); err != nil {
				zap.L().Warn("failed to parse snapshot data", zap.Error(err))
				continue
			}

			latestSnapshot = &SnapshotInfo{
				Data:      &snapshotData,
				Offset:    msg.Offset,
				Partition: partition.ID,
			}

			zap.L().Debug("snapshot found in partition",
				zap.Int("partition", partition.ID),
				zap.Int64("offset", msg.Offset),
			)
		}

		if msg.Offset >= partition.LastOffset-1 {
			break
		}
	}

	return latestSnapshot, nil
}

func (r *ProductStateRecovery) loadSnapshot(ctx context.Context, snapshot *SnapshotData) error {
	zap.L().Info("loading snapshot to redis",
		zap.Int("products", len(snapshot.ActiveProducts)),
	)

	successCount := 0
	for _, productID := range snapshot.ActiveProducts {
		if err := r.productStateRepo.MarkActive(ctx, productID); err != nil {
			zap.L().Error("failed to mark product active",
				zap.String("product_id", productID),
				zap.Error(err),
			)
			continue
		}
		successCount++
	}

	zap.L().Info("snapshot loaded",
		zap.Int("success", successCount),
		zap.Int("total", len(snapshot.ActiveProducts)),
	)

	return nil
}

func (r *ProductStateRecovery) replayAfterSnapshotWithOffsets(ctx context.Context, snapshot *SnapshotInfo, partitions []PartitionInfo) error {
	zap.L().Info("replaying events after snapshot using partition offsets")
	for _, partition := range partitions {
		zap.L().Info(
			fmt.Sprintf(
				"partition %d: [start offset %d, end offset %d]",
				partition.ID,
				snapshot.Data.PartitionOffsets[fmt.Sprint(partition.ID)],
				partition.LastOffset,
			),
		)
	}

	totalEvents := 0
	startTime := time.Now()

	for _, partition := range partitions {
		if partition.LastOffset == 0 {
			continue
		}

		// Get the snapshot offset for this partition
		startOffset := r.getReplayStartOffset(snapshot, partition.ID)

		if startOffset >= partition.LastOffset {
			zap.L().Debug("no events to replay in partition",
				zap.Int("partition", partition.ID),
				zap.Int64("snapshot_offset", startOffset),
				zap.Int64("last_offset", partition.LastOffset),
			)
			continue
		}

		zap.L().Info("replaying partition",
			zap.Int("partition", partition.ID),
			zap.Int64("start_offset", startOffset),
			zap.Int64("end_offset", partition.LastOffset),
		)

		events, err := r.replayPartition(ctx, partition, startOffset)
		if err != nil {
			zap.L().Warn("failed to replay partition",
				zap.Int("partition", partition.ID),
				zap.Error(err),
			)
			continue
		}

		totalEvents += events
	}

	zap.L().Info("replay completed across all partitions",
		zap.Int("total_events", totalEvents),
		zap.Duration("duration", time.Since(startTime)),
	)

	return nil
}

func (r *ProductStateRecovery) getReplayStartOffset(snapshot *SnapshotInfo, partitionID int) int64 {
	// Get the recorded offset for this partition from snapshot
	if snapshot.Data.PartitionOffsets != nil {
		if offset, ok := snapshot.Data.PartitionOffsets[fmt.Sprintf("%d", partitionID)]; ok {
			return offset + 1 // Start from next offset after snapshot
		}
	}

	// Fallback: if partition offset not recorded, start from first offset
	// This shouldn't happen with proper snapshot generation
	zap.L().Warn("partition offset not found in snapshot, starting from beginning",
		zap.Int("partition", partitionID),
	)
	return 0
}

func (r *ProductStateRecovery) replayPartition(ctx context.Context, partition PartitionInfo, startOffset int64) (int, error) {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:   r.kafkaBrokers,
		Topic:     r.productTopic,
		Partition: partition.ID,
	})
	defer reader.Close()

	if err := reader.SetOffset(startOffset); err != nil {
		return 0, fmt.Errorf("failed to set offset: %w", err)
	}

	eventsProcessed := 0

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				break
			}
			continue
		}

		var event kafka.EventMessage
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			continue
		}

		r.applyIncrementalEvent(ctx, &event)
		eventsProcessed++

		if eventsProcessed%1000 == 0 {
			zap.L().Debug("replay progress",
				zap.Int("partition", partition.ID),
				zap.Int("events", eventsProcessed),
				zap.Int64("current_offset", msg.Offset),
			)
		}

		if msg.Offset >= partition.LastOffset-1 {
			break
		}
	}

	if eventsProcessed > 0 {
		zap.L().Info("partition replay completed",
			zap.Int("partition", partition.ID),
			zap.Int("events_processed", eventsProcessed),
		)
	}

	return eventsProcessed, nil
}

func (r *ProductStateRecovery) applyIncrementalEvent(ctx context.Context, event *kafka.EventMessage) {
	productID := r.extractProductID(event)
	if productID == "" {
		return
	}

	switch event.EventType {
	case "product.published":
		r.productStateRepo.MarkActive(ctx, productID)
	case "product.deactivated":
		r.productStateRepo.MarkInactive(ctx, productID)
	case "product.deleted":
		r.productStateRepo.Remove(ctx, productID)
	case "product.snapshot":
		// Skip snapshots during replay
	}
}

func (r *ProductStateRecovery) extractProductID(event *kafka.EventMessage) string {
	if event.AggregateID != "" && event.AggregateID != "__snapshot__" {
		return event.AggregateID
	}

	if productID, ok := event.Data["product_id"].(string); ok {
		return productID
	}

	return ""
}

func (r *ProductStateRecovery) clearCache(ctx context.Context) error {
	members, err := r.productStateRepo.GetAllActive(ctx)
	if err != nil {
		return err
	}

	for _, productID := range members {
		r.productStateRepo.Remove(ctx, productID)
	}

	zap.L().Info("cache cleared", zap.Int("removed", len(members)))
	return nil
}
