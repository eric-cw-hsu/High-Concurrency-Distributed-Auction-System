package handler

import (
	"context"
	"net/http"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/clients"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/dto"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authClient *clients.AuthClient
}

func NewAuthHandler(authClient *clients.AuthClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

// POST /register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.authClient.Register(context.Background(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.RegisterResponse{UserID: userID})
}

// POST /login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, accessToken, refreshToken, err := h.authClient.Login(context.Background(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := h.authClient.RefreshToken(context.Background(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.RefreshTokenResponse{
		AccessToken: accessToken,
	})
}
