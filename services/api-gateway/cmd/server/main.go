package main

import (
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/clients"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/handler"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/middleware"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/router"
	authpb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/auth/v1"
	"github.com/gin-gonic/gin"

	grpcInfra "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/infrastructure/grpc"
)

func main() {
	cfg := config.Load()

	authConn := grpcInfra.NewAuthConn(cfg.GRPC.AuthService)
	authClient := clients.NewAuthClient(authConn)

	jwtMiddleware := middleware.NewJWTMiddleware(authpb.NewAuthServiceClient(authConn))

	r := gin.New()
	authHandler := handler.NewAuthHandler(authClient)
	router.Register(r, authHandler, jwtMiddleware)

	r.Run(fmt.Sprintf(":%s", cfg.HTTP.Port))
}
