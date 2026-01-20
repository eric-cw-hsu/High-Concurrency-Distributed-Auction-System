package grpc

import (
	"context"
	"fmt"

	productprice "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/product_price"
	pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/product/v1"
	"go.uber.org/zap"
)

type productClient struct {
	client pb.ProductServiceClient
}

func NewProductClient(client pb.ProductServiceClient) productprice.ProductClient {
	return &productClient{client: client}
}

func (c *productClient) FetchProductDetail(ctx context.Context, productID string) (*productprice.ProductPrice, error) {
	// Boundary Logging: (Outbound Call)
	zap.L().Debug("fetching product detail via gRPC", zap.String("product_id", productID))

	resp, err := c.client.GetProduct(ctx, &pb.GetProductRequest{ProductId: productID})
	if err != nil {
		// Boundary Logging: RPC Fail
		zap.L().Error("gRPC product client error",
			zap.String("product_id", productID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}
	var money *pb.Money
	if resp.Product.Pricing.FlashSalePrice != nil {
		money = resp.Product.Pricing.FlashSalePrice
	} else {
		money = resp.Product.Pricing.RegularPrice
	}

	// Convert Proto to Domain Model
	return productprice.NewProductPrice(
		resp.Product.Id,
		money.Amount,
		money.Currency,
	), nil
}
