package grpc

import (
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewAuthConn(cfg config.GRPCAuthConfig) *grpc.ClientConn {
	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("failed to connect to Auth Service at %s: %v", address, err))
	}
	return conn
}
