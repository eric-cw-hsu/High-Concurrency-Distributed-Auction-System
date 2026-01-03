package jwt

import (
	"errors"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/auth/model"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/port"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/config"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
	"github.com/golang-jwt/jwt/v5"
)

type JWTProvider struct {
	accessSecret  []byte
	refreshSecret []byte
	issuer        string
	audience      string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

var _ port.TokenIssuer = (*JWTProvider)(nil)
var _ port.TokenVerifier = (*JWTProvider)(nil)

func NewJWTProvider(cfg config.JWTConfig) *JWTProvider {
	return &JWTProvider{
		accessSecret:  []byte(cfg.AccessSecretKey),
		refreshSecret: []byte(cfg.RefreshSecretKey),
		issuer:        cfg.Issuer,
		audience:      cfg.Audience,
		accessTTL:     time.Duration(cfg.AccessTTL) * time.Second,
		refreshTTL:    time.Duration(cfg.RefreshTTL) * time.Second,
	}
}

func (p *JWTProvider) IssueAccess(userID kernel.UserID, role kernel.Role) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"iss":  p.issuer,
		"aud":  p.audience,
		"iat":  now.Unix(),
		"exp":  now.Add(p.accessTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(p.accessSecret)
}

func (p *JWTProvider) IssueRefresh(tokenID kernel.TokenID, userID kernel.UserID) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(p.refreshTTL)
	claims := jwt.MapClaims{
		"sub": string(userID),
		"jti": string(tokenID),
		"iss": p.issuer,
		"aud": p.audience,
		"iat": now.Unix(),
		"exp": exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(p.refreshSecret)
	return signedToken, exp, err
}

func (p *JWTProvider) VerifyAccess(tokenStr string) (*model.Access, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return p.accessSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Invalid claims")
	}

	subClaim, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("Invalid user_id")
	}
	var userID kernel.UserID
	userID, err = kernel.NewUserID(subClaim)
	if err != nil {
		return nil, errors.New("Invalid user_id")
	}

	roleClaim, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("Invalid role")
	}
	var role kernel.Role
	role, err = kernel.NewRole(roleClaim)
	if err != nil {
		return nil, errors.New("Invalid role")
	}

	return &model.Access{
		UserID:   userID,
		Role:     role,
		IssuedAt: time.Unix(int64(claims["iat"].(float64)), 0),
		ExpireAt: time.Unix(int64(claims["exp"].(float64)), 0),
	}, nil
}

func (j *JWTProvider) VerifyRefresh(tokenStr string) (kernel.TokenID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.refreshSecret, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	jtiClaim, ok := claims["jti"].(string)
	if !ok {
		return "", errors.New("invalid token_id")
	}

	return kernel.NewTokenID(jtiClaim)
}
