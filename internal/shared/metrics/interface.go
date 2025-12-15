package metrics

import "time"

// MetricsCollector is the main interface for recording metrics
// All microservices use this unified interface
type MetricsCollector interface {
	// IncrementCounter increments a counter metric by 1
	IncrementCounter(name string, labels map[string]string)

	// IncrementCounterWithValue increments a counter metric by a specific value
	IncrementCounterWithValue(name string, value float64, labels map[string]string)

	// RecordDuration records a duration metric (histogram)
	RecordDuration(name string, duration time.Duration, labels map[string]string)

	// SetGauge sets a gauge metric to a specific value
	SetGauge(name string, value float64, labels map[string]string)

	// IncGauge increments a gauge metric by 1
	IncGauge(name string, labels map[string]string)

	// DecGauge decrements a gauge metric by 1
	DecGauge(name string, labels map[string]string)

	// AddGauge adds a specific value to a gauge metric
	AddGauge(name string, value float64, labels map[string]string)
}

// MetricsServer interface for both standalone and embedded metrics servers
type MetricsServer interface {
	// Start starts the metrics server (for standalone servers)
	Start() error

	// Stop stops the metrics server gracefully
	Stop() error
}
