package handler

import (
	"net/http"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/order"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/user"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/httphelper"
	orderUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/order"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	placeOrderUsecase *orderUsecase.PlaceOrderUsecase
}

func NewOrderHandler(placeOrderUsecase *orderUsecase.PlaceOrderUsecase) *OrderHandler {
	return &OrderHandler{
		placeOrderUsecase: placeOrderUsecase,
	}
}

func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req order.PlaceOrderCommand

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": httphelper.ParseValidationErrors(err),
		})
		return
	}

	req.BuyerId = c.MustGet("user").(*user.User).Id

	err := h.placeOrderUsecase.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusCreated)
}
