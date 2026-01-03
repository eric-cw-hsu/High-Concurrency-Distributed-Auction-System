package router

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/handler"
	v1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/router/v1"
	"github.com/gin-gonic/gin"
)

func Register(
	r *gin.Engine,
	authHandler *handler.AuthHandler,
	jwtMiddleware gin.HandlerFunc,
) {
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	api := r.Group("/api")
	{
		v1.Register(api.Group("/v1"), authHandler, jwtMiddleware)
	}
}
