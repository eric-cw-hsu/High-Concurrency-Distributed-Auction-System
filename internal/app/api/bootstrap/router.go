package bootstrap

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/http/middleware"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/http/router"
	"github.com/gin-gonic/gin"
)

func RouterSetup(deps *Dependencies) *gin.Engine {
	r := gin.Default()

	// Setup all routes using the centralized router
	router.SetupRoutes(r,
		deps.UserHandler,
		deps.WalletHandler,
		deps.ProductHandler,
		deps.StockHandler,
		deps.OrderHandler)

	// Apply middleware to protected routes
	v1 := r.Group("/api/v1")
	protectedRoutes := v1.Group("")
	protectedRoutes.Use(middleware.JWTAuthMiddleware(deps.TokenService))
	{
		// Add any protected-only routes here if needed
	}

	return r
}
