package router

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/http/handler"
	"github.com/gin-gonic/gin"
)

// SetupWalletRoutes sets up wallet-related routes
func SetupWalletRoutes(router *gin.RouterGroup, walletHandler *handler.WalletHandler) {
	wallets := router.Group("/wallets")
	{
		// Wallet management
		wallets.POST("", walletHandler.CreateWallet)
		wallets.GET("/:userId", walletHandler.GetWallet)

		// Fund operations
		wallets.POST("/add-fund", walletHandler.AddFund)
		wallets.POST("/subtract-fund", walletHandler.SubtractFund)

		// Payment operations
		wallets.POST("/process-payment", walletHandler.ProcessPayment)
		wallets.POST("/process-refund", walletHandler.ProcessRefund)

		// Wallet management
		wallets.POST("/suspend", walletHandler.SuspendWallet)
		wallets.POST("/activate", walletHandler.ActivateWallet)
	}
}
