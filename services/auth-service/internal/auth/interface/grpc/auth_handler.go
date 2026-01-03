package grpc

import (
	"context"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/application/service"
	pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/auth/v1"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (s *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userID, accessToken, refreshToken, err := s.authService.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{
		UserId:       string(userID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	userID, accessToken, refreshToken, err := s.authService.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		UserId:       string(userID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthHandler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	accessToken, err := s.authService.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &pb.RefreshResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *AuthHandler) VerifyToken(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	access, err := s.authService.VerifyAccessToken(req.GetToken())

	if err != nil {
		return &pb.VerifyResponse{Valid: false}, nil
	}

	return &pb.VerifyResponse{
		Valid:  true,
		UserId: string(access.UserID),
		Role:   string(access.Role),
	}, nil
}
