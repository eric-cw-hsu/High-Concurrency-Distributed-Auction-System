package logger

import (
	"context"

	"go.uber.org/zap"
)

// Context keys
type contextKey string

const (
	traceIDKey contextKey = "trace_id"
	userIDKey  contextKey = "user_id"
)

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID gets trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID gets user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// fromContext returns a logger with context metadata
func fromContext(ctx context.Context) *zap.Logger {
	logger := zap.L() // Get global logger

	fields := make([]zap.Field, 0, 2)

	if traceID := GetTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if userID := GetUserID(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	if len(fields) == 0 {
		return logger
	}

	return logger.With(fields...)
}

// InfoContext logs info with context metadata
func InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	log := fromContext(ctx)
	log.WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

// WarnContext logs warning with context metadata
func WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	log := fromContext(ctx)
	log.WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

// ErrorContext logs error with context metadata
func ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	log := fromContext(ctx)
	log.WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

// DebugContext logs debug with context metadata
func DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	log := fromContext(ctx)
	log.WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}
