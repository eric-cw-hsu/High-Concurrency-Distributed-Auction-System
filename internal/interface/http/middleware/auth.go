package middleware

import (
	"net/http"
	"strings"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/user"
	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware(tokenService user.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		user, err := tokenService.VerifyToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Set to context for downstream usecase
		c.Set("user", user)

		c.Next()
	}
}
