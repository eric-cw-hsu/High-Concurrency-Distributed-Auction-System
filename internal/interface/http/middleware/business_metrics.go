package middleware

import (
	"sync/atomic"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
	"github.com/gin-gonic/gin"
)

// BusinessMetricsMiddleware records business-specific metrics
// This middleware should be applied to specific business endpoints
func BusinessMetricsMiddleware(collector metrics.MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()
		method := c.Request.Method

		// Track specific business operations
		switch {
		case route == "/api/v1/orders" && method == "POST":
			collector.IncrementCounter("business_order_attempts_total", map[string]string{
				"service": "api",
			})

		case route == "/api/v1/users" && method == "POST":
			collector.IncrementCounter("business_user_registration_attempts_total", map[string]string{
				"service": "api",
			})

		case route == "/api/v1/products" && method == "POST":
			collector.IncrementCounter("business_product_creation_attempts_total", map[string]string{
				"service": "api",
			})

		case route == "/api/v1/wallet" && method == "POST":
			collector.IncrementCounter("business_wallet_operations_total", map[string]string{
				"service":   "api",
				"operation": "funding",
			})
		}

		c.Next()

		// Record success/failure based on status code
		statusCode := c.Writer.Status()
		success := statusCode >= 200 && statusCode < 400

		switch {
		case route == "/api/v1/orders" && method == "POST":
			status := "failure"
			if success {
				status = "success"
			}
			collector.IncrementCounter("business_order_completions_total", map[string]string{
				"service": "api",
				"status":  status,
			})

		case route == "/api/v1/users" && method == "POST":
			status := "failure"
			if success {
				status = "success"
			}
			collector.IncrementCounter("business_user_registrations_total", map[string]string{
				"service": "api",
				"status":  status,
			})
		}
	}
}

// DatabaseMetricsHelper provides helper functions for recording database metrics
type DatabaseMetricsHelper struct {
	collector metrics.MetricsCollector
}

func NewDatabaseMetricsHelper(collector metrics.MetricsCollector) *DatabaseMetricsHelper {
	return &DatabaseMetricsHelper{collector: collector}
}

func (d *DatabaseMetricsHelper) RecordQuery(operation string, table string, success bool) {
	status := "failure"
	if success {
		status = "success"
	}

	d.collector.IncrementCounter("database_queries_total", map[string]string{
		"service":   "api",
		"operation": operation, // SELECT, INSERT, UPDATE, DELETE
		"table":     table,
		"status":    status,
	})
}

func (d *DatabaseMetricsHelper) RecordConnectionPoolStats(active, idle, total int) {
	d.collector.SetGauge("database_connections_active", float64(active), map[string]string{
		"service": "api",
	})
	d.collector.SetGauge("database_connections_idle", float64(idle), map[string]string{
		"service": "api",
	})
	d.collector.SetGauge("database_connections_total", float64(total), map[string]string{
		"service": "api",
	})
}

// RedisMetricsHelper provides helper functions for recording Redis metrics
type RedisMetricsHelper struct {
	collector metrics.MetricsCollector
}

func NewRedisMetricsHelper(collector metrics.MetricsCollector) *RedisMetricsHelper {
	return &RedisMetricsHelper{collector: collector}
}

func (r *RedisMetricsHelper) RecordCacheOperation(operation string, hit bool) {
	hitStatus := "miss"
	if hit {
		hitStatus = "hit"
	}

	r.collector.IncrementCounter("cache_operations_total", map[string]string{
		"service":   "api",
		"operation": operation, // GET, SET, DEL
		"result":    hitStatus,
	})
}

func (r *RedisMetricsHelper) RecordCacheHitRate(rate float64) {
	r.collector.SetGauge("cache_hit_rate", rate, map[string]string{
		"service": "api",
	})
}

// KafkaMetricsHelper provides helper functions for recording Kafka metrics
type KafkaMetricsHelper struct {
	collector metrics.MetricsCollector
}

func NewKafkaMetricsHelper(collector metrics.MetricsCollector) *KafkaMetricsHelper {
	return &KafkaMetricsHelper{collector: collector}
}

func (k *KafkaMetricsHelper) RecordMessagePublished(topic string, success bool) {
	status := "failure"
	if success {
		status = "success"
	}

	k.collector.IncrementCounter("kafka_messages_published_total", map[string]string{
		"service": "api",
		"topic":   topic,
		"status":  status,
	})
}

func (k *KafkaMetricsHelper) RecordMessageProcessed(topic string, success bool) {
	status := "failure"
	if success {
		status = "success"
	}

	k.collector.IncrementCounter("kafka_messages_processed_total", map[string]string{
		"service": "api",
		"topic":   topic,
		"status":  status,
	})
}

// Global counters for more accurate tracking
var (
	activeRequests int64
)

func IncrementActiveRequests() {
	atomic.AddInt64(&activeRequests, 1)
}

func DecrementActiveRequests() {
	atomic.AddInt64(&activeRequests, -1)
}

func GetActiveRequests() int64 {
	return atomic.LoadInt64(&activeRequests)
}
