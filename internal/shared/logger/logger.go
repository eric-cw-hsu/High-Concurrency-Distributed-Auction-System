package logger

import (
	"os"
	"time"
)

// Logger handles logging with configurable output
type Logger struct {
	serviceName string
	senders     []LogSender
}

// NewLogger creates a new Logger instance
func NewLogger(serviceName string, senders ...LogSender) *Logger {
	return &Logger{
		serviceName: serviceName,
		senders:     senders,
	}
}

func (l *Logger) AddSender(sender LogSender) {
	if l != nil && sender != nil {
		l.senders = append(l.senders, sender)
	}
}

// log handles the actual logging logic
func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}) {
	l.dispatchLog(level, l.serviceName, msg, fields)

	// Handle Fatal level
	if level == FATAL {
		os.Exit(1)
	}
}

func (l *Logger) dispatchLog(level LogLevel, service, msg string, fields map[string]interface{}) {
	if l == nil {
		return
	}

	// enrich fields with caller info and stack if needed
	fields = addCallerInfo(fields, level)

	payload := LogPayload{
		Timestamp: time.Now(),
		Level:     level,
		Service:   service,
		Message:   msg,
		Fields:    fields,
	}

	// Always send to configured sender (Kafka)
	for _, sender := range l.senders {
		if err := sender(payload); err != nil {
			continue
		}
	}
}

// Instance methods
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DEBUG, msg, f)
}

func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(INFO, msg, f)
}

func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WARN, msg, f)
}

func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ERROR, msg, f)
}

func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(FATAL, msg, f)
}

func (l *Logger) Close() error {
	// If needed, we can add cleanup logic here
	return nil
}
