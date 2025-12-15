package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sharedMetrics "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusCollector(t *testing.T) {
	t.Run("Create collector", func(t *testing.T) {
		collector := NewPrometheusCollector()
		assert.NotNil(t, collector)
		assert.NotNil(t, collector.GetRegistry())
	})

	t.Run("Increment counter", func(t *testing.T) {
		collector := NewPrometheusCollector()

		// Test basic counter increment
		collector.IncrementCounter("test_counter", map[string]string{
			"label1": "value1",
		})

		// Test counter increment with value
		collector.IncrementCounterWithValue("test_counter_with_value", 5.0, map[string]string{
			"label1": "value1",
		})

		// Should not panic
		assert.NotNil(t, collector)
	})

	t.Run("Record duration", func(t *testing.T) {
		collector := NewPrometheusCollector()

		collector.RecordDuration("test_duration", 100*time.Millisecond, map[string]string{
			"operation": "test",
		})

		// Should not panic
		assert.NotNil(t, collector)
	})

	t.Run("Set gauge", func(t *testing.T) {
		collector := NewPrometheusCollector()

		collector.SetGauge("test_gauge", 42.0, map[string]string{
			"type": "test",
		})

		collector.IncGauge("test_gauge_inc", nil)
		collector.DecGauge("test_gauge_dec", nil)
		collector.AddGauge("test_gauge_add", 10.0, map[string]string{
			"type": "test",
		})

		// Should not panic
		assert.NotNil(t, collector)
	})
}

func TestStandaloneMetricsServer(t *testing.T) {
	t.Run("Create server", func(t *testing.T) {
		collector := NewPrometheusCollector()
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

		server := NewStandaloneMetricsServer(0, collector) // Port 0 for testing
		assert.NotNil(t, server)

		// Should be able to stop without starting
		err := server.Stop()
		assert.NoError(t, err)
	})
}

func TestEmbeddedMetricsEndpoints(t *testing.T) {
	t.Run("Create embedded endpoints", func(t *testing.T) {
		collector := NewPrometheusCollector()
		endpoints := NewEmbeddedMetricsEndpoints(collector)

		assert.NotNil(t, endpoints)
		assert.NotNil(t, endpoints.MetricsHandler())
		assert.NotNil(t, endpoints.HealthHandler())
	})
}

func TestInterfaceImplementation(t *testing.T) {
	t.Run("Implements MetricsCollector interface", func(t *testing.T) {
		var collector sharedMetrics.MetricsCollector = NewPrometheusCollector()

		// Test interface methods with unique metric names to avoid conflicts
		collector.IncrementCounter("interface_test_counter", map[string]string{"key": "value"})
		collector.IncrementCounterWithValue("interface_test_counter_with_value", 1.0, map[string]string{"key": "value"})
		collector.RecordDuration("interface_test_duration", time.Millisecond, map[string]string{"key": "value"})
		collector.SetGauge("interface_test_gauge", 1.0, map[string]string{"key": "value"})
		collector.IncGauge("interface_test_gauge_inc", map[string]string{"key": "value"})
		collector.DecGauge("interface_test_gauge_dec", map[string]string{"key": "value"})
		collector.AddGauge("interface_test_gauge_add", 1.0, map[string]string{"key": "value"})

		assert.NotNil(t, collector)
	})
}

func TestRealWorldUsage(t *testing.T) {
	t.Run("API service metrics", func(t *testing.T) {
		collector := NewPrometheusCollector()

		// Simulate API requests
		collector.IncrementCounter("http_requests_total", map[string]string{
			"method":      "GET",
			"endpoint":    "/api/orders",
			"status_code": "200",
		})

		collector.RecordDuration("http_request_duration_seconds", 150*time.Millisecond, map[string]string{
			"method":   "GET",
			"endpoint": "/api/orders",
		})

		collector.IncGauge("active_requests", nil)
		collector.DecGauge("active_requests", nil)

		assert.NotNil(t, collector)
	})

	t.Run("Order processing metrics", func(t *testing.T) {
		collector := NewPrometheusCollector()

		// Simulate order processing
		collector.IncrementCounter("orders_total", map[string]string{
			"status": "success",
		})

		collector.RecordDuration("order_processing_duration_seconds", 250*time.Millisecond, map[string]string{
			"operation": "place_order",
		})

		collector.SetGauge("active_orders", 5, nil)

		assert.NotNil(t, collector)
	})
}

func TestGinIntegration(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("Register with Gin router", func(t *testing.T) {
		collector := NewPrometheusCollector()
		endpoints := NewEmbeddedMetricsEndpoints(collector)

		router := gin.New()
		endpoints.RegisterWithGin(router, "/metrics", "/health")

		// Test metrics endpoint
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metrics", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")

		// Test health endpoint
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})

	t.Run("Register with Gin group", func(t *testing.T) {
		collector := NewPrometheusCollector()
		endpoints := NewEmbeddedMetricsEndpoints(collector)

		router := gin.New()
		adminGroup := router.Group("/admin")
		endpoints.RegisterWithGinGroup(adminGroup, "/metrics", "/health")

		// Test metrics endpoint
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/metrics", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		// Test health endpoint
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/admin/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})
}

func TestDynamicLabelMerging(t *testing.T) {
	t.Run("Same metric with different labels should log error", func(t *testing.T) {
		collector := NewPrometheusCollector()

		// First call with 3 labels
		collector.IncrementCounter("api_requests_total", map[string]string{
			"method":   "GET",
			"endpoint": "/orders",
			"status":   "200",
		})

		// Second call with different labels should log error but not panic
		collector.IncrementCounter("api_requests_total", map[string]string{
			"method":    "POST",
			"endpoint":  "/users",
			"status":    "201",
			"user_type": "premium", // This extra label should cause error logging
		})

		// No panic should occur, test should pass
		assert.NotNil(t, collector)
	})
}
