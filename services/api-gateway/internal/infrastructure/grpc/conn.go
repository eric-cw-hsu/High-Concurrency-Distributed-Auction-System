package grpc

import (
	"fmt"
	"os"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func MustConnect(cfg config.GRPCClientConfig) *grpc.ClientConn {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Duration(cfg.KeepaliveTime) * time.Second,
			Timeout:             time.Duration(cfg.KeepaliveTimeout) * time.Second,
			PermitWithoutStream: cfg.PermitWithoutStream,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallSendMsgSize(10*1024*1024),
		),
	)
	if err != nil {
		fmt.Printf("failed to connect to GRPC SERVER [%s]: %w", addr, err)
		os.Exit(1)
	}

	fmt.Printf("[GRPC] Connected to %s\n", addr)

	return conn
}
