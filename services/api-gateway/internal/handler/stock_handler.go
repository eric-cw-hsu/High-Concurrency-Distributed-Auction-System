package handler

import (
	"net/http"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/clients"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/common/errors"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/api-gateway/internal/dto"
	stockv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/stock/v1"
	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	stockClient *clients.StockClient
}

func NewStockHandler(stockClient *clients.StockClient) *StockHandler {
	return &StockHandler{
		stockClient: stockClient,
	}
}

// SetStock handles POST /api/v1/stock/admin/products/:product_id/stock
func (h *StockHandler) SetStock(c *gin.Context) {
	productID := c.Param("product_id")

	var req dto.SetStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := h.stockClient.SetStock(c.Request.Context(), productID, req.Quantity)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, protoToStockResponse(grpcResp.Stock))
}

// GetStock handles GET /api/v1/stock/products/:product_id
func (h *StockHandler) GetStock(c *gin.Context) {
	productID := c.Param("product_id")

	grpcResp, err := h.stockClient.GetStock(c.Request.Context(), productID)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, protoToStockResponse(grpcResp.Stock))
}

// ReserveStock handles POST /api/v1/stock/reserve
func (h *StockHandler) ReserveStock(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := h.stockClient.Reserve(
		c.Request.Context(),
		req.ProductID,
		userID.(string),
		req.Quantity,
	)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ReserveStockResponse{
		Reservation:    protoToReservationResponse(grpcResp.Reservation),
		RemainingStock: grpcResp.RemainingStock,
	})
}

// ReleaseReservation handles DELETE /api/v1/stock/reservations/:reservation_id
func (h *StockHandler) ReleaseReservation(c *gin.Context) {
	reservationID := c.Param("reservation_id")

	grpcResp, err := h.stockClient.Release(c.Request.Context(), reservationID)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ReleaseReservationResponse{
		Success:  grpcResp.Success,
		NewStock: grpcResp.NewStock,
	})
}

// GetReservation handles GET /api/v1/stock/reservations/:reservation_id
func (h *StockHandler) GetReservation(c *gin.Context) {
	reservationID := c.Param("reservation_id")

	grpcResp, err := h.stockClient.GetReservation(c.Request.Context(), reservationID)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, protoToReservationResponse(grpcResp.Reservation))
}

// TriggerRecovery handles POST /api/v1/stock/admin/recovery
func (h *StockHandler) TriggerRecovery(c *gin.Context) {
	var req dto.TriggerRecoveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := h.stockClient.TriggerRecovery(c.Request.Context(), req.RecoveryType)
	if err != nil {
		errors.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.TriggerRecoveryResponse{
		Success:               grpcResp.Success,
		Message:               grpcResp.Message,
		ReservationsRecovered: grpcResp.ReservationsRecovered,
	})
}

// Helper functions

func protoToStockResponse(s *stockv1.Stock) dto.StockResponse {
	return dto.StockResponse{
		ProductID:       s.ProductId,
		Quantity:        s.Quantity,
		InitialQuantity: s.InitialQuantity,
		UpdatedAt:       s.UpdatedAt.AsTime(),
	}
}

func protoToReservationResponse(r *stockv1.Reservation) dto.ReservationResponse {
	return dto.ReservationResponse{
		ID:         r.Id,
		ProductID:  r.ProductId,
		UserID:     r.UserId,
		Quantity:   r.Quantity,
		Status:     r.Status,
		ReservedAt: r.ReservedAt.AsTime(),
		ExpiredAt:  r.ExpiredAt.AsTime(),
		OrderID:    r.OrderId,
	}
}
