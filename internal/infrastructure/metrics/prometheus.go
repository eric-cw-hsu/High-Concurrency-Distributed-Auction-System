package metrics

import (
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	sharedMetrics "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusCollector implements MetricsCollector using Prometheus
type PrometheusCollector struct {
	counters     map[string]*prometheus.CounterVec
	histograms   map[string]*prometheus.HistogramVec
	gauges       map[string]*prometheus.GaugeVec
	registry     *prometheus.Registry
	allLabelKeys map[string]map[string]bool // metric name -> set of label keys
}

// NewPrometheusCollector creates a new Prometheus-based metrics collector
func NewPrometheusCollector() *PrometheusCollector {
	return &PrometheusCollector{
		counters:     make(map[string]*prometheus.CounterVec),
		histograms:   make(map[string]*prometheus.HistogramVec),
		gauges:       make(map[string]*prometheus.GaugeVec),
		registry:     prometheus.NewRegistry(),
		allLabelKeys: make(map[string]map[string]bool),
	}
} // GetRegistry returns the Prometheus registry for server setup
func (p *PrometheusCollector) GetRegistry() *prometheus.Registry {
	return p.registry
}

// getOrCreateCounter gets or creates a counter metric
func (p *PrometheusCollector) getOrCreateCounter(name string, labelKeys []string) *prometheus.CounterVec {
	if counter, exists := p.counters[name]; exists {
		// Check if label keys are consistent with previously registered metric
		existingKeysSet := p.allLabelKeys[name]
		newKeysSet := make(map[string]bool)
		for _, key := range labelKeys {
			newKeysSet[key] = true
		}

		if !equalKeySets(existingKeysSet, newKeysSet) {
			existingKeys := setToSlice(existingKeysSet)
			logger.Error("Metric already exists with different label keys", map[string]interface{}{
				"metric_name":   name,
				"existing_keys": existingKeys,
				"new_keys":      labelKeys,
				"metric_type":   "counter",
			})
			return nil // Return nil to indicate error
		}
		return counter
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: "Counter metric for " + name,
	}, labelKeys)

	p.counters[name] = counter

	// Store label keys in set format
	labelKeysSet := make(map[string]bool)
	for _, key := range labelKeys {
		labelKeysSet[key] = true
	}
	p.allLabelKeys[name] = labelKeysSet

	p.registry.MustRegister(counter)
	return counter
}

// getOrCreateHistogram gets or creates a histogram metric
func (p *PrometheusCollector) getOrCreateHistogram(name string, labelKeys []string) *prometheus.HistogramVec {
	if histogram, exists := p.histograms[name]; exists {
		// Check if label keys are consistent with previously registered metric
		existingKeysSet := p.allLabelKeys[name]
		newKeysSet := make(map[string]bool)
		for _, key := range labelKeys {
			newKeysSet[key] = true
		}

		if !equalKeySets(existingKeysSet, newKeysSet) {
			existingKeys := setToSlice(existingKeysSet)
			logger.Error("Histogram metric already exists with different label keys", map[string]interface{}{
				"metric_name":   name,
				"existing_keys": existingKeys,
				"new_keys":      labelKeys,
				"metric_type":   "histogram",
			})
			return nil // Return nil to indicate error
		}
		return histogram
	}

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    "Histogram metric for " + name,
		Buckets: prometheus.DefBuckets,
	}, labelKeys)

	p.histograms[name] = histogram

	// Store label keys in set format
	labelKeysSet := make(map[string]bool)
	for _, key := range labelKeys {
		labelKeysSet[key] = true
	}
	p.allLabelKeys[name] = labelKeysSet

	p.registry.MustRegister(histogram)
	return histogram
}

// getOrCreateGauge gets or creates a gauge metric
func (p *PrometheusCollector) getOrCreateGauge(name string, labelKeys []string) *prometheus.GaugeVec {
	if gauge, exists := p.gauges[name]; exists {
		// Check if label keys are consistent with previously registered metric
		existingKeysSet := p.allLabelKeys[name]
		newKeysSet := make(map[string]bool)
		for _, key := range labelKeys {
			newKeysSet[key] = true
		}

		if !equalKeySets(existingKeysSet, newKeysSet) {
			existingKeys := setToSlice(existingKeysSet)
			logger.Error("Gauge metric already exists with different label keys", map[string]interface{}{
				"metric_name":   name,
				"existing_keys": existingKeys,
				"new_keys":      labelKeys,
				"metric_type":   "gauge",
			})
			return nil // Return nil to indicate error
		}
		return gauge
	}

	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: "Gauge metric for " + name,
	}, labelKeys)

	p.gauges[name] = gauge

	// Store label keys in set format
	labelKeysSet := make(map[string]bool)
	for _, key := range labelKeys {
		labelKeysSet[key] = true
	}
	p.allLabelKeys[name] = labelKeysSet

	p.registry.MustRegister(gauge)
	return gauge
}

// extractLabelKeys extracts keys from labels map
func extractLabelKeys(labels map[string]string) []string {
	if labels == nil {
		return []string{}
	}

	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	return keys
}

// equalKeySets checks if two key sets are equal
func equalKeySets(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}

	for key := range a {
		if !b[key] {
			return false
		}
	}
	return true
}

// setToSlice converts a key set to a slice
func setToSlice(keySet map[string]bool) []string {
	result := make([]string, 0, len(keySet))
	for key := range keySet {
		result = append(result, key)
	}
	return result
}

// MetricsCollector interface implementation

func (p *PrometheusCollector) IncrementCounter(name string, labels map[string]string) {
	p.IncrementCounterWithValue(name, 1, labels)
}

func (p *PrometheusCollector) IncrementCounterWithValue(name string, value float64, labels map[string]string) {
	labelKeys := extractLabelKeys(labels)
	counter := p.getOrCreateCounter(name, labelKeys)
	if counter != nil {
		counter.With(labels).Add(value)
	}
}

func (p *PrometheusCollector) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	labelKeys := extractLabelKeys(labels)
	histogram := p.getOrCreateHistogram(name, labelKeys)
	if histogram != nil {
		histogram.With(labels).Observe(duration.Seconds())
	}
}

func (p *PrometheusCollector) SetGauge(name string, value float64, labels map[string]string) {
	labelKeys := extractLabelKeys(labels)
	gauge := p.getOrCreateGauge(name, labelKeys)
	if gauge != nil {
		gauge.With(labels).Set(value)
	}
}

func (p *PrometheusCollector) IncGauge(name string, labels map[string]string) {
	p.AddGauge(name, 1, labels)
}

func (p *PrometheusCollector) DecGauge(name string, labels map[string]string) {
	p.AddGauge(name, -1, labels)
}

func (p *PrometheusCollector) AddGauge(name string, value float64, labels map[string]string) {
	labelKeys := extractLabelKeys(labels)
	gauge := p.getOrCreateGauge(name, labelKeys)
	if gauge != nil {
		gauge.With(labels).Add(value)
	}
}

// Verify interface implementation
var _ sharedMetrics.MetricsCollector = (*PrometheusCollector)(nil)
