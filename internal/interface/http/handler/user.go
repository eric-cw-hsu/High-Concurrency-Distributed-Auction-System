package handler

import (
	"net/http"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/user"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/httphelper"
	userUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/user"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	loginUserUsecase    *userUsecase.LoginUserUsecase
	registerUserUsecase *userUsecase.RegisterUserUsecase
}

func NewUserHandler(
	loginUserUsecase *userUsecase.LoginUserUsecase,
	registerUserUsecase *userUsecase.RegisterUserUsecase,
) *UserHandler {
	return &UserHandler{
		loginUserUsecase:    loginUserUsecase,
		registerUserUsecase: registerUserUsecase,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req user.RegisterUserCommand

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": httphelper.ParseValidationErrors(err),
		})
		return
	}

	user, err := h.registerUserUsecase.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req user.LoginUserCommand

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": httphelper.ParseValidationErrors(err),
		})
		return
	}

	user, token, err := h.loginUserUsecase.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"created_at": user.CreatedAt,
		},
		"token": token,
	})
}
