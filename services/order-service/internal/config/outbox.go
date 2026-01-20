package config

import (
	"fmt"
	"time"
)

type OutboxConfig struct {
	Interval  time.Duration
	BatchSize int
}

func loadOutboxConfig() OutboxConfig {
	return OutboxConfig{
		Interval:  getEnvDuration("OUTBOX_RELAY_INTERVAL", 500*time.Millisecond),
		BatchSize: getEnvInt("OUTBOX_BATCH_SIZE", 50),
	}
}

func (c *OutboxConfig) Validate() error {
	if c.Interval <= 0 {
		return fmt.Errorf("outbox_relay_interval must be positive")
	}
	return nil
}
