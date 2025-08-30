package testutil

import (
	"context"
	"fmt"
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
)

// KafkaTestHelper provides utilities for managing Kafka topics in tests
type KafkaTestHelper struct {
	kafkaManager  *bootstrap.KafkaManager
	topicManager  *kafkaInfra.TopicManager
	brokers       []string
	ctx           context.Context
	createdTopics []string // Track topics created for cleanup
}

// NewKafkaTestHelper creates a new Kafka test helper
func NewKafkaTestHelper(ctx context.Context, brokers []string) (*KafkaTestHelper, error) {
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"} // Default to local Kafka
	}

	// Create test-specific config with shorter retention for cleanup
	testConfig := config.KafkaConfig{
		Brokers: brokers,
		Topics:  []config.KafkaTopicConfig{}, // Start with empty topics for tests
	}

	kafkaManager, err := bootstrap.NewKafkaManager(testConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka manager: %w", err)
	}

	topicManager := kafkaInfra.NewTopicManager(brokers)

	// Test connection
	if err := topicManager.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka brokers: %w", err)
	}

	return &KafkaTestHelper{
		kafkaManager:  kafkaManager,
		topicManager:  topicManager,
		brokers:       brokers,
		ctx:           ctx,
		createdTopics: make([]string, 0),
	}, nil
}

// CreateTestTopic creates a test topic with a unique name
func (h *KafkaTestHelper) CreateTestTopic(baseName string) (string, error) {
	// Create unique topic name for test isolation
	topicName := fmt.Sprintf("%s.test.%d", baseName, time.Now().UnixNano())

	configEntries := map[string]string{
		"retention.ms":   "300000", // 5 minutes for tests
		"cleanup.policy": "delete",
	}

	err := h.topicManager.CreateTopic(
		h.ctx,
		topicName,
		1, // numPartitions
		1, // replicationFactor
		configEntries,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create test topic %s: %w", topicName, err)
	}

	// Wait for topic to be ready
	if err := h.topicManager.WaitForTopicReady(h.ctx, topicName, 10*time.Second); err != nil {
		return "", fmt.Errorf("test topic %s not ready: %w", topicName, err)
	}

	// Track for cleanup
	h.createdTopics = append(h.createdTopics, topicName)

	return topicName, nil
} // CreateOrderPlacedTestTopic creates a test topic specifically for order.placed events
func (h *KafkaTestHelper) CreateOrderPlacedTestTopic() (string, error) {
	return h.CreateTestTopic("order.placed")
}

// CreateStockUpdatedTestTopic creates a test topic specifically for stock.updated events
func (h *KafkaTestHelper) CreateStockUpdatedTestTopic() (string, error) {
	return h.CreateTestTopic("stock.updated")
}

// InitializeSystemTopics creates all predefined system topics (for integration tests)
func (h *KafkaTestHelper) InitializeSystemTopics() error {
	return h.kafkaManager.EnsureAllTopics(h.ctx)
}

// TopicExists checks if a topic exists
func (h *KafkaTestHelper) TopicExists(topicName string) (bool, error) {
	return h.topicManager.TopicExists(h.ctx, topicName)
}

// CleanupTestTopics deletes all test topics created by this helper
func (h *KafkaTestHelper) CleanupTestTopics() error {
	var errors []error

	for _, topicName := range h.createdTopics {
		if err := h.topicManager.DeleteTopic(h.ctx, topicName); err != nil {
			errors = append(errors, fmt.Errorf("failed to delete topic %s: %w", topicName, err))
		}
	}

	// Clear the list
	h.createdTopics = make([]string, 0)

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// Close closes the Kafka connections
func (h *KafkaTestHelper) Close() error {
	// First cleanup test topics
	if err := h.CleanupTestTopics(); err != nil {
		// Log error but don't fail close operation
		fmt.Printf("Warning: failed to cleanup test topics: %v\n", err)
	}

	// Close topic manager connections
	return h.topicManager.Close()
}

// GetBrokers returns the configured brokers
func (h *KafkaTestHelper) GetBrokers() []string {
	return h.brokers
}
