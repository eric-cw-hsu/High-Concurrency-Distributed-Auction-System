package main

import (
	"fmt"
	"os"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/clients"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/handler"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/middleware"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/router"
	authpb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/auth/v1"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	grpcInfra "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/infrastructure/grpc"
)

func main() {
	if os.Getenv("ENVIRONMENT") != "production" {
		_ = godotenv.Load()
	}

	cfg := config.Load()

	authConn := grpcInfra.MustConnect(cfg.GRPC.AuthService)
	defer authConn.Close()

	productConn := grpcInfra.MustConnect(cfg.GRPC.ProductService)
	defer productConn.Close()

	stockConn := grpcInfra.MustConnect(cfg.GRPC.StockService)
	defer stockConn.Close()

	orderConn := grpcInfra.MustConnect(cfg.GRPC.OrderService)
	defer orderConn.Close()

	authClient := clients.NewAuthClient(authConn)
	productClient := clients.NewProductClient(productConn)
	stockClient := clients.NewStockClient(stockConn)
	orderClient := clients.NewOrderClient(orderConn)

	jwtMiddleware := middleware.NewJWTMiddleware(authpb.NewAuthServiceClient(authConn))
	productOwnershipMiddleware := middleware.NewProductOwnershipMiddleware(productClient)

	authHandler := handler.NewAuthHandler(authClient)
	productHandler := handler.NewProductHandler(productClient)
	stockHandler := handler.NewStockHandler(stockClient)
	orderHandler := handler.NewOrderHandler(orderClient)

	r := gin.New()
	router.Register(r, authHandler, jwtMiddleware, productHandler, stockHandler, productOwnershipMiddleware, orderHandler)

	r.Run(fmt.Sprintf(":%s", cfg.HTTP.Port))
}
