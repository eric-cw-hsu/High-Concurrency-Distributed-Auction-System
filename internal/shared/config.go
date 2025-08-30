package shared

import (
	"os"
	"strings"
)

// GetEnv gets environment variable with fallback
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetFileDir extracts directory from file path
func GetFileDir(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "/")
	}
	return "/app/logs"
}

// ParseTopics parses comma-separated topics from environment variable
func ParseTopics(topicsEnv string) []string {
	if topicsEnv == "" {
		return []string{}
	}

	topics := strings.Split(topicsEnv, ",")
	result := make([]string, 0, len(topics))

	for _, topic := range topics {
		topic = strings.TrimSpace(topic)
		if topic != "" {
			result = append(result, topic)
		}
	}

	return result
}
