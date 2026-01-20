package handler

import (
	"net/http"
	"strconv"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/clients"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/common/errors"
	pb "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/order/v1"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderClient *clients.OrderClient
}

func NewOrderHandler(client *clients.OrderClient) *OrderHandler {
	return &OrderHandler{
		orderClient: client,
	}
}

// GetOrder handles GET /v1/orders/:order_id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("order_id")

	resp, err := h.orderClient.GetOrder(c.Request.Context(), &pb.GetOrderRequest{
		OrderId: orderID,
	})

	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListUserOrders handles GET /v1/orders
func (h *OrderHandler) ListUserOrders(c *gin.Context) {
	// Retrieve user_id from JWT middleware context
	userID := c.GetString("userID")

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	resp, err := h.orderClient.ListUserOrders(c.Request.Context(), &pb.ListUserOrdersRequest{
		UserId: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
