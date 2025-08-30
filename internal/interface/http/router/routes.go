package router

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/http/handler"
	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up all application routes
func SetupRoutes(engine *gin.Engine,
	userHandler *handler.UserHandler,
	walletHandler *handler.WalletHandler,
	productHandler *handler.ProductHandler,
	stockHandler *handler.StockHandler,
	orderHandler *handler.OrderHandler) {

	// API versioning
	api := engine.Group("/api")
	v1 := api.Group("/v1")

	// Setup module routes
	SetupUserRoutes(v1, userHandler)
	SetupWalletRoutes(v1, walletHandler)
	SetupProductRoutes(v1, productHandler)
	SetupStockRoutes(v1, stockHandler)
	SetupOrderRoutes(v1, orderHandler)

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "scalable-auction-system",
		})
	})
}
