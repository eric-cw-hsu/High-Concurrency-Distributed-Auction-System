package service

import (
	"context"
	"errors"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/auth/model"
	repository "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/auth/repositroy"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/port"
	domainSvc "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/user"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
)

type AuthService struct {
	users         user.UserRepository
	verifier      domainSvc.PasswordVerifier
	tokenIssuer   port.TokenIssuer
	tokenVerifier port.TokenVerifier
	refreshRepo   repository.RefreshTokenRepository
	idGenerator   kernel.IDGenerator
}

func NewAuthService(users user.UserRepository, verifier domainSvc.PasswordVerifier, tokenIssuer port.TokenIssuer, tokenVerifier port.TokenVerifier, refreshRepo repository.RefreshTokenRepository, idGenerator kernel.IDGenerator) *AuthService {
	return &AuthService{
		users:         users,
		tokenIssuer:   tokenIssuer,
		tokenVerifier: tokenVerifier,
		refreshRepo:   refreshRepo,
		verifier:      verifier,
		idGenerator:   idGenerator,
	}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (kernel.UserID, string, string, error) {
	existing, _ := s.users.FindByEmail(ctx, email)
	if existing != nil {
		return "", "", "", errors.New("email already in use")
	}

	hash, err := s.verifier.Hash(password)
	if err != nil {
		return "", "", "", err
	}

	userID := s.idGenerator.NewUserID()
	user := user.NewUserFromRegister(userID, email, hash)

	if err := s.users.Save(ctx, user); err != nil {
		return "", "", "", err
	}

	access, err := s.tokenIssuer.IssueAccess(user.ID(), kernel.RoleUser)
	if err != nil {
		return "", "", "", err
	}

	tokenID := s.idGenerator.NewTokenID()
	refresh, expiredAt, err := s.tokenIssuer.IssueRefresh(tokenID, user.ID())
	if err != nil {
		return "", "", "", err
	}

	rt := model.RefreshToken{
		ID:       tokenID,
		UserID:   userID,
		ExpireAt: expiredAt,
		Revoked:  false,
	}

	if err := s.refreshRepo.Save(ctx, rt); err != nil {
		return "", "", "", err
	}

	return user.ID(), access, refresh, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (kernel.UserID, string, string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", "", "", err
	}

	if err := user.Login(password, s.verifier); err != nil {
		return "", "", "", err
	}

	access, err := s.tokenIssuer.IssueAccess(user.ID(), "user")
	if err != nil {
		return "", "", "", err
	}

	tokenID := s.idGenerator.NewTokenID()
	refresh, expiredAt, err := s.tokenIssuer.IssueRefresh(tokenID, user.ID())
	if err != nil {
		return "", "", "", err
	}

	rt := model.RefreshToken{
		ID:       tokenID,
		UserID:   user.ID(),
		ExpireAt: expiredAt,
		Revoked:  false,
	}

	if err := s.refreshRepo.Save(ctx, rt); err != nil {
		return "", "", "", err
	}

	return user.ID(), access, refresh, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	tokenID, err := s.tokenVerifier.VerifyRefresh(refreshToken)
	if err != nil {
		return "", err
	}

	var rt *model.RefreshToken
	rt, err = s.refreshRepo.Find(ctx, tokenID)
	if err != nil {
		return "", errors.New("refresh token revoked")
	}

	return s.tokenIssuer.IssueAccess(rt.UserID, kernel.RoleUser)
}

func (s *AuthService) VerifyAccessToken(accessToken string) (*model.Access, error) {
	return s.tokenVerifier.VerifyAccess(accessToken)
}
