package router

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/http/handler"
	"github.com/gin-gonic/gin"
)

// SetupOrderRoutes sets up order-related routes
func SetupOrderRoutes(router *gin.RouterGroup, orderHandler *handler.OrderHandler) {
	orders := router.Group("/orders")
	{
		orders.POST("", orderHandler.PlaceOrder)
	}
}
