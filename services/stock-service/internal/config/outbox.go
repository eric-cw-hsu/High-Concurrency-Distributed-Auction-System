package config

import "time"

// OutboxConfig holds outbox pattern configuration
type OutboxConfig struct {
	PollInterval  time.Duration
	BatchSize     int
	MaxRetries    int
	CleanupAge    time.Duration
	CleanupPeriod time.Duration
}

// loadOutboxConfig loads outbox configuration
func loadOutboxConfig() OutboxConfig {
	return OutboxConfig{
		PollInterval:  getEnvDuration("OUTBOX_POLL_INTERVAL", 1*time.Second),
		BatchSize:     getEnvInt("OUTBOX_BATCH_SIZE", 100),
		MaxRetries:    getEnvInt("OUTBOX_MAX_RETRIES", 3),
		CleanupAge:    getEnvDuration("OUTBOX_CLEANUP_AGE", 7*24*time.Hour),    // 7 days
		CleanupPeriod: getEnvDuration("OUTBOX_CLEANUP_PERIOD", 1*24*time.Hour), // 1 day
	}
}
