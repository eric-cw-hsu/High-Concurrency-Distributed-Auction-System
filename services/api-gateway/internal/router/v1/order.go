package v1

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterOrder(
	r *gin.RouterGroup,
	orderHandler *handler.OrderHandler,
	jwtMiddleware gin.HandlerFunc,
) {
	order := r.Group("/orders")

	// Apply JWT middleware - all order queries must be authenticated
	order.Use(jwtMiddleware)
	{
		// GET /v1/orders/:order_id - Get specific order details
		order.GET("/:order_id", orderHandler.GetOrder)

		// GET /v1/orders - List all orders for the authenticated user
		order.GET("", orderHandler.ListUserOrders)
	}
}
