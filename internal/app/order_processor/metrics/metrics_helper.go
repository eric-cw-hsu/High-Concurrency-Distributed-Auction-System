package metrics

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
)

// OrderProcessorMetricsHelper is a thin wrapper around metrics collector for the order processor
type OrderProcessorMetricsHelper struct {
	collector metrics.MetricsCollector
}

func NewOrderProcessorMetricsHelper(collector metrics.MetricsCollector) *OrderProcessorMetricsHelper {
	return &OrderProcessorMetricsHelper{collector: collector}
}

func (h *OrderProcessorMetricsHelper) IncProcessed() {
	h.collector.IncrementCounter("order_processor_processed_total", nil)
}

func (h *OrderProcessorMetricsHelper) IncErrors() {
	h.collector.IncrementCounter("order_processor_errors_total", nil)
}

func (h *OrderProcessorMetricsHelper) SetActiveProcessorOrderConsumers(count int) {
	h.collector.SetGauge("order_processor_active_consumers", float64(count), nil)
}

func (h *OrderProcessorMetricsHelper) RecordOrderProcessingError(eventType, errorType string) {
	h.collector.IncrementCounter("order_processor_event_errors_total", map[string]string{
		"event_type": eventType,
		"error_type": errorType,
	})
}

func (h *OrderProcessorMetricsHelper) RecordOrderMessageReceived(eventType string) {
	h.collector.IncrementCounter("order_processor_messages_received_total", map[string]string{
		"event_type": eventType,
	})
}

func (h *OrderProcessorMetricsHelper) RecordOrderMessageProcessed(topic string, eventType string) {
	h.collector.IncrementCounter("order_processor_messages_processed_total", map[string]string{
		"topic":      topic,
		"event_type": eventType,
	})
}
