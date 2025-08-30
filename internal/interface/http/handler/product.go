package handler

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/product"
	productUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/product"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	createProductUsecase *productUsecase.CreateProductUsecase
}

func NewProductHandler(createProductUsecase *productUsecase.CreateProductUsecase) *ProductHandler {
	return &ProductHandler{
		createProductUsecase: createProductUsecase,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req product.CreateProductCommand

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	product, err := h.createProductUsecase.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(409, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"id":          product.Id,
		"name":        product.Name,
		"description": product.Description,
		"created_at":  product.CreatedAt,
	})
}
