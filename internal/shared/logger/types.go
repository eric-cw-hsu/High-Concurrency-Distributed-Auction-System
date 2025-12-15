package logger

import "time"

// LogLevel represents the severity of the log
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// LogPayload is the minimal payload provided to a LogSender.
// This keeps the logger package decoupled from the transport/event schema
type LogPayload struct {
	Timestamp time.Time
	Level     LogLevel
	Service   string
	Message   string
	Fields    map[string]interface{}
}

// LogSender is a function type for sending logs. Senders are responsible
// for converting LogPayload into the transport/event format (e.g. message.LogMessage)
type LogSender func(LogPayload) error
