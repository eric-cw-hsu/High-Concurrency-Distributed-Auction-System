package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// KafkaTopicConfig represents the configuration for a single Kafka topic
type KafkaTopicConfig struct {
	Name              string            `yaml:"name" json:"name"`
	NumPartitions     int               `yaml:"partitions" json:"partitions"`
	ReplicationFactor int               `yaml:"replication_factor" json:"replication_factor"`
	ConfigEntries     map[string]string `yaml:"config" json:"config"`
}

type KafkaConfig struct {
	Brokers []string           `yaml:"brokers" json:"brokers"`
	Topics  []KafkaTopicConfig `yaml:"topics" json:"topics"`
}

// createDefaultTopicConfig creates a default topic configuration
func createDefaultTopicConfig(topicName string) KafkaTopicConfig {
	return KafkaTopicConfig{
		Name:              topicName,
		NumPartitions:     3,
		ReplicationFactor: 1,
		ConfigEntries:     make(map[string]string),
	}
}

// createTopicConfigFromYAML creates topic configuration from YAML data
func createTopicConfigFromYAML(topicName string, topicData map[string]interface{}) KafkaTopicConfig {
	return KafkaTopicConfig{
		Name:              topicName,
		NumPartitions:     getIntFromMap(topicData, "num_partitions", 3),
		ReplicationFactor: getIntFromMap(topicData, "replication_factor", 1),
		ConfigEntries:     getMapFromMap(topicData, "config"),
	}
}

// processTopicList processes a list of topic names and creates configurations
func processTopicList(topicNames []string, topicsMap map[string]map[string]interface{}) []KafkaTopicConfig {
	var topics []KafkaTopicConfig
	for _, topicName := range topicNames {
		if topicData, exists := topicsMap[topicName]; exists {
			topics = append(topics, createTopicConfigFromYAML(topicName, topicData))
		} else {
			topics = append(topics, createDefaultTopicConfig(topicName))
		}
	}
	return topics
}

func loadKafkaTopicConfigFromYAML(producedTopics, consumedTopics []string) []KafkaTopicConfig {
	// Try to load from YAML file
	configPaths := []string{
		"configs/topics.yaml",
		"./topics.yaml",
		"../topics.yaml",
	}

	var configData []byte
	var err error

	for _, path := range configPaths {
		if configData, err = os.ReadFile(path); err == nil {
			fmt.Printf("Loaded topic config from: %s\n", path)
			break
		}
	}

	if err != nil {
		fmt.Println("No topic config file found, using empty config")
		return []KafkaTopicConfig{}
	}

	// Parse YAML as map[string]interface{} first to handle dynamic topic names
	var topicsMap map[string]map[string]interface{}
	if err := yaml.Unmarshal(configData, &topicsMap); err != nil {
		fmt.Printf("Failed to parse topics YAML: %v\n", err)
		return []KafkaTopicConfig{}
	}

	// Process all topics and avoid duplicates
	allTopics := make(map[string]bool)
	var topics []KafkaTopicConfig

	// Add produced topics
	for _, config := range processTopicList(producedTopics, topicsMap) {
		if !allTopics[config.Name] {
			topics = append(topics, config)
			allTopics[config.Name] = true
		}
	}

	// Add consumed topics (skip if already added)
	for _, config := range processTopicList(consumedTopics, topicsMap) {
		if !allTopics[config.Name] {
			topics = append(topics, config)
			allTopics[config.Name] = true
		}
	}

	return topics
}

// LoadKafkaConfig loads Kafka configuration from environment or defaults
func LoadKafkaConfig() KafkaConfig {
	producedTopics := parseSeparatedList(getEnv("KAFKA_PRODUCED_TOPICS", ""), ",")
	consumedTopics := parseSeparatedList(getEnv("KAFKA_CONSUMED_TOPICS", ""), ",")

	return KafkaConfig{
		Brokers: parseSeparatedList(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		Topics:  loadKafkaTopicConfigFromYAML(producedTopics, consumedTopics),
	}
}
