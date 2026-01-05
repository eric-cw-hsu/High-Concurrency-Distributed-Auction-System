package grpc

import (
	"context"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/common/logger"
	"github.com/samborkent/uuidv7"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns a gRPC unary interceptor for tracing and metrics
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Extract or generate trace ID
		traceID := extractOrGenerateTraceID(ctx)
		ctx = logger.WithTraceID(ctx, traceID)

		// Log request (Debug level - only in development)
		logger.DebugContext(ctx, "grpc request started",
			zap.String("method", info.FullMethod),
		)

		// Call handler
		resp, err := handler(ctx, req)

		// Calculate latency
		latency := time.Since(start)

		// Log completion (Debug level)
		if err != nil {
			logger.DebugContext(ctx, "grpc request failed",
				zap.String("method", info.FullMethod),
				zap.Duration("latency", latency),
				zap.Error(err),
			)
		} else {
			logger.DebugContext(ctx, "grpc request completed",
				zap.String("method", info.FullMethod),
				zap.Duration("latency", latency),
			)
		}

		return resp, err
	}
}

// extractOrGenerateTraceID extracts trace ID from metadata or generates a new one
func extractOrGenerateTraceID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuidv7.New().String()
	}

	traceIDs := md.Get("x-trace-id")
	if len(traceIDs) > 0 && traceIDs[0] != "" {
		return traceIDs[0]
	}

	return uuidv7.New().String()
}
