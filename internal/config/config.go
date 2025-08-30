package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvMust(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	log.Fatalf("Environment variable %s is required but not set", key)

	return ""
}

func parseSeparatedList(list, separator string) []string {
	var result []string

	for _, item := range strings.Split(list, separator) {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}

func getIntFromMap(data map[string]interface{}, key string, defaultValue int) int {
	if value, exists := data[key]; exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return defaultValue
}

func getMapFromMap(data map[string]interface{}, key string) map[string]string {
	result := make(map[string]string)
	if value, exists := data[key]; exists {
		if mapValue, ok := value.(map[interface{}]interface{}); ok {
			for k, v := range mapValue {
				strKey, keyOk := k.(string)
				strValue, valueOk := v.(string)
				if keyOk && valueOk {
					result[strKey] = strValue
				}
			}
		}
	}
	return result
}
