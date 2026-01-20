package grpc

import (
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// MustConnProductClient initializes a gRPC connection to the Product Service.
// If the connection cannot be established within the timeout, it logs a Fatal error and terminates.
func MustConnProductClient(cfg config.GRPCConfig) *grpc.ClientConn {
	addr := cfg.ProductClient.Addr
	timeout := time.Duration(cfg.ProductClient.Timeout) * time.Second

	// 1. Initialize the client shell (non-blocking)
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                timeout,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		zap.L().Fatal("failed to initialize gRPC client shell",
			zap.String("address", addr),
			zap.Error(err),
		)
	}

	return conn

	// // 2. Readiness Check Loop
	// ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// defer cancel()

	// for {
	// 	state := conn.GetState()
	// 	if state.String() == "READY" {
	// 		zap.L().Info("successfully connected to product service", zap.String("address", addr))
	// 		return conn
	// 	}

	// 	if !conn.WaitForStateChange(ctx, state) {
	// 		conn.Close()
	// 		zap.L().Fatal("could not establish connection to product service within timeout",
	// 			zap.String("address", addr),
	// 			zap.Duration("timeout", timeout),
	// 		)
	// 	}
	// }
}
