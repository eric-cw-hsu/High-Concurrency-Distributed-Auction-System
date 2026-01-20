package config

import (
	"fmt"
	"strings"
	"time"
)

type KafkaConfig struct {
	Brokers         []string
	ConsumerGroupID string
	ProducerTopic   string
	ConsumerTopic   string

	// Producer specific configs
	ProducerMaxAttempts  int
	ProducerBatchSize    int
	ProducerBatchTimeout time.Duration

	// Consumer specific configs
	ConsumerMinBytes int
	ConsumerMaxBytes int
	ConsumerMaxWait  time.Duration
}

func loadKafkaConfig() KafkaConfig {
	return KafkaConfig{
		// Expected format: "localhost:9092,localhost:9093"
		Brokers:         strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		ConsumerGroupID: getEnv("KAFKA_CONSUMER_GROUP_ID", "order-service-group"),
		ProducerTopic:   getEnv("KAFKA_PRODUCER_TOPIC", "order-events"),
		ConsumerTopic:   getEnv("KAFKA_CONSUMER_TOPIC", "stock-events"),

		// Tuning for reliability and throughput
		ProducerMaxAttempts:  getEnvInt("KAFKA_PRODUCER_MAX_ATTEMPTS", 5),
		ProducerBatchSize:    getEnvInt("KAFKA_PRODUCER_BATCH_SIZE", 100),
		ProducerBatchTimeout: getEnvDuration("KAFKA_PRODUCER_BATCH_TIMEOUT", 10*time.Millisecond),

		// Consumer performance tuning
		ConsumerMinBytes: getEnvInt("KAFKA_CONSUMER_MIN_BYTES", 1e3),  // 1KB
		ConsumerMaxBytes: getEnvInt("KAFKA_CONSUMER_MAX_BYTES", 10e6), // 10MB
		ConsumerMaxWait:  getEnvDuration("KAFKA_CONSUMER_MAX_WAIT", 100*time.Millisecond),
	}
}

func (c *KafkaConfig) Validate() error {
	if len(c.Brokers) == 0 || c.Brokers[0] == "" {
		return fmt.Errorf("kafka brokers must be specified")
	}
	if c.ProducerTopic == "" {
		return fmt.Errorf("kafka producer topic is required")
	}
	if c.ConsumerGroupID == "" {
		return fmt.Errorf("kafka consumer group id is required")
	}
	if c.ProducerMaxAttempts <= 0 {
		return fmt.Errorf("producer_max_attempts must be positive")
	}
	return nil
}
