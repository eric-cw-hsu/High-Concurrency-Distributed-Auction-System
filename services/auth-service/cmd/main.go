package main

import (
	"log"
	"net"
	"os"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/infrastructure/crypto"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/infrastructure/identity"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/infrastructure/jwt"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/infrastructure/persistence/postgres"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/infrastructure/persistence/redis"
	grpcHandler "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/interface/grpc"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/config"
	pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/auth/v1"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	// 1. load config
	if os.Getenv("ENV") != "production" {
		_ = godotenv.Load()
	}

	cfg := config.Load()

	// 2. init infras
	db := postgres.MustConnect(cfg.Database.DSN)
	redisClient := redis.MustConnect(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

	// 3. init repos
	userRepo := postgres.NewUserRepository(db)
	refreshRepo := redis.NewRedisRefreshRepo(redisClient)

	// 4. init domain services
	jwtProvider := jwt.NewJWTProvider(cfg.JWT)
	bcryptVerifier := crypto.NewBcryptVerifier(cfg.Bcrypt.Cost)
	idGenerator := identity.NewUUIDGenerator()

	// 5. init application services
	authService := service.NewAuthService(userRepo, bcryptVerifier, jwtProvider, jwtProvider, refreshRepo, idGenerator)

	// 6. init gRPC server
	grpcServer := grpc.NewServer()
	authHandler := grpcHandler.NewAuthHandler(authService)
	pb.RegisterAuthServiceServer(grpcServer, authHandler)

	// 7. Serve
	lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		panic(err)
	}

	log.Println("auth-service listening on", cfg.GRPC.Port)
	grpcServer.Serve(lis)
}
