package v1

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProduct(
	r *gin.RouterGroup,
	productHandler *handler.ProductHandler,
	jwtMiddleware gin.HandlerFunc,
) {
	// Product routes
	products := r.Group("/products")
	{
		// Public routes (no auth required)
		products.GET("/:id", productHandler.GetProduct)
		products.GET("", productHandler.GetActiveProducts)

		// Protected routes (require authentication)
		authenticated := products.Group("")
		authenticated.Use(jwtMiddleware)
		{
			authenticated.POST("", productHandler.CreateProduct)
			authenticated.PUT("/:id", productHandler.UpdateProductInfo)
			authenticated.PUT("/:id/pricing", productHandler.UpdateProductPricing)
			authenticated.POST("/:id/publish", productHandler.PublishProduct)
			authenticated.POST("/:id/deactivate", productHandler.DeactivateProduct)
			authenticated.DELETE("/:id", productHandler.DeleteProduct)
		}
	}

	// Seller routes
	sellers := r.Group("/sellers")
	{
		sellers.GET("/:id/products", productHandler.GetProductsBySeller)
	}
}
