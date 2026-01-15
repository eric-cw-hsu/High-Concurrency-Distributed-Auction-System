package config

import (
	"fmt"
	"time"
)

// ScannerConfig holds reservation scanner configuration
type ExpiredReservationScannerConfig struct {
	ScanInterval time.Duration
	TimeWindow   time.Duration
	BatchSize    int
}

func loadExpiredReservationScannerConfig() ExpiredReservationScannerConfig {
	return ExpiredReservationScannerConfig{
		ScanInterval: getEnvDuration("SCANNER_SCAN_INTERVAL", 5*time.Minute),
		TimeWindow:   getEnvDuration("SCANNER_TIME_WINDOW", 1*time.Hour),
		BatchSize:    getEnvInt("SCANNER_BATCH_SIZE", 100),
	}
}

func (c *ExpiredReservationScannerConfig) Validate() error {
	if c.ScanInterval <= 0 {
		return fmt.Errorf("scan_interval must be positive")
	}
	if c.TimeWindow <= 0 {
		return fmt.Errorf("time_window must be positive")
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be positive")
	}
	return nil
}
