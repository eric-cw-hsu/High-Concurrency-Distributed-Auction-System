package test

import (
	"fmt"
	"testing"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

// TestGlobalLoggerPattern tests the global logger design pattern
func TestGlobalLoggerPattern(t *testing.T) {
	// Test sender that captures messages
	var capturedMessages []logger.LogMessage
	testSender := func(msg logger.LogMessage) error {
		capturedMessages = append(capturedMessages, msg)
		return nil
	}

	// Create logger with test sender
	testLogger := logger.NewLogger("test-service", testSender)

	// Set as default global logger
	logger.SetDefault(testLogger)

	// Test global functions
	logger.Info("Test info message", map[string]interface{}{
		"test_field": "test_value",
	})

	logger.Error("Test error message", map[string]interface{}{
		"error_code": 500,
	})

	// Verify messages were captured
	if len(capturedMessages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(capturedMessages))
	}

	// Check first message
	if capturedMessages[0].Level != logger.INFO {
		t.Errorf("Expected INFO level, got %s", capturedMessages[0].Level)
	}
	if capturedMessages[0].Service != "test-service" {
		t.Errorf("Expected service 'test-service', got %s", capturedMessages[0].Service)
	}
	if capturedMessages[0].Message != "Test info message" {
		t.Errorf("Expected 'Test info message', got %s", capturedMessages[0].Message)
	}

	// Check second message
	if capturedMessages[1].Level != logger.ERROR {
		t.Errorf("Expected ERROR level, got %s", capturedMessages[1].Level)
	}

	fmt.Printf("✅ Global logger pattern working correctly\n")
	fmt.Printf("📨 Captured %d log messages\n", len(capturedMessages))
	for i, msg := range capturedMessages {
		fmt.Printf("   %d. [%s] %s: %s\n", i+1, msg.Level, msg.Service, msg.Message)
	}
}

// TestKafkaSenderCreation tests the Kafka sender factory
func TestKafkaSenderCreation(t *testing.T) {
	brokers := []string{"localhost:9092"}
	topic := "test.logs"

	// This should not panic
	sender := logger.NewKafkaSender(brokers, topic)
	if sender == nil {
		t.Error("Kafka sender should not be nil")
	}

	fmt.Printf("✅ Kafka sender creation working correctly\n")
}

// TestLoggerWithoutGlobalSetup tests behavior when no global logger is set
func TestLoggerWithoutGlobalSetup(t *testing.T) {
	// Reset global logger
	logger.SetDefault(nil)

	// These should not panic even without a global logger set
	logger.Info("This should not crash")
	logger.Error("This should also not crash")

	fmt.Printf("✅ Global logger gracefully handles nil state\n")
}
