package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/envloader"

	inventoryProcessorBootstrap "eric-cw-hsu.github.io/scalable-auction-system/internal/app/inventory_processor/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize dependencies using bootstrap
	deps, err := inventoryProcessorBootstrap.InitDependencies(ctx)
	if err != nil {
		logger.Fatal("Failed to initialize inventory processor dependencies", map[string]interface{}{
			"error": err.Error(),
		})
	}

	deps.MetricsServer.StartAsync()

	// Start consumer
	go func() {
		logger.Info("Starting Inventory Processor", nil)
		if err := deps.Consumer.Start(ctx); err != nil {
			logger.Error("Inventory Processor failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal, stopping inventory processor")
	cancel()

	deps.Shutdown()

	logger.Info("Inventory Processor Service stopped gracefully")
}
