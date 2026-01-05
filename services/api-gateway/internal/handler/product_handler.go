package handler

import (
	"net/http"
	"strconv"

	productv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/product/v1"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/clients"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/common/errors"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/dto"
	"github.com/gin-gonic/gin"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	productClient *clients.ProductClient
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(productClient *clients.ProductClient) *ProductHandler {
	return &ProductHandler{
		productClient: productClient,
	}
}

// CreateProduct handles POST /api/v1/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// Extract seller_id from JWT (set by auth middleware)
	sellerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse request body
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate flash sale price
	if req.FlashSalePrice != nil && *req.FlashSalePrice >= req.RegularPrice {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flash_sale_price must be less than regular_price"})
		return
	}

	// Call Product Service via gRPC
	grpcReq := &productv1.CreateProductRequest{
		SellerId:       sellerID.(string),
		Name:           req.Name,
		Description:    req.Description,
		RegularPrice:   req.RegularPrice,
		FlashSalePrice: req.FlashSalePrice,
		Currency:       req.Currency,
	}

	grpcResp, err := h.productClient.CreateProduct(c.Request.Context(), grpcReq)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	// Convert to HTTP response
	c.JSON(http.StatusCreated, protoToProductResponse(grpcResp.Product))
}

// UpdateProductInfo handles PUT /api/v1/products/:id
func (h *ProductHandler) UpdateProductInfo(c *gin.Context) {
	productID := c.Param("id")

	var req dto.UpdateProductInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &productv1.UpdateProductInfoRequest{
		ProductId:   productID,
		Name:        req.Name,
		Description: req.Description,
	}

	grpcResp, err := h.productClient.UpdateProductInfo(c.Request.Context(), grpcReq)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, protoToProductResponse(grpcResp.Product))
}

// UpdateProductPricing handles PUT /api/v1/products/:id/pricing
func (h *ProductHandler) UpdateProductPricing(c *gin.Context) {
	productID := c.Param("id")

	var req dto.UpdateProductPricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate flash sale price
	if req.FlashSalePrice != nil && *req.FlashSalePrice >= req.RegularPrice {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flash_sale_price must be less than regular_price"})
		return
	}

	grpcReq := &productv1.UpdateProductPricingRequest{
		ProductId:      productID,
		RegularPrice:   req.RegularPrice,
		FlashSalePrice: req.FlashSalePrice,
		Currency:       req.Currency,
	}

	grpcResp, err := h.productClient.UpdateProductPricing(c.Request.Context(), grpcReq)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, protoToProductResponse(grpcResp.Product))
}

// PublishProduct handles POST /api/v1/products/:id/publish
func (h *ProductHandler) PublishProduct(c *gin.Context) {
	productID := c.Param("id")

	grpcReq := &productv1.PublishProductRequest{
		ProductId: productID,
	}

	grpcResp, err := h.productClient.PublishProduct(c.Request.Context(), grpcReq)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, protoToProductResponse(grpcResp.Product))
}

// DeactivateProduct handles POST /api/v1/products/:id/deactivate
func (h *ProductHandler) DeactivateProduct(c *gin.Context) {
	productID := c.Param("id")

	grpcReq := &productv1.DeactivateProductRequest{
		ProductId: productID,
	}

	grpcResp, err := h.productClient.DeactivateProduct(c.Request.Context(), grpcReq)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, protoToProductResponse(grpcResp.Product))
}

// DeleteProduct handles DELETE /api/v1/products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	// Extract seller_id from JWT
	sellerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grpcReq := &productv1.DeleteProductRequest{
		ProductId: productID,
		SellerId:  sellerID.(string),
	}

	_, err := h.productClient.DeleteProduct(c.Request.Context(), grpcReq)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}

// GetProduct handles GET /api/v1/products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")

	grpcReq := &productv1.GetProductRequest{
		ProductId: productID,
	}

	grpcResp, err := h.productClient.GetProduct(c.Request.Context(), grpcReq)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, protoToProductResponse(grpcResp.Product))
}

// GetActiveProducts handles GET /api/v1/products
func (h *ProductHandler) GetActiveProducts(c *gin.Context) {
	// Parse pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	grpcReq := &productv1.GetActiveProductsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	grpcResp, err := h.productClient.GetActiveProducts(c.Request.Context(), grpcReq)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	// Convert to HTTP response
	products := make([]dto.ProductResponse, 0, len(grpcResp.Products))
	for _, p := range grpcResp.Products {
		products = append(products, protoToProductResponse(p))
	}

	response := dto.ProductListResponse{
		Products: products,
		Total:    grpcResp.Total,
		Page:     grpcResp.Page,
		PageSize: grpcResp.PageSize,
	}

	c.JSON(http.StatusOK, response)
}

// GetProductsBySeller handles GET /api/v1/sellers/:id/products
func (h *ProductHandler) GetProductsBySeller(c *gin.Context) {
	sellerID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	grpcReq := &productv1.GetProductsBySellerRequest{
		SellerId: sellerID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	grpcResp, err := h.productClient.GetProductsBySeller(c.Request.Context(), grpcReq)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	products := make([]dto.ProductResponse, 0, len(grpcResp.Products))
	for _, p := range grpcResp.Products {
		products = append(products, protoToProductResponse(p))
	}

	response := dto.ProductListResponse{
		Products: products,
		Total:    grpcResp.Total,
		Page:     grpcResp.Page,
		PageSize: grpcResp.PageSize,
	}

	c.JSON(http.StatusOK, response)
}

// protoToProductResponse converts proto Product to DTO
func protoToProductResponse(p *productv1.Product) dto.ProductResponse {
	pricing := dto.PricingDTO{
		RegularPrice: dto.MoneyDTO{
			Amount:   p.Pricing.RegularPrice.Amount,
			Currency: p.Pricing.RegularPrice.Currency,
		},
	}

	if p.Pricing.FlashSalePrice != nil {
		pricing.FlashSalePrice = &dto.MoneyDTO{
			Amount:   p.Pricing.FlashSalePrice.Amount,
			Currency: p.Pricing.FlashSalePrice.Currency,
		}
	}

	return dto.ProductResponse{
		ID:          p.Id,
		SellerID:    p.SellerId,
		Name:        p.Name,
		Description: p.Description,
		Pricing:     pricing,
		Status:      p.Status,
		StockStatus: p.StockStatus,
		CreatedAt:   p.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}
