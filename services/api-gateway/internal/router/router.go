package router

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/handler"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/middleware"
	v1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/router/v1"
	"github.com/gin-gonic/gin"
)

func Register(
	r *gin.Engine,
	authHandler *handler.AuthHandler,
	jwtMiddleware gin.HandlerFunc,
	productHandler *handler.ProductHandler,
	stockHandler *handler.StockHandler,
	productOwnershipMiddleware *middleware.ProductOwnershipMiddleware,
	orderHandler *handler.OrderHandler,
) {
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	api := r.Group("/api")
	{
		v1Router := api.Group("v1")
		{
			v1.RegisterAuth(v1Router, authHandler, jwtMiddleware)
			v1.RegisterProduct(v1Router, productHandler, jwtMiddleware)
			v1.RegisterStock(v1Router, stockHandler, jwtMiddleware, productOwnershipMiddleware)
			v1.RegisterOrder(v1Router, orderHandler, jwtMiddleware)
		}

	}
}
