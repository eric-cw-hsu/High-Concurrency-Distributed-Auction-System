package grpc

import (
	"fmt"
	"net"

	stockv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/stock/v1"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/recovery"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server wraps gRPC server
type Server struct {
	grpcServer *grpc.Server
	handler    *StockHandler
	config     *config.ServerConfig
}

// NewServer creates a new gRPC server
func NewServer(
	cfg *config.ServerConfig,
	stockService *service.StockService,
	recovery *recovery.RedisRecovery,
) *Server {
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
		grpc.ChainUnaryInterceptor(
			UnaryServerInterceptor(),
		),
	)

	handler := NewStockHandler(stockService, recovery)

	stockv1.RegisterStockServiceServer(grpcServer, handler)

	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		handler:    handler,
		config:     cfg,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	fmt.Printf("[gRPC Server] Listening on %s\n", addr)

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	fmt.Println("[gRPC Server] Shutting down...")
	s.grpcServer.GracefulStop()
}
