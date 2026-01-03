package clients

import (
	"context"

	pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/auth/v1"
	"google.golang.org/grpc"
)

type AuthClient struct {
	cli pb.AuthServiceClient
}

func NewAuthClient(conn *grpc.ClientConn) *AuthClient {
	return &AuthClient{cli: pb.NewAuthServiceClient(conn)}
}

func (c *AuthClient) Register(ctx context.Context, email, password string) (string, error) {
	resp, err := c.cli.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.UserId, nil
}

func (c *AuthClient) Login(ctx context.Context, email, password string) (string, string, string, error) {
	resp, err := c.cli.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", "", "", err
	}

	return resp.UserId, resp.AccessToken, resp.RefreshToken, nil
}

func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	resp, err := c.cli.Refresh(ctx, &pb.RefreshRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return "", err
	}
	return resp.AccessToken, nil
}
