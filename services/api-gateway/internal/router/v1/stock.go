package v1

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/handler"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterStock(
	r *gin.RouterGroup,
	stockHandler *handler.StockHandler,
	jwtMiddleware gin.HandlerFunc,
	productOwnershipMiddleware *middleware.ProductOwnershipMiddleware,
) {
	// Stock routes
	stock := r.Group("/stock")
	{
		// Public routes
		stock.GET("/products/:product_id", stockHandler.GetStock)

		// Protected routes (require authentication)
		authenticated := stock.Group("")
		authenticated.Use(jwtMiddleware)
		{
			authenticated.POST("/reserve", stockHandler.ReserveStock)
			authenticated.DELETE("/reservations/:reservation_id", stockHandler.ReleaseReservation)
			authenticated.GET("/reservations/:reservation_id", stockHandler.GetReservation)
		}

		seller := stock.Group("")
		seller.Use(jwtMiddleware, productOwnershipMiddleware.VerifySeller())
		{
			seller.POST("/products/:product_id/stock", stockHandler.SetStock)
		}
	}
}
