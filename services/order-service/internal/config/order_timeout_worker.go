package config

import (
	"fmt"
	"time"
)

type OrderTimeoutWorkerConfig struct {
	CheckInterval time.Duration
	BatchSize     int
}

func loadOrderTimeoutWorkerConfig() OrderTimeoutWorkerConfig {
	return OrderTimeoutWorkerConfig{
		CheckInterval: getEnvDuration("TIMEOUT_CHECK_INTERVAL", 10*time.Second),
		BatchSize:     getEnvInt("TIMEOUT_BATCH_SIZE", 100),
	}
}

func (c *OrderTimeoutWorkerConfig) Validate() error {
	if c.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive")
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be positive")
	}
	return nil
}
