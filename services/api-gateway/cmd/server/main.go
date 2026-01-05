package main

import (
	"fmt"
	"log"

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

	authConn, err := grpcInfra.NewAuthConn(cfg.GRPC.AuthService)
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer authConn.Close()

	productConn, err := grpcInfra.NewProductConn(cfg.GRPC.ProductService)
	if err != nil {
		log.Fatalf("Failed to connect to product service: %v", err)
	}
	defer productConn.Close()

	authClient := clients.NewAuthClient(authConn)
	productClient := clients.NewProductClient(productConn)

	jwtMiddleware := middleware.NewJWTMiddleware(authpb.NewAuthServiceClient(authConn))

	r := gin.New()
	authHandler := handler.NewAuthHandler(authClient)
	productHandler := handler.NewProductHandler(productClient)
	router.Register(r, authHandler, jwtMiddleware, productHandler)

	r.Run(fmt.Sprintf(":%s", cfg.HTTP.Port))
}
