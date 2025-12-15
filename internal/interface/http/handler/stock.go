package handler

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/stock"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/user"
	stockUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/stock"
	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	putOnMarketUsecase *stockUsecase.PutOnMarketUsecase
}

func NewStockHandler(putOnMarketUsecase *stockUsecase.PutOnMarketUsecase) *StockHandler {
	return &StockHandler{
		putOnMarketUsecase: putOnMarketUsecase,
	}
}

func (h *StockHandler) PutOnMarket(c *gin.Context) {
	var req stock.PutOnMarketCommand

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	req.SellerID = c.MustGet("user").(*user.User).ID // Assuming user ID is stored in context

	if err := h.putOnMarketUsecase.Execute(c.Request.Context(), req); err != nil {
		c.JSON(409, gin.H{"error": err.Error()})
		return
	}

	c.Status(201)
}
