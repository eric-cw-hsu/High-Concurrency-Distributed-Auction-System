package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	kafkaProducer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/producer"
)

// KafkaEventOrderingTestSuite tests event ordering with real Kafka infrastructure
type KafkaEventOrderingTestSuite struct {
	suite.Suite
	producer        *kafkaProducer.OrderProducer
	ctx             context.Context
	topicName       string
	consumerGroup   string
	processedEvents []ProcessedEvent
}

type ProcessedEvent struct {
	OrderId     string
	StockId     string
	Quantity    int
	Timestamp   time.Time
	ProcessedAt time.Time
}

func (suite *KafkaEventOrderingTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.topicName = "test-event-ordering-clean" // Use a clean topic name
	suite.consumerGroup = "test-ordering-consumer"
	suite.processedEvents = make([]ProcessedEvent, 0)

	// Setup Kafka producer
	writer := kafkaInfra.NewWriter([]string{"localhost:9092"}, suite.topicName)
	suite.producer = kafkaProducer.NewOrderProducer(writer).(*kafkaProducer.OrderProducer)
}

func (suite *KafkaEventOrderingTestSuite) SetupTest() {
	suite.processedEvents = make([]ProcessedEvent, 0)
}

func (suite *KafkaEventOrderingTestSuite) TestKafka_SimpleConnectionTest() {
	// Simple test to check if we can create a topic and write to it
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		suite.T().Skip("Kafka not available")
		return
	}
	defer conn.Close()

	// Try to create topic
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             suite.topicName,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		suite.T().Logf("Topic creation result: %v", err)
	}

	suite.T().Logf("✅ Kafka connection successful")
}

func (suite *KafkaEventOrderingTestSuite) TestKafkaEventOrdering_TimestampSequence() {
	// Core test: Are events processed in timestamp order or publish order?

	// First ensure topic exists
	suite.createTopicIfNotExists()

	stockId := "timestamp-ordering-test"
	baseTime := time.Now()

	// Create events with INTENTIONALLY different publish vs timestamp order
	events := []order.OrderPlacedEvent{
		{
			OrderId:    "order-001",
			StockId:    stockId,
			BuyerId:    "buyer-1",
			Quantity:   10,
			TotalPrice: 500.0,
			Timestamp:  baseTime.Add(3 * time.Second), // Latest timestamp (published first)
		},
		{
			OrderId:    "order-002",
			StockId:    stockId,
			BuyerId:    "buyer-2",
			Quantity:   15,
			TotalPrice: 750.0,
			Timestamp:  baseTime.Add(1 * time.Second), // Earliest timestamp (published second)
		},
		{
			OrderId:    "order-003",
			StockId:    stockId,
			BuyerId:    "buyer-3",
			Quantity:   5,
			TotalPrice: 250.0,
			Timestamp:  baseTime.Add(2 * time.Second), // Middle timestamp (published third)
		},
	}

	// Publish events in publish-time order (NOT timestamp order)
	for _, event := range events {
		err := suite.producer.PublishEvent(suite.ctx, &event)
		require.NoError(suite.T(), err)
		suite.T().Logf("Published: OrderId=%s, Timestamp=%v", event.OrderId, event.Timestamp)
	}

	// Setup reader to consume messages
	reader := kafkaInfra.NewReader([]string{"localhost:9092"}, suite.topicName, suite.consumerGroup)
	defer reader.Close()

	// Give Kafka a moment to propagate the messages
	time.Sleep(2 * time.Second)

	// Process all events
	ctx, cancel := context.WithTimeout(suite.ctx, 15*time.Second)
	defer cancel()

	for i := 0; i < len(events); i++ {
		if err := suite.processOneMessage(ctx, reader); err != nil {
			suite.T().Fatalf("Error processing message %d: %v", i+1, err)
		}
	}

	// CORE VERIFICATION: Events should be processed in PUBLISH order, NOT timestamp order
	require.Len(suite.T(), suite.processedEvents, len(events), "All events should be processed")

	// Verify processing order matches publish order (Kafka partition ordering guarantee)
	assert.Equal(suite.T(), "order-001", suite.processedEvents[0].OrderId, "First processed should be first published")
	assert.Equal(suite.T(), "order-002", suite.processedEvents[1].OrderId, "Second processed should be second published")
	assert.Equal(suite.T(), "order-003", suite.processedEvents[2].OrderId, "Third processed should be third published")

	// Verify timestamps are preserved but don't affect processing order
	assert.True(suite.T(), suite.processedEvents[0].Timestamp.After(suite.processedEvents[1].Timestamp),
		"First processed event has later timestamp than second")
	assert.True(suite.T(), suite.processedEvents[1].Timestamp.Before(suite.processedEvents[2].Timestamp),
		"Second processed event has earlier timestamp than third")

	suite.T().Logf("✅ Kafka guarantees partition ordering (publish order), not timestamp ordering")
}

// Helper methods

func (suite *KafkaEventOrderingTestSuite) createTopicIfNotExists() {
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		suite.T().Skip("Kafka not available")
		return
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		suite.T().Logf("Failed to read partitions: %v", err)
		return
	}

	// Check if topic already exists
	topicExists := false
	for _, p := range partitions {
		if p.Topic == suite.topicName {
			topicExists = true
			break
		}
	}

	if !topicExists {
		suite.T().Logf("Creating topic: %s", suite.topicName)
		topicConfigs := []kafka.TopicConfig{
			{
				Topic:             suite.topicName,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		}

		err = conn.CreateTopics(topicConfigs...)
		if err != nil {
			suite.T().Logf("Failed to create topic: %v", err)
		} else {
			suite.T().Logf("Topic %s created successfully", suite.topicName)
		}
	}
}

func (suite *KafkaEventOrderingTestSuite) processOneMessage(ctx context.Context, reader *kafka.Reader) error {
	message, err := reader.ReadMessage(ctx)
	if err != nil {
		return err
	}

	suite.T().Logf("Raw message: Key=%s, Value=%s", string(message.Key), string(message.Value))

	// Parse the event
	var orderEvent order.OrderPlacedEvent
	if err := json.Unmarshal(message.Value, &orderEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w, raw: %s", err, string(message.Value))
	}

	// Record processing
	processed := ProcessedEvent{
		OrderId:     orderEvent.OrderId,
		StockId:     orderEvent.StockId,
		Quantity:    orderEvent.Quantity,
		Timestamp:   orderEvent.Timestamp,
		ProcessedAt: time.Now(),
	}

	suite.processedEvents = append(suite.processedEvents, processed)

	suite.T().Logf("Processed: OrderId=%s, Timestamp=%v", processed.OrderId, processed.Timestamp)

	// Commit the message
	return reader.CommitMessages(ctx, message)
}

// TestKafkaEventOrderingTestSuite runs the test suite
func TestKafkaEventOrderingTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaEventOrderingTestSuite))
}
