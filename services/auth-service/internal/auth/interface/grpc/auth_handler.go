package grpc

import (
	"context"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/common/logger"
	pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/auth/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/status"
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

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	logger.InfoContext(ctx, "handling Register request",
		zap.String("email", req.GetEmail()),
	)

	userID, accessToken, refreshToken, err := h.authService.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if isSystemError(code) {
			logger.ErrorContext(ctx, "register failed",
				zap.String("email", req.GetEmail()),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else if isBusinessError(code) {
			logger.WarnContext(ctx, "register failed",
				zap.String("email", req.GetEmail()),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else {
			logger.DebugContext(ctx, "register failed",
				zap.String("email", req.GetEmail()),
				zap.String("error", err.Error()),
			)
		}

		return nil, grpcErr
	}

	logger.InfoContext(ctx, "user registered successfully",
		zap.String("user_id", string(userID)),
		zap.String("email", req.GetEmail()),
	)

	return &pb.RegisterResponse{
		UserId:       string(userID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	logger.InfoContext(ctx, "handling Login request",
		zap.String("email", req.GetEmail()),
	)

	userID, accessToken, refreshToken, err := h.authService.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if isSystemError(code) {
			logger.ErrorContext(ctx, "login failed",
				zap.String("email", req.GetEmail()),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else if isBusinessError(code) {
			logger.WarnContext(ctx, "login failed",
				zap.String("email", req.GetEmail()),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else {
			logger.DebugContext(ctx, "login failed",
				zap.String("email", req.GetEmail()),
				zap.String("error", err.Error()),
			)
		}

		return nil, grpcErr
	}

	logger.InfoContext(ctx, "user logged in successfully",
		zap.String("user_id", string(userID)),
		zap.String("email", req.GetEmail()),
	)

	return &pb.LoginResponse{
		UserId:       string(userID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	logger.DebugContext(ctx, "handling Refresh request")

	accessToken, err := h.authService.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if isSystemError(code) {
			logger.ErrorContext(ctx, "refresh token failed",
				zap.String("error", err.Error()),
			)
		} else {
			logger.DebugContext(ctx, "refresh token failed",
				zap.String("error", err.Error()),
			)
		}

		return nil, grpcErr
	}

	logger.DebugContext(ctx, "token refreshed successfully")

	return &pb.RefreshResponse{
		AccessToken: accessToken,
	}, nil
}

func (h *AuthHandler) VerifyToken(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	logger.DebugContext(ctx, "handling VerifyToken request")

	access, err := h.authService.VerifyAccessToken(req.GetToken())
	if err != nil {
		logger.DebugContext(ctx, "token verification failed",
			zap.String("error", err.Error()),
		)
		return &pb.VerifyResponse{Valid: false}, nil
	}

	logger.DebugContext(ctx, "token verified successfully",
		zap.String("user_id", string(access.UserID)),
	)

	return &pb.VerifyResponse{
		Valid:  true,
		UserId: string(access.UserID),
		Role:   string(access.Role),
	}, nil
}
