package config

import "time"

type SnapshotConfig struct {
	Interval time.Duration
}

func loadSnapshotConfig() SnapshotConfig {
	return SnapshotConfig{
		Interval: getEnvDuration("SNAPSHOT_INTERVAL", 6*time.Hour),
	}
}
