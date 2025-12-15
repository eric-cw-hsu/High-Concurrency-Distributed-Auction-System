package metrics

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
)

// InventoryProcessorMetricsHelper is a thin wrapper around metrics collector for the inventory processor
type InventoryProcessorMetricsHelper struct {
	collector metrics.MetricsCollector
}

func NewInventoryProcessorMetricsHelper(collector metrics.MetricsCollector) *InventoryProcessorMetricsHelper {
	return &InventoryProcessorMetricsHelper{collector: collector}
}

func (h *InventoryProcessorMetricsHelper) IncProcessed() {
	h.collector.IncrementCounter("inventory_processor_processed_total", nil)
}

func (h *InventoryProcessorMetricsHelper) IncErrors() {
	h.collector.IncrementCounter("inventory_processor_errors_total", nil)
}

func (h *InventoryProcessorMetricsHelper) SetActiveInventoryConsumers(count int) {
	h.collector.SetGauge("inventory_processor_active_consumers", float64(count), nil)
}

func (h *InventoryProcessorMetricsHelper) RecordInventoryProcessingError(topic, errorType string) {
	h.collector.IncrementCounter("inventory_processor_processing_errors_total", map[string]string{
		"topic":      topic,
		"error_type": errorType,
	})
}

func (h *InventoryProcessorMetricsHelper) RecordInventoryMessageReceived(topic string) {
	h.collector.IncrementCounter("inventory_processor_messages_received_total", map[string]string{
		"topic": topic,
	})
}

func (h *InventoryProcessorMetricsHelper) RecordInventoryMessageProcessed(topic, action string) {
	h.collector.IncrementCounter("inventory_processor_messages_processed_total", map[string]string{
		"topic":  topic,
		"action": action,
	})
}
