package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/envloader"

	orderProcessorBootstrap "eric-cw-hsu.github.io/scalable-auction-system/internal/app/order_processor/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize all dependencies (config, logger, metrics, kafka, db, consumer)
	deps, err := orderProcessorBootstrap.InitDependencies(ctx)
	if err != nil {
		logger.Fatal("Failed to initialize order processor dependencies", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Start metrics server (async)
	deps.MetricsServer.StartAsync()

	// Start consumer (async)
	logger.Info("Starting order processor")
	go func() {
		if err := deps.Consumer.Start(ctx); err != nil {
			logger.Warn("Consumer stopped", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Order Processor service is running")

	<-sigChan
	logger.Info("Received shutdown signal, stopping consumer")
	cancel()

	// Cleanup
	deps.Shutdown()

	logger.Info("Order Processor service stopped gracefully")
}
