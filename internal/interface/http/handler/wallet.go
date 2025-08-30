package handler

import (
	"net/http"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/httphelper"
	walletUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/wallet"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	addFundUsecase      *walletUsecase.AddFundUsecase
	subtractFundUsecase *walletUsecase.SubtractFundUsecase
	createWalletUsecase *walletUsecase.CreateWalletUsecase
}

func NewWalletHandler(
	addFundUsecase *walletUsecase.AddFundUsecase,
	subtractFundUsecase *walletUsecase.SubtractFundUsecase,
	createWalletUsecase *walletUsecase.CreateWalletUsecase,
) *WalletHandler {
	return &WalletHandler{
		addFundUsecase:      addFundUsecase,
		subtractFundUsecase: subtractFundUsecase,
		createWalletUsecase: createWalletUsecase,
	}
}

// CreateWallet creates a new wallet for a user
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var command wallet.CreateWalletCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		httphelper.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// For now, create wallet if needed using CreateWallet usecase
	walletInfo, err := h.createWalletUsecase.Execute(c.Request.Context(), command.UserId)
	if err != nil {
		httphelper.ErrorResponse(c, http.StatusInternalServerError, "Failed to create wallet", err)
		return
	}

	httphelper.SuccessResponse(c, http.StatusCreated, "Wallet created successfully", walletInfo)
}

// GetWallet retrieves wallet information for a user
func (h *WalletHandler) GetWallet(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		httphelper.ErrorResponse(c, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	// Get wallet - for now we'll create if not exists (should be a separate GetWallet usecase)
	walletInfo, err := h.createWalletUsecase.Execute(c.Request.Context(), userId)
	if err != nil {
		httphelper.ErrorResponse(c, http.StatusInternalServerError, "Failed to get wallet", err)
		return
	}

	if walletInfo == nil {
		httphelper.ErrorResponse(c, http.StatusNotFound, "Wallet not found", nil)
		return
	}

	httphelper.SuccessResponse(c, http.StatusOK, "Wallet retrieved successfully", walletInfo)
}

// AddFund adds funds to a wallet
func (h *WalletHandler) AddFund(c *gin.Context) {
	var command wallet.AddFundCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		httphelper.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	walletInfo, err := h.addFundUsecase.Execute(c.Request.Context(), &command)
	if err != nil {
		httphelper.ErrorResponse(c, http.StatusInternalServerError, "Failed to add fund", err)
		return
	}

	httphelper.SuccessResponse(c, http.StatusOK, "Fund added successfully", walletInfo)
}

// SubtractFund subtracts funds from a wallet
func (h *WalletHandler) SubtractFund(c *gin.Context) {
	var command wallet.SubtractFundCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		httphelper.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	walletInfo, err := h.subtractFundUsecase.Execute(c.Request.Context(), &command)
	if err != nil {
		httphelper.ErrorResponse(c, http.StatusInternalServerError, "Failed to subtract fund", err)
		return
	}

	httphelper.SuccessResponse(c, http.StatusOK, "Fund subtracted successfully", walletInfo)
}

// ProcessPayment processes a payment from a wallet
func (h *WalletHandler) ProcessPayment(c *gin.Context) {
	// TODO: Implement ProcessPaymentUsecase
	httphelper.ErrorResponse(c, http.StatusNotImplemented, "ProcessPayment not yet implemented with new usecase pattern", nil)
}

// ProcessRefund processes a refund to a wallet
func (h *WalletHandler) ProcessRefund(c *gin.Context) {
	// TODO: Implement ProcessRefundUsecase
	httphelper.ErrorResponse(c, http.StatusNotImplemented, "ProcessRefund not yet implemented with new usecase pattern", nil)
}

// SuspendWallet suspends a wallet
func (h *WalletHandler) SuspendWallet(c *gin.Context) {
	// TODO: Implement SuspendWalletUsecase
	httphelper.ErrorResponse(c, http.StatusNotImplemented, "SuspendWallet not yet implemented with new usecase pattern", nil)
}

// ActivateWallet activates a wallet
func (h *WalletHandler) ActivateWallet(c *gin.Context) {
	// TODO: Implement ActivateWalletUsecase
	httphelper.ErrorResponse(c, http.StatusNotImplemented, "ActivateWallet not yet implemented with new usecase pattern", nil)
}
