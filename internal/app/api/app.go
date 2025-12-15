package api

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/api/bootstrap"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

func Run() error {
	ctx := context.Background()

	// Initialize all dependencies
	deps, err := bootstrap.InitDependencies(ctx)
	if err != nil {
		logger.Error("Failed to initialize dependencies", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Load API configuration
	appConfig := config.LoadAPIConfig()

	// Setup router
	r := bootstrap.RouterSetup(deps)

	if err := r.Run(":" + appConfig.Port); err != nil {
		return err
	}

	logger.Info("API Server started successfully", map[string]interface{}{
		"port": appConfig.Port,
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal, stopping API server")

	// Cleanup
	deps.Shutdown()

	logger.Info("API Server stopped gracefully", map[string]interface{}{
		"port": appConfig.Port,
	})

	return nil
}
