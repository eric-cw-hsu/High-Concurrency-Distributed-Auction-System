package main

import (
	_ "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/envloader"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/app/api"
)

func main() {
	// Start the API service
	if err := api.Run(); err != nil {
		logger.Fatal("Failed to run the application", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
