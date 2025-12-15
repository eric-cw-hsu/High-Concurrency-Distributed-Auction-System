package logger

import (
	"fmt"
	"log"

	// ensure environment variables from .env are loaded before this package init
	_ "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/envloader"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared"
)

// Global default logger instance
var globalLogger *Logger

func init() {
	service_name := shared.GetEnv("SERVICE_NAME", "local")
	globalLogger = NewLogger(service_name, NewConsoleSender())
}

func AddSender(sender LogSender) {
	globalLogger.AddSender(sender)
}

// Global functions that use the default logger
func Debug(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Fatal(msg, fields...)
	} else {
		// Fallback behavior if no default logger is set
		log.Fatal(msg)
	}
}

// string format of the log message
func Infof(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(fmt.Sprintf(format, args...))
	}
}

func Warnf(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(fmt.Sprintf(format, args...))
	}
}

func Errorf(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(fmt.Sprintf(format, args...))
	}
}

func Fatalf(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Fatal(fmt.Sprintf(format, args...))
	} else {
		// Fallback behavior if no default logger is set
		log.Fatalf(format, args...)
	}
}

func LogWithPayload(payload LogPayload) {
	if globalLogger != nil {
		globalLogger.dispatchLog(payload.Level, payload.Service, payload.Message, payload.Fields)
	}
}
