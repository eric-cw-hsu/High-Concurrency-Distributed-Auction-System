package config

import (
	"fmt"
	"time"
)

// ServiceConfig holds business logic configuration
type ServiceConfig struct {
	PersistQueueSize   int
	PersistBatchSize   int
	PersistFlushWindow time.Duration
	LowStockThreshold  float64 // percentage, e.g., 0.1 = 10%
}

// loadServiceConfig loads service configuration
func loadServiceConfig() ServiceConfig {
	return ServiceConfig{
		PersistQueueSize:   getEnvInt("SERVICE_PERSIST_QUEUE_SIZE", 10000),
		PersistBatchSize:   getEnvInt("SERVICE_PERSIST_BATCH_SIZE", 1000),
		PersistFlushWindow: getEnvDuration("SERVICE_PERSIST_FLUSH_WINDOW", 100*time.Millisecond),
		LowStockThreshold:  getEnvFloat("SERVICE_LOW_STOCK_THRESHOLD", 0.1),
	}
}

// Validate validates service configuration
func (c *ServiceConfig) Validate() error {
	if c.PersistQueueSize <= 0 {
		return fmt.Errorf("persist_queue_size must be positive")
	}
	if c.PersistBatchSize <= 0 {
		return fmt.Errorf("persist_batch_size must be positive")
	}
	if c.LowStockThreshold < 0 || c.LowStockThreshold > 1 {
		return fmt.Errorf("low_stock_threshold must be between 0 and 1")
	}
	return nil
}
