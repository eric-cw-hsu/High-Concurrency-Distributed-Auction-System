package router

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/http/handler"
	"github.com/gin-gonic/gin"
)

// SetupProductRoutes sets up product-related routes
func SetupProductRoutes(router *gin.RouterGroup, productHandler *handler.ProductHandler) {
	products := router.Group("/products")
	{
		products.POST("", productHandler.CreateProduct)
	}
}
