package middleware

import (
	"net/http"

	productv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/product/v1"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/clients"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductOwnershipMiddleware struct {
	productClient *clients.ProductClient
}

func NewProductOwnershipMiddleware(productClient *clients.ProductClient) *ProductOwnershipMiddleware {
	return &ProductOwnershipMiddleware{
		productClient: productClient,
	}
}

func (m *ProductOwnershipMiddleware) VerifySeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("product_id")
		if productID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
			c.Abort()
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Call product-service to get product details
		grpcReq := &productv1.GetProductRequest{
			ProductId: productID,
		}

		resp, err := m.productClient.GetProduct(c.Request.Context(), grpcReq)
		if err != nil {
			zap.L().Error("failed to get product for ownership verification",
				zap.String("product_id", productID),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify product ownership"})
			c.Abort()
			return
		}

		// Verify seller ownership
		if resp.Product.SellerId != userID.(string) {
			zap.L().Warn("user attempted to modify product they don't own",
				zap.String("user_id", userID.(string)),
				zap.String("product_id", productID),
				zap.String("actual_seller_id", resp.Product.SellerId),
			)
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not the seller of this product"})
			c.Abort()
			return
		}

		c.Next()
	}
}
