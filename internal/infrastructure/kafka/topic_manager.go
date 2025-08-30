package kafka

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

// TopicManager manages Kafka topics creation and configuration
type TopicManager struct {
	brokers []string
	conn    *kafka.Conn
}

// NewTopicManager creates a new topic manager
func NewTopicManager(brokers []string) *TopicManager {
	return &TopicManager{
		brokers: brokers,
	}
}

// Connect establishes connection to Kafka cluster
func (tm *TopicManager) Connect(ctx context.Context) error {
	if len(tm.brokers) == 0 {
		return fmt.Errorf("no brokers specified")
	}

	conn, err := kafka.DialLeader(ctx, "tcp", tm.brokers[0], "", 0)
	if err != nil {
		return fmt.Errorf("failed to connect to kafka broker: %w", err)
	}

	tm.conn = conn
	return nil
}

// Close closes the connection to Kafka
func (tm *TopicManager) Close() error {
	if tm.conn != nil {
		return tm.conn.Close()
	}
	return nil
}

// CreateTopic creates a new topic with the specified configuration
func (tm *TopicManager) CreateTopic(ctx context.Context, name string, numPartitions int, replicationFactor int, configEntries map[string]string) error {
	if tm.conn == nil {
		if err := tm.Connect(ctx); err != nil {
			return err
		}
	}

	// Check if topic already exists
	exists, err := tm.TopicExists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if exists {
		log.Printf("Topic %s already exists, skipping creation", name)
		return nil
	}

	// Set default values if not specified
	if numPartitions <= 0 {
		numPartitions = 1
	}
	if replicationFactor <= 0 {
		replicationFactor = 1
	}

	// Create topic
	controller, err := tm.conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, fmt.Sprintf("%d", controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	// Create topic request
	configEntriesKafka := make([]kafka.ConfigEntry, 0, len(configEntries))
	for key, value := range configEntries {
		configEntriesKafka = append(configEntriesKafka, kafka.ConfigEntry{
			ConfigName:  key,
			ConfigValue: value,
		})
	}

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             name,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
			ConfigEntries:     configEntriesKafka,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("failed to create topic %s: %w", name, err)
	}

	log.Printf("Successfully created topic: %s (partitions: %d, replication: %d)",
		name, numPartitions, replicationFactor)

	return nil
}

// TopicExists checks if a topic exists
func (tm *TopicManager) TopicExists(ctx context.Context, topicName string) (bool, error) {
	if tm.conn == nil {
		if err := tm.Connect(ctx); err != nil {
			return false, err
		}
	}

	partitions, err := tm.conn.ReadPartitions(topicName)
	if err != nil {
		// If error contains "unknown topic", it means topic doesn't exist
		if fmt.Sprintf("%v", err) == "unknown topic or partition" {
			return false, nil
		}
		return false, fmt.Errorf("failed to read partitions for topic %s: %w", topicName, err)
	}

	return len(partitions) > 0, nil
}

// DeleteTopic deletes a topic (use with caution in production)
func (tm *TopicManager) DeleteTopic(ctx context.Context, topicName string) error {
	if tm.conn == nil {
		if err := tm.Connect(ctx); err != nil {
			return err
		}
	}

	controller, err := tm.conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, fmt.Sprintf("%d", controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	err = controllerConn.DeleteTopics(topicName)
	if err != nil {
		return fmt.Errorf("failed to delete topic %s: %w", topicName, err)
	}

	log.Printf("Successfully deleted topic: %s", topicName)
	return nil
}

// WaitForTopicReady waits for a topic to be ready for use
func (tm *TopicManager) WaitForTopicReady(ctx context.Context, topicName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		exists, err := tm.TopicExists(ctx, topicName)
		if err != nil {
			log.Printf("Error checking topic %s: %v", topicName, err)
		} else if exists {
			log.Printf("Topic %s is ready", topicName)
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
			// Continue checking
		}
	}

	return fmt.Errorf("timeout waiting for topic %s to be ready", topicName)
}
