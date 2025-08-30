package router

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/http/handler"
	"github.com/gin-gonic/gin"
)

// SetupStockRoutes sets up stock-related routes
func SetupStockRoutes(router *gin.RouterGroup, stockHandler *handler.StockHandler) {
	stocks := router.Group("/stocks")
	{
		stocks.POST("", stockHandler.PutOnMarket)
	}
}
