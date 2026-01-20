package grpc

import (
	"fmt"
	"net"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/config"
	pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	config     *config.GRPCConfig
}

func NewServer(cfg *config.GRPCConfig, handler *OrderHandler) *Server {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryServerInterceptor()),
	)
	pb.RegisterOrderServiceServer(s, handler)

	// Enable reflection for debugging tools (e.g., Evans, Postman)
	reflection.Register(s)

	return &Server{
		grpcServer: s,
		config:     cfg,
	}
}

// Start opens the TCP listener and begins serving requests
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Server.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.config.Server.Port, err)
	}
	return s.grpcServer.Serve(lis)
}

// GracefulStop stops the server and ensures active RPCs are completed
func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}
