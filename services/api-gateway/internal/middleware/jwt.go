package middleware

import (
	"fmt"
	"net/http"
	"strings"

	authpb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/auth/v1"
	"github.com/gin-gonic/gin"
)

func NewJWTMiddleware(authClient authpb.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		accessToken := parts[1]
		resp, err := authClient.VerifyToken(c.Request.Context(), &authpb.VerifyRequest{
			Token: accessToken,
		})
		if err != nil || !resp.Valid {
			fmt.Println(err)
			fmt.Println(resp.Valid)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Store user ID in context for downstream handlers
		c.Set("userID", resp.UserId)
		c.Set("role", resp.Role)
		c.Next()
	}
}
