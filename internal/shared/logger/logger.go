package logger

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"github.com/segmentio/kafka-go"
)

// LogLevel represents the severity of the log
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// LogMessage represents a structured log message
type LogMessage struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Service     string                 `json:"service"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Operation   string                 `json:"operation,omitempty"`
	Duration    *time.Duration         `json:"duration,omitempty"`
	ErrorDetail string                 `json:"error_detail,omitempty"`
}

// LogSender is a function type for sending log messages
type LogSender func(LogMessage) error

// Logger handles logging with configurable output
type Logger struct {
	serviceName   string
	isDevelopment bool
	sender        LogSender
}

// Global default logger instance
var defaultLogger *Logger

// NewKafkaSender creates a Kafka log sender
func NewKafkaSender(brokers []string, topic string) LogSender {
	writer := kafkaInfra.NewWriter(brokers, topic)
	return func(msg LogMessage) error {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		return writer.WriteMessages(context.Background(), kafka.Message{
			Value: data,
			Time:  time.Now(),
		})
	}
}

func NewConsoleSender() LogSender {
	return func(msg LogMessage) error {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		log.Println(string(data))
		return nil
	}
}

// NewLogger creates a new Logger instance
func NewLogger(serviceName string, sender LogSender) *Logger {
	isDev := os.Getenv("ENVIRONMENT") == "development"
	return &Logger{
		serviceName:   serviceName,
		isDevelopment: isDev,
		sender:        sender,
	}
}

// SetDefault sets the global default logger
func SetDefault(logger *Logger) {
	defaultLogger = logger
}

// log handles the actual logging logic
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if l == nil {
		return
	}

	logMsg := LogMessage{
		Timestamp: time.Now(),
		Level:     level,
		Service:   l.serviceName,
		Message:   message,
		Fields:    fields,
	}

	// Extract common fields if they exist
	if fields != nil {
		if traceID, ok := fields["trace_id"].(string); ok {
			logMsg.TraceID = traceID
		}
		if requestID, ok := fields["request_id"].(string); ok {
			logMsg.RequestID = requestID
		}
		if userID, ok := fields["user_id"].(string); ok {
			logMsg.UserID = userID
		}
		if operation, ok := fields["operation"].(string); ok {
			logMsg.Operation = operation
		}
		if duration, ok := fields["duration"].(time.Duration); ok {
			logMsg.Duration = &duration
		}
		if errorDetail, ok := fields["error_detail"].(string); ok {
			logMsg.ErrorDetail = errorDetail
		}
	}

	// Always send to configured sender (Kafka)
	if l.sender != nil {
		if err := l.sender(logMsg); err != nil {
			// Fallback to console if sender fails
			log.Printf("[ERROR] %s: Failed to send log: %v", l.serviceName, err)
		}
	}

	// Output to console in development mode
	if l.isDevelopment {
		l.logToConsole(level, message, fields)
	}

	// Handle Fatal level
	if level == FATAL {
		os.Exit(1)
	}
}

// logToConsole outputs log to console in a readable format
func (l *Logger) logToConsole(level LogLevel, message string, fields map[string]interface{}) {
	timestamp := time.Now().Format("15:04:05")

	if fields != nil && len(fields) > 0 {
		log.Printf("[%s] %s: %s (%s) %+v", level, l.serviceName, message, timestamp, fields)
	} else {
		log.Printf("[%s] %s: %s (%s)", level, l.serviceName, message, timestamp)
	}
}

// Instance methods
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DEBUG, message, f)
}

func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(INFO, message, f)
}

func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WARN, message, f)
}

func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ERROR, message, f)
}

func (l *Logger) Fatal(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(FATAL, message, f)
}

func (l *Logger) Close() error {
	// If needed, we can add cleanup logic here
	return nil
}

// Global functions that use the default logger
func Debug(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(message, fields...)
	}
}

func Info(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(message, fields...)
	}
}

func Warn(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(message, fields...)
	}
}

func Error(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(message, fields...)
	}
}

func Fatal(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(message, fields...)
	} else {
		// Fallback behavior if no default logger is set
		log.Fatal(message)
	}
}
