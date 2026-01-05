package config

import "time"

// OutboxConfig holds outbox relay configuration
type OutboxConfig struct {
	BatchSize     int
	PollInterval  time.Duration
	CleanupAge    time.Duration
	CleanupPeriod time.Duration
}

func loadOutboxConfig() OutboxConfig {
	return OutboxConfig{
		BatchSize:     getEnvInt("OUTBOX_BATCH_SIZE", 100),
		PollInterval:  getEnvDuration("OUTBOX_POLL_INTERVAL", 1*time.Second),
		CleanupAge:    getEnvDuration("OUTBOX_CLEANUP_AGE", 7*24*time.Hour),    // 7 days
		CleanupPeriod: getEnvDuration("OUTBOX_CLEANUP_PERIOD", 1*24*time.Hour), // 1 day
	}
}
