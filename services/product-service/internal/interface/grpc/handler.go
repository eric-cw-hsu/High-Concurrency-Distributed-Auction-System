package grpc

import (
	"context"
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/common/logger"
	productv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/product/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProductHandler implements ProductService gRPC server
type ProductHandler struct {
	productv1.UnimplementedProductServiceServer
	productService *service.ProductService
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProduct creates a new product
func (h *ProductHandler) CreateProduct(
	ctx context.Context,
	req *productv1.CreateProductRequest,
) (*productv1.CreateProductResponse, error) {
	logger.InfoContext(ctx, "handling CreateProduct request",
		zap.String("seller_id", req.SellerId),
		zap.String("name", req.Name),
	)

	if err := validateCreateProductRequest(req); err != nil {
		logger.DebugContext(ctx, "invalid create product request",
			zap.String("error", err.Error()),
		)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	var flashSalePrice *int64
	if req.FlashSalePrice != nil {
		flashSalePrice = req.FlashSalePrice
	}

	p, err := h.productService.CreateProduct(
		ctx,
		req.SellerId,
		req.Name,
		req.Description,
		req.RegularPrice,
		flashSalePrice,
		req.Currency,
	)
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if isSystemError(code) {
			logger.ErrorContext(ctx, "failed to create product",
				zap.String("seller_id", req.SellerId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else if isBusinessError(code) {
			logger.WarnContext(ctx, "create product failed",
				zap.String("seller_id", req.SellerId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else {
			logger.DebugContext(ctx, "create product failed",
				zap.String("seller_id", req.SellerId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		}

		return nil, grpcErr
	}

	logger.InfoContext(ctx, "product created successfully",
		zap.String("product_id", p.ID().String()),
		zap.String("seller_id", req.SellerId),
	)

	return &productv1.CreateProductResponse{
		Product: domainToProto(p),
	}, nil
}

// UpdateProductInfo updates product information
func (h *ProductHandler) UpdateProductInfo(
	ctx context.Context,
	req *productv1.UpdateProductInfoRequest,
) (*productv1.UpdateProductInfoResponse, error) {
	logger.InfoContext(ctx, "handling UpdateProductInfo request",
		zap.String("product_id", req.ProductId),
	)

	if err := validateUpdateProductInfoRequest(req); err != nil {
		logger.DebugContext(ctx, "invalid update product info request",
			zap.String("error", err.Error()),
		)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	if err := h.productService.UpdateProductInfo(ctx, req.ProductId, req.Name, req.Description); err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if isSystemError(code) {
			logger.ErrorContext(ctx, "failed to update product info",
				zap.String("product_id", req.ProductId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else if isBusinessError(code) {
			logger.WarnContext(ctx, "update product info failed",
				zap.String("product_id", req.ProductId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		}

		return nil, grpcErr
	}

	p, err := h.productService.GetProduct(ctx, req.ProductId)
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		logger.ErrorContext(ctx, "failed to get product after update",
			zap.String("product_id", req.ProductId),
			zap.Error(err),
		)
		return nil, grpcErr
	}

	logger.InfoContext(ctx, "product info updated successfully",
		zap.String("product_id", req.ProductId),
	)

	return &productv1.UpdateProductInfoResponse{
		Product: domainToProto(p),
	}, nil
}

// UpdateProductPricing updates product pricing
func (h *ProductHandler) UpdateProductPricing(
	ctx context.Context,
	req *productv1.UpdateProductPricingRequest,
) (*productv1.UpdateProductPricingResponse, error) {
	logger.InfoContext(ctx, "handling UpdateProductPricing request",
		zap.String("product_id", req.ProductId),
	)

	if err := validateUpdateProductPricingRequest(req); err != nil {
		logger.DebugContext(ctx, "invalid update pricing request",
			zap.String("error", err.Error()),
		)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	var flashSalePrice *int64
	if req.FlashSalePrice != nil {
		flashSalePrice = req.FlashSalePrice
	}

	if err := h.productService.UpdateProductPricing(
		ctx,
		req.ProductId,
		req.RegularPrice,
		flashSalePrice,
		req.Currency,
	); err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if isSystemError(code) {
			logger.ErrorContext(ctx, "failed to update pricing",
				zap.String("product_id", req.ProductId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else if isBusinessError(code) {
			logger.WarnContext(ctx, "update pricing failed",
				zap.String("product_id", req.ProductId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		}

		return nil, grpcErr
	}

	p, err := h.productService.GetProduct(ctx, req.ProductId)
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		logger.ErrorContext(ctx, "failed to get product after pricing update",
			zap.String("product_id", req.ProductId),
			zap.Error(err),
		)
		return nil, grpcErr
	}

	logger.InfoContext(ctx, "product pricing updated successfully",
		zap.String("product_id", req.ProductId),
	)

	return &productv1.UpdateProductPricingResponse{
		Product: domainToProto(p),
	}, nil
}

// PublishProduct publishes a product
func (h *ProductHandler) PublishProduct(
	ctx context.Context,
	req *productv1.PublishProductRequest,
) (*productv1.PublishProductResponse, error) {
	logger.InfoContext(ctx, "handling PublishProduct request",
		zap.String("product_id", req.ProductId),
	)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	if err := h.productService.PublishProduct(ctx, req.ProductId); err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if isSystemError(code) {
			logger.ErrorContext(ctx, "failed to publish product",
				zap.String("product_id", req.ProductId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else if isBusinessError(code) {
			logger.WarnContext(ctx, "publish product failed",
				zap.String("product_id", req.ProductId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		}

		return nil, grpcErr
	}

	p, err := h.productService.GetProduct(ctx, req.ProductId)
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		logger.ErrorContext(ctx, "failed to get product after publish",
			zap.String("product_id", req.ProductId),
			zap.Error(err),
		)
		return nil, grpcErr
	}

	logger.InfoContext(ctx, "product published successfully",
		zap.String("product_id", req.ProductId),
	)

	return &productv1.PublishProductResponse{
		Product: domainToProto(p),
	}, nil
}

// DeactivateProduct deactivates a product
func (h *ProductHandler) DeactivateProduct(
	ctx context.Context,
	req *productv1.DeactivateProductRequest,
) (*productv1.DeactivateProductResponse, error) {
	logger.InfoContext(ctx, "handling DeactivateProduct request",
		zap.String("product_id", req.ProductId),
	)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	if err := h.productService.DeactivateProduct(ctx, req.ProductId); err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if isSystemError(code) {
			logger.ErrorContext(ctx, "failed to deactivate product",
				zap.String("product_id", req.ProductId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else if isBusinessError(code) {
			logger.WarnContext(ctx, "deactivate product failed",
				zap.String("product_id", req.ProductId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		}

		return nil, grpcErr
	}

	p, err := h.productService.GetProduct(ctx, req.ProductId)
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		logger.ErrorContext(ctx, "failed to get product after deactivate",
			zap.String("product_id", req.ProductId),
			zap.Error(err),
		)
		return nil, grpcErr
	}

	logger.InfoContext(ctx, "product deactivated successfully",
		zap.String("product_id", req.ProductId),
	)

	return &productv1.DeactivateProductResponse{
		Product: domainToProto(p),
	}, nil
}

// DeleteProduct deletes a product
func (h *ProductHandler) DeleteProduct(
	ctx context.Context,
	req *productv1.DeleteProductRequest,
) (*productv1.DeleteProductResponse, error) {
	logger.InfoContext(ctx, "handling DeleteProduct request",
		zap.String("product_id", req.ProductId),
		zap.String("seller_id", req.SellerId),
	)

	if req.ProductId == "" || req.SellerId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id and seller_id are required")
	}

	if err := h.productService.DeleteProduct(ctx, req.ProductId, req.SellerId); err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if isSystemError(code) {
			logger.ErrorContext(ctx, "failed to delete product",
				zap.String("product_id", req.ProductId),
				zap.String("seller_id", req.SellerId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		} else if isBusinessError(code) {
			logger.WarnContext(ctx, "delete product failed",
				zap.String("product_id", req.ProductId),
				zap.String("seller_id", req.SellerId),
				zap.String("error", err.Error()),
				zap.String("grpc_code", code.String()),
			)
		}

		return nil, grpcErr
	}

	logger.InfoContext(ctx, "product deleted successfully",
		zap.String("product_id", req.ProductId),
	)

	return &productv1.DeleteProductResponse{
		Success: true,
	}, nil
}

// GetProduct retrieves a product by ID
func (h *ProductHandler) GetProduct(
	ctx context.Context,
	req *productv1.GetProductRequest,
) (*productv1.GetProductResponse, error) {
	logger.DebugContext(ctx, "handling GetProduct request",
		zap.String("product_id", req.ProductId),
	)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	p, err := h.productService.GetProduct(ctx, req.ProductId)
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if code == codes.NotFound {
			logger.DebugContext(ctx, "product not found",
				zap.String("product_id", req.ProductId),
			)
		} else {
			logger.ErrorContext(ctx, "failed to get product",
				zap.String("product_id", req.ProductId),
				zap.Error(err),
			)
		}

		return nil, grpcErr
	}

	return &productv1.GetProductResponse{
		Product: domainToProto(p),
	}, nil
}

// GetActiveProducts retrieves all active products
func (h *ProductHandler) GetActiveProducts(
	ctx context.Context,
	req *productv1.GetActiveProductsRequest,
) (*productv1.GetActiveProductsResponse, error) {
	logger.DebugContext(ctx, "handling GetActiveProducts request",
		zap.Int32("page", req.Page),
		zap.Int32("page_size", req.PageSize),
	)

	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	products, err := h.productService.GetActiveProducts(ctx, int(pageSize), int(offset))
	if err != nil {
		logger.ErrorContext(ctx, "failed to get active products",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, "failed to get products")
	}

	total := len(products)

	protoProducts := make([]*productv1.Product, 0, len(products))
	for _, p := range products {
		protoProducts = append(protoProducts, domainToProto(p))
	}

	logger.DebugContext(ctx, "active products retrieved",
		zap.Int("count", len(products)),
	)

	return &productv1.GetActiveProductsResponse{
		Products: protoProducts,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetProductsBySeller retrieves products by seller
func (h *ProductHandler) GetProductsBySeller(
	ctx context.Context,
	req *productv1.GetProductsBySellerRequest,
) (*productv1.GetProductsBySellerResponse, error) {
	logger.DebugContext(ctx, "handling GetProductsBySeller request",
		zap.String("seller_id", req.SellerId),
		zap.Int32("page", req.Page),
		zap.Int32("page_size", req.PageSize),
	)

	if req.SellerId == "" {
		return nil, status.Error(codes.InvalidArgument, "seller_id is required")
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	products, err := h.productService.GetProductsBySeller(ctx, req.SellerId, int(pageSize), int(offset))
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		logger.ErrorContext(ctx, "failed to get products by seller",
			zap.String("seller_id", req.SellerId),
			zap.Error(err),
		)
		return nil, grpcErr
	}

	total := len(products)

	protoProducts := make([]*productv1.Product, 0, len(products))
	for _, p := range products {
		protoProducts = append(protoProducts, domainToProto(p))
	}

	logger.DebugContext(ctx, "seller products retrieved",
		zap.String("seller_id", req.SellerId),
		zap.Int("count", len(products)),
	)

	return &productv1.GetProductsBySellerResponse{
		Products: protoProducts,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Validation helpers

func validateCreateProductRequest(req *productv1.CreateProductRequest) error {
	if req.SellerId == "" {
		return fmt.Errorf("seller_id is required")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.RegularPrice <= 0 {
		return fmt.Errorf("regular_price must be positive")
	}
	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if req.FlashSalePrice != nil && *req.FlashSalePrice <= 0 {
		return fmt.Errorf("flash_sale_price must be positive")
	}
	if req.FlashSalePrice != nil && *req.FlashSalePrice >= req.RegularPrice {
		return fmt.Errorf("flash_sale_price must be less than regular_price")
	}
	return nil
}

func validateUpdateProductInfoRequest(req *productv1.UpdateProductInfoRequest) error {
	if req.ProductId == "" {
		return fmt.Errorf("product_id is required")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func validateUpdateProductPricingRequest(req *productv1.UpdateProductPricingRequest) error {
	if req.ProductId == "" {
		return fmt.Errorf("product_id is required")
	}
	if req.RegularPrice <= 0 {
		return fmt.Errorf("regular_price must be positive")
	}
	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if req.FlashSalePrice != nil && *req.FlashSalePrice <= 0 {
		return fmt.Errorf("flash_sale_price must be positive")
	}
	if req.FlashSalePrice != nil && *req.FlashSalePrice >= req.RegularPrice {
		return fmt.Errorf("flash_sale_price must be less than regular_price")
	}
	return nil
}
