package testutil

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// PerformanceThresholds defines performance expectations for tests
type PerformanceThresholds struct {
	MinSuccessRate    float64       `json:"min_success_rate"`    // Minimum acceptable success rate (0.0-1.0)
	MaxAverageLatency time.Duration `json:"max_average_latency"` // Maximum acceptable average latency
	MinThroughput     float64       `json:"min_throughput"`      // Minimum throughput (orders/sec)
	MaxTestDuration   time.Duration `json:"max_test_duration"`   // Maximum total test duration
}

// TestConfiguration defines test execution parameters
type TestConfiguration struct {
	Concurrency     int     `json:"concurrency"`       // Number of concurrent goroutines
	OrdersPerThread int     `json:"orders_per_thread"` // Orders per goroutine
	InitialStock    int64   `json:"initial_stock"`     // Initial stock amount
	OrderPrice      float64 `json:"order_price"`       // Price per order
	WalletBalance   float64 `json:"wallet_balance"`    // Initial wallet balance per user
}

// GetDefaultPerformanceThresholds returns conservative default thresholds
func GetDefaultPerformanceThresholds() PerformanceThresholds {
	return PerformanceThresholds{
		MinSuccessRate:    getEnvAsFloat("TEST_MIN_SUCCESS_RATE", 0.95),                  // 95% success rate
		MaxAverageLatency: getEnvAsDuration("TEST_MAX_AVG_LATENCY", 50*time.Millisecond), // 50ms
		MinThroughput:     getEnvAsFloat("TEST_MIN_THROUGHPUT", 200.0),                   // 200 orders/sec
		MaxTestDuration:   getEnvAsDuration("TEST_MAX_DURATION", 5*time.Second),          // 5 seconds
	}
}

// GetHighPerformanceThresholds returns aggressive thresholds for high-performance tests
func GetHighPerformanceThresholds() PerformanceThresholds {
	return PerformanceThresholds{
		MinSuccessRate:    getEnvAsFloat("TEST_HIGH_MIN_SUCCESS_RATE", 0.99),                  // 99% success rate
		MaxAverageLatency: getEnvAsDuration("TEST_HIGH_MAX_AVG_LATENCY", 20*time.Millisecond), // 20ms
		MinThroughput:     getEnvAsFloat("TEST_HIGH_MIN_THROUGHPUT", 500.0),                   // 500 orders/sec
		MaxTestDuration:   getEnvAsDuration("TEST_HIGH_MAX_DURATION", 2*time.Second),          // 2 seconds
	}
}

// GetDefaultTestConfiguration returns default test configuration
func GetDefaultTestConfiguration() TestConfiguration {
	return TestConfiguration{
		Concurrency:     getEnvAsInt("TEST_CONCURRENCY", 20),
		OrdersPerThread: getEnvAsInt("TEST_ORDERS_PER_THREAD", 2),
		InitialStock:    int64(getEnvAsInt("TEST_INITIAL_STOCK", 1000)),
		OrderPrice:      getEnvAsFloat("TEST_ORDER_PRICE", 100.0),
		WalletBalance:   getEnvAsFloat("TEST_WALLET_BALANCE", 10000.0),
	}
}

// GetStressTestConfiguration returns configuration for stress testing
func GetStressTestConfiguration() TestConfiguration {
	return TestConfiguration{
		Concurrency:     getEnvAsInt("TEST_STRESS_CONCURRENCY", 100),
		OrdersPerThread: getEnvAsInt("TEST_STRESS_ORDERS_PER_THREAD", 5),
		InitialStock:    int64(getEnvAsInt("TEST_STRESS_INITIAL_STOCK", 500)),
		OrderPrice:      getEnvAsFloat("TEST_STRESS_ORDER_PRICE", 100.0),
		WalletBalance:   getEnvAsFloat("TEST_STRESS_WALLET_BALANCE", 50000.0),
	}
}

// PerformanceResult holds the results of a performance test
type PerformanceResult struct {
	Duration         time.Duration `json:"duration"`
	TotalOrders      int64         `json:"total_orders"`
	SuccessfulOrders int64         `json:"successful_orders"`
	FailedOrders     int64         `json:"failed_orders"`
	SuccessRate      float64       `json:"success_rate"`
	AverageLatency   time.Duration `json:"average_latency"`
	Throughput       float64       `json:"throughput"`
	InitialStock     int64         `json:"initial_stock"`
	FinalStock       int64         `json:"final_stock"`
	StockConsumed    int64         `json:"stock_consumed"`
}

// MeetsThresholds checks if the performance result meets the given thresholds
func (r *PerformanceResult) MeetsThresholds(thresholds PerformanceThresholds) bool {
	return r.SuccessRate >= thresholds.MinSuccessRate &&
		r.AverageLatency <= thresholds.MaxAverageLatency &&
		r.Throughput >= thresholds.MinThroughput &&
		r.Duration <= thresholds.MaxTestDuration
}

// GetFailedThresholds returns a list of failed threshold descriptions
func (r *PerformanceResult) GetFailedThresholds(thresholds PerformanceThresholds) []string {
	var failures []string

	if r.SuccessRate < thresholds.MinSuccessRate {
		failures = append(failures, fmt.Sprintf("Success rate %.2f%% < %.2f%%",
			r.SuccessRate*100, thresholds.MinSuccessRate*100))
	}

	if r.AverageLatency > thresholds.MaxAverageLatency {
		failures = append(failures, fmt.Sprintf("Average latency %v > %v",
			r.AverageLatency, thresholds.MaxAverageLatency))
	}

	if r.Throughput < thresholds.MinThroughput {
		failures = append(failures, fmt.Sprintf("Throughput %.2f orders/sec < %.2f orders/sec",
			r.Throughput, thresholds.MinThroughput))
	}

	if r.Duration > thresholds.MaxTestDuration {
		failures = append(failures, fmt.Sprintf("Test duration %v > %v",
			r.Duration, thresholds.MaxTestDuration))
	}

	return failures
}

// Helper functions to get environment variables with defaults
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
