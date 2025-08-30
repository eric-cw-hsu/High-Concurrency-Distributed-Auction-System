package router

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/http/handler"
	"github.com/gin-gonic/gin"
)

// SetupUserRoutes sets up user-related routes
func SetupUserRoutes(router *gin.RouterGroup, userHandler *handler.UserHandler) {
	// User authentication routes
	router.POST("/register", userHandler.Register)
	router.POST("/login", userHandler.Login)
}
