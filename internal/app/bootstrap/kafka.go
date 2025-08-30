package bootstrap

import (
	"context"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

// KafkaManager handles Kafka topic management from YAML configuration
type KafkaManager struct {
	topicManager *kafkaInfra.TopicManager
	kafkaConfig  config.KafkaConfig
}

// NewKafkaManager creates a new Kafka manager using environment variables
func NewKafkaManager(kafkaConfig config.KafkaConfig) (*KafkaManager, error) {
	topicManager := kafkaInfra.NewTopicManager(kafkaConfig.Brokers)
	return &KafkaManager{
		topicManager: topicManager,
		kafkaConfig:  kafkaConfig,
	}, nil
}

func (km *KafkaManager) EnsureAllTopics(ctx context.Context) error {
	for _, topic := range km.kafkaConfig.Topics {
		if err := km.EnsureTopicExists(ctx, topic); err != nil {
			return fmt.Errorf("failed to ensure topic %s exists: %w", topic.Name, err)
		}
	}
	return nil
}

// EnsureTopicExists ensures a topic exists, creates it if not
func (km *KafkaManager) EnsureTopicExists(ctx context.Context, topicConfig config.KafkaTopicConfig) error {
	// Check if topic already exists
	exists, err := km.topicManager.TopicExists(ctx, topicConfig.Name)
	if err != nil {
		return fmt.Errorf("failed to check if topic exists: %w", err)
	}
	if exists {
		return nil
	}

	logger.Info("Creating Kafka topic", map[string]interface{}{
		"topic":       topicConfig.Name,
		"partitions":  topicConfig.NumPartitions,
		"replication": topicConfig.ReplicationFactor,
	})

	return km.topicManager.CreateTopic(
		ctx,
		topicConfig.Name,
		topicConfig.NumPartitions,
		topicConfig.ReplicationFactor,
		topicConfig.ConfigEntries,
	)
}
