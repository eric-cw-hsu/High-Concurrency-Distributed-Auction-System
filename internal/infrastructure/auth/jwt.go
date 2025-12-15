package auth

import (
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	jwtConfig config.JWTConfig
}

func NewJWTService(jwtConfig config.JWTConfig) *JWTService {
	return &JWTService{
		jwtConfig: jwtConfig,
	}
}

func (j *JWTService) GenerateToken(user *user.User) (string, error) {
	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"iat":   time.Now().Unix(), // Issued at time
		"exp": time.Now().Add(
			time.Duration(j.jwtConfig.ExpiresIn) * time.Second,
		).Unix(), // Expiration time
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.jwtConfig.SecretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (j *JWTService) VerifyToken(token string) (*user.User, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(j.jwtConfig.SecretKey), nil
	})
	if err != nil || !parsedToken.Valid {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	user := &user.User{
		ID:    claims["id"].(string),
		Email: claims["email"].(string),
		Name:  claims["name"].(string),
	}

	return user, nil
}
