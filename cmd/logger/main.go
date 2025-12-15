package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/envloader"

	loggerBootstrap "eric-cw-hsu.github.io/scalable-auction-system/internal/app/logger/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize dependencies using bootstrap
	deps, err := loggerBootstrap.InitDependencies(ctx)
	if err != nil {
		logger.Fatal("Failed to initialize logger dependencies", map[string]interface{}{
			"error": err.Error(),
		})
	}

	deps.MetricsServer.StartAsync()

	// Start the logger consumer
	go func() {
		logger.Info("Starting logger consumer", nil)
		if err := deps.LoggerConsumer.Start(ctx); err != nil {
			logger.Fatal("Logger consumer stopped with error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	logger.Info("Logger consumer started successfully", nil)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	logger.Info("Received interrupt signal, shutting down...", nil)
	cancel()

	// Cleanup on exit
	deps.Shutdown()
	logger.Info("Logger service stopped gracefully", nil)
}
