package config

import (
	"errors"
	"strings"
	"time"
)

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers              []string
	ProducerTopic        string
	ConsumerTopic        string
	ConsumerGroupID      string
	ProducerMaxAttempts  int
	ProducerBatchSize    int
	ProducerBatchTimeout time.Duration
}

func loadKafkaConfig() KafkaConfig {
	brokersStr := getEnv("KAFKA_BROKERS", "localhost:9092")
	brokers := strings.Split(brokersStr, ",")

	return KafkaConfig{
		Brokers:              brokers,
		ProducerTopic:        getEnv("KAFKA_PRODUCER_TOPIC", "product-events"),
		ConsumerTopic:        getEnv("KAFKA_CONSUMER_TOPIC", "stock-events"),
		ConsumerGroupID:      getEnv("KAFKA_CONSUMER_GROUP_ID", "product-service-consumer"),
		ProducerMaxAttempts:  getEnvInt("KAFKA_PRODUCER_MAX_ATTEMPTS", 3),
		ProducerBatchSize:    getEnvInt("KAFKA_PRODUCER_BATCH_SIZE", 100),
		ProducerBatchTimeout: getEnvDuration("KAFKA_PRODUCER_BATCH_TIMEOUT", 10*time.Millisecond),
	}
}

func (c KafkaConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("kafka brokers are required")
	}
	if c.ProducerTopic == "" {
		return errors.New("kafka producer topic is required")
	}
	if c.ConsumerTopic == "" {
		return errors.New("kafka consumer topic is required")
	}
	if c.ConsumerGroupID == "" {
		return errors.New("kafka consumer group ID is required")
	}
	return nil
}
