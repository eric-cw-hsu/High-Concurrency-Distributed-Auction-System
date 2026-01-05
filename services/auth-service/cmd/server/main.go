package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/infrastructure/crypto"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/infrastructure/identity"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/infrastructure/jwt"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/infrastructure/persistence/postgres"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/infrastructure/persistence/redis"
	grpcHandler "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/interface/grpc"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/config"
	pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/auth/v1"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// Load .env
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Warning: .env file not found\n")
		}
	}

	// Load config
	cfg := config.Load()

	// Initialize logger
	log := logger.Init(&cfg.Logger)
	defer log.Sync()

	log.Info("starting auth service",
		zap.String("environment", cfg.Logger.Environment),
	)

	// Initialize infrastructure
	log.Info("connecting to database")
	db := postgres.MustConnect(cfg.Database.DSN)

	log.Info("connecting to redis",
		zap.String("addr", cfg.Redis.Addr),
	)
	redisClient := redis.MustConnect(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	refreshRepo := redis.NewRedisRefreshRepo(redisClient)

	// Initialize domain services
	jwtProvider := jwt.NewJWTProvider(cfg.JWT)
	bcryptVerifier := crypto.NewBcryptVerifier(cfg.Bcrypt.Cost)
	idGenerator := identity.NewUUIDGenerator()

	// Initialize application services
	authService := service.NewAuthService(
		userRepo,
		bcryptVerifier,
		jwtProvider,
		jwtProvider,
		refreshRepo,
		idGenerator,
	)

	// Initialize gRPC server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcHandler.UnaryServerInterceptor(),
		),
	)
	authHandler := grpcHandler.NewAuthHandler(authService)
	pb.RegisterAuthServiceServer(grpcServer, authHandler)

	// Start listening
	lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		log.Fatal("failed to listen",
			zap.String("port", cfg.GRPC.Port),
			zap.Error(err),
		)
	}

	// Start gRPC server in goroutine
	go func() {
		log.Info("auth service listening",
			zap.String("port", cfg.GRPC.Port),
		)
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("grpc server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gracefully")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Graceful stop gRPC server
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or timeout
	select {
	case <-stopped:
		log.Info("grpc server stopped gracefully")
	case <-ctx.Done():
		log.Warn("shutdown timeout, forcing stop")
		grpcServer.Stop() // Force stop
	}

	log.Info("auth service stopped")
}
