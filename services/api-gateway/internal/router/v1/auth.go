package v1

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterAuth(r *gin.RouterGroup, auth *handler.AuthHandler, jwtMiddleware gin.HandlerFunc) {
	r.POST("/register", auth.Register)
	r.POST("/login", auth.Login)
	r.POST("/refresh", auth.RefreshToken)

	secured := r.Group("/me")
	secured.Use(jwtMiddleware)
	{
		secured.GET("/", gin.HandlerFunc(func(c *gin.Context) {
			userID, _ := c.Get("userID")
			role, _ := c.Get("Role")
			c.JSON(200, gin.H{
				"userID": userID,
				"role":   role,
			})
		}))
	}
}
