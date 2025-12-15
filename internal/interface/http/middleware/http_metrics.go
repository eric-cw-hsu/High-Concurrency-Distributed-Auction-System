package middleware

import (
	"strconv"
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
	"github.com/gin-gonic/gin"
)

// HTTPMetricsMiddleware creates a middleware that records comprehensive HTTP metrics
// Records: request count, response time, request/response sizes, error rates, active connections
func HTTPMetricsMiddleware(collector metrics.MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Track active requests
		IncrementActiveRequests()
		defer DecrementActiveRequests()

		// Record active connections
		collector.IncGauge("http_active_connections", map[string]string{
			"service": "api",
		})
		defer func() {
			collector.DecGauge("http_active_connections", map[string]string{
				"service": "api",
			})
		}()

		start := time.Now()

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		method := c.Request.Method
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}
		statusCode := c.Writer.Status()
		statusClass := strconv.Itoa(statusCode/100) + "xx"

		// Base labels for all metrics
		baseLabels := map[string]string{
			"method":  method,
			"route":   route,
			"service": "api",
		}

		// Labels with status information
		statusLabels := map[string]string{
			"method":       method,
			"route":        route,
			"service":      "api",
			"status_code":  strconv.Itoa(statusCode),
			"status_class": statusClass,
		}

		// 1. Record total request count
		collector.IncrementCounter("http_requests_total", statusLabels)

		// 2. Record request duration histogram
		collector.RecordDuration("http_request_duration_seconds", duration, map[string]string{
			"method":       method,
			"route":        route,
			"service":      "api",
			"status_class": statusClass,
		})

		// 3. Record request size (if available)
		if c.Request.ContentLength > 0 {
			collector.IncrementCounterWithValue("http_request_size_bytes_total", float64(c.Request.ContentLength), baseLabels)
		}

		// 4. Record response size
		responseSize := float64(c.Writer.Size())
		if responseSize > 0 {
			collector.IncrementCounterWithValue("http_response_size_bytes_total", responseSize, map[string]string{
				"method":       method,
				"route":        route,
				"service":      "api",
				"status_class": statusClass,
			})
		}

		// 5. Track slow requests (> 1 second)
		if duration > time.Second {
			collector.IncrementCounter("http_slow_requests_total", statusLabels)
			logger.Warn("Slow HTTP request detected", map[string]interface{}{
				"method":      method,
				"route":       route,
				"duration_ms": duration.Milliseconds(),
				"status_code": statusCode,
			})
		}

		// 6. Track error rates
		if statusCode >= 400 {
			collector.IncrementCounter("http_errors_total", map[string]string{
				"method":      method,
				"route":       route,
				"service":     "api",
				"status_code": strconv.Itoa(statusCode),
				"error_type":  getErrorType(statusCode),
			})
		}

		// 7. Record concurrent request gauge (current active requests)
		collector.SetGauge("http_requests_in_flight", float64(GetActiveRequests()), map[string]string{
			"service": "api",
		})
	}
}

// getErrorType categorizes HTTP errors
func getErrorType(statusCode int) string {
	switch {
	case statusCode >= 400 && statusCode < 500:
		return "client_error"
	case statusCode >= 500:
		return "server_error"
	default:
		return "unknown"
	}
}
