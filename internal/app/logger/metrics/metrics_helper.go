package metrics

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
)

// LoggerMetricsHelper provides logger-specific metrics tracking
type LoggerMetricsHelper struct {
	collector metrics.MetricsCollector
}

func NewLoggerMetricsHelper(collector metrics.MetricsCollector) *LoggerMetricsHelper {
	return &LoggerMetricsHelper{
		collector: collector,
	}
}

// RecordLogMessageReceived records a received log message from Kafka
func (h *LoggerMetricsHelper) RecordLogMessageReceived(topic string) {
	h.collector.IncrementCounter("kafka_messages_received_total", map[string]string{
		"service": "logger",
		"topic":   topic,
		"type":    "log_message",
	})
}

// RecordLogMessageProcessed records a successfully processed log message
func (h *LoggerMetricsHelper) RecordLogMessageProcessed(topic, level string) {
	h.collector.IncrementCounter("kafka_messages_processed_total", map[string]string{
		"service": "logger",
		"topic":   topic,
		"level":   level,
	})
}

// RecordLogProcessingError records a log processing error
func (h *LoggerMetricsHelper) RecordLogProcessingError(topic, errorType string) {
	h.collector.IncrementCounter("kafka_processing_errors_total", map[string]string{
		"service":    "logger",
		"topic":      topic,
		"error_type": errorType,
	})
}

// SetActiveLogConsumers sets the number of active logger consumers
func (h *LoggerMetricsHelper) SetActiveLogConsumers(count float64) {
	h.collector.SetGauge("kafka_active_consumers", count, map[string]string{
		"service": "logger",
	})
}

// RecordLogStorageOperation records log storage operations
func (h *LoggerMetricsHelper) RecordLogStorageOperation(operation string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	h.collector.IncrementCounter("storage_operations_total", map[string]string{
		"service":   "logger",
		"operation": operation,
		"status":    status,
	})
}
