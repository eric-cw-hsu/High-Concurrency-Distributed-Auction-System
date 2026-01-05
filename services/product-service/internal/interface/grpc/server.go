package grpc

import (
	"fmt"
	"net"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/config"
	productv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/product/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server wraps gRPC server
type Server struct {
	grpcServer *grpc.Server
	handler    *ProductHandler
	config     *config.ServerConfig
}

// NewServer creates a new gRPC server
func NewServer(
	cfg *config.ServerConfig,
	productService *service.ProductService,
) *Server {
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
		grpc.ChainUnaryInterceptor(
			UnaryServerInterceptor(), // Logging and tracing
		),
	)

	handler := NewProductHandler(productService)

	// Register service
	productv1.RegisterProductServiceServer(grpcServer, handler)

	// Register reflection
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
