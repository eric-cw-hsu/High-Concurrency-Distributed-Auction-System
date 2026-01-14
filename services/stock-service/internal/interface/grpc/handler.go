package grpc

import (
	"context"
	"fmt"

	stockv1 "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared/proto/stock/v1"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/application/service"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/infrastructure/recovery"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StockHandler implements StockService gRPC server
type StockHandler struct {
	stockv1.UnimplementedStockServiceServer
	stockService *service.StockService
	recovery     *recovery.RedisRecovery
}

// NewStockHandler creates a new StockHandler
func NewStockHandler(
	stockService *service.StockService,
	recovery *recovery.RedisRecovery,
) *StockHandler {
	return &StockHandler{
		stockService: stockService,
		recovery:     recovery,
	}
}

// SetStock sets initial stock for a product
func (h *StockHandler) SetStock(
	ctx context.Context,
	req *stockv1.SetStockRequest,
) (*stockv1.SetStockResponse, error) {
	logger.InfoContext(ctx, "handling SetStock request",
		zap.String("product_id", req.ProductId),
		zap.Int32("quantity", req.Quantity),
	)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	if req.Quantity < 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity cannot be negative")
	}

	if err := h.stockService.SetStock(ctx, req.ProductId, int(req.Quantity)); err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		logError(ctx, grpcErr, "set stock failed",
			zap.String("product_id", req.ProductId),
			zap.String("error", err.Error()),
		)
		return nil, grpcErr
	}

	// Get updated stock
	stk, err := h.stockService.GetStock(ctx, req.ProductId)
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		logger.ErrorContext(ctx, "failed to get stock after set",
			zap.String("product_id", req.ProductId),
			zap.Error(err),
		)
		return nil, grpcErr
	}

	logger.InfoContext(ctx, "stock set successfully",
		zap.String("product_id", req.ProductId),
		zap.Int32("quantity", req.Quantity),
	)

	return &stockv1.SetStockResponse{
		Stock: domainStockToProto(stk),
	}, nil
}

// GetStock gets current stock for a product
func (h *StockHandler) GetStock(
	ctx context.Context,
	req *stockv1.GetStockRequest,
) (*stockv1.GetStockResponse, error) {
	logger.DebugContext(ctx, "handling GetStock request",
		zap.String("product_id", req.ProductId),
	)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	stk, err := h.stockService.GetStock(ctx, req.ProductId)
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if code == codes.NotFound {
			logger.DebugContext(ctx, "stock not found",
				zap.String("product_id", req.ProductId),
			)
		} else {
			logger.ErrorContext(ctx, "failed to get stock",
				zap.String("product_id", req.ProductId),
				zap.Error(err),
			)
		}

		return nil, grpcErr
	}

	return &stockv1.GetStockResponse{
		Stock: domainStockToProto(stk),
	}, nil
}

// Reserve reserves stock for a user
func (h *StockHandler) Reserve(
	ctx context.Context,
	req *stockv1.ReserveRequest,
) (*stockv1.ReserveResponse, error) {
	logger.InfoContext(ctx, "handling Reserve request",
		zap.String("product_id", req.ProductId),
		zap.String("user_id", req.UserId),
		zap.Int32("quantity", req.Quantity),
	)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
	}
	if req.Quantity > 10 {
		return nil, status.Error(codes.InvalidArgument, "quantity cannot exceed 10")
	}

	res, remainingStock, err := h.stockService.Reserve(ctx, req.ProductId, req.UserId, int(req.Quantity))
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		logError(ctx, grpcErr, "reserve stock failed",
			zap.String("product_id", req.ProductId),
			zap.String("user_id", req.UserId),
			zap.Int32("quantity", req.Quantity),
			zap.String("error", err.Error()),
		)
		return nil, grpcErr
	}

	logger.InfoContext(ctx, "stock reserved successfully",
		zap.String("product_id", req.ProductId),
		zap.String("reservation_id", res.ID().String()),
		zap.String("user_id", req.UserId),
		zap.Int("remaining_stock", remainingStock),
	)

	return &stockv1.ReserveResponse{
		Reservation:    domainReservationToProto(res),
		RemainingStock: int32(remainingStock),
	}, nil
}

// Release releases a reservation
func (h *StockHandler) Release(
	ctx context.Context,
	req *stockv1.ReleaseRequest,
) (*stockv1.ReleaseResponse, error) {
	logger.InfoContext(ctx, "handling Release request",
		zap.String("reservation_id", req.ReservationId),
	)

	if req.ReservationId == "" {
		return nil, status.Error(codes.InvalidArgument, "reservation_id is required")
	}

	newStock, err := h.stockService.Release(ctx, req.ReservationId)
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		logError(ctx, grpcErr, "release reservation failed",
			zap.String("reservation_id", req.ReservationId),
			zap.String("error", err.Error()),
		)
		return nil, grpcErr
	}

	logger.InfoContext(ctx, "reservation released successfully",
		zap.String("reservation_id", req.ReservationId),
		zap.Int("new_stock", newStock),
	)

	return &stockv1.ReleaseResponse{
		Success:  true,
		NewStock: int32(newStock),
	}, nil
}

// GetReservation gets reservation details
func (h *StockHandler) GetReservation(
	ctx context.Context,
	req *stockv1.GetReservationRequest,
) (*stockv1.GetReservationResponse, error) {
	logger.DebugContext(ctx, "handling GetReservation request",
		zap.String("reservation_id", req.ReservationId),
	)

	if req.ReservationId == "" {
		return nil, status.Error(codes.InvalidArgument, "reservation_id is required")
	}

	res, err := h.stockService.GetReservation(ctx, req.ReservationId)
	if err != nil {
		grpcErr := mapDomainErrorToGRPC(err)
		code := status.Code(grpcErr)

		if code == codes.NotFound {
			logger.DebugContext(ctx, "reservation not found",
				zap.String("reservation_id", req.ReservationId),
			)
		} else {
			logger.ErrorContext(ctx, "failed to get reservation",
				zap.String("reservation_id", req.ReservationId),
				zap.Error(err),
			)
		}

		return nil, grpcErr
	}

	return &stockv1.GetReservationResponse{
		Reservation: domainReservationToProto(res),
	}, nil
}

// TriggerRecovery triggers Redis recovery (admin operation)
func (h *StockHandler) TriggerRecovery(
	ctx context.Context,
	req *stockv1.TriggerRecoveryRequest,
) (*stockv1.TriggerRecoveryResponse, error) {
	logger.InfoContext(ctx, "admin: triggering recovery",
		zap.String("recovery_type", req.RecoveryType),
	)

	// TODO: Add authorization check (only admin can trigger)

	var err error
	var count int

	switch req.RecoveryType {
	case "reservations":
		err = h.recovery.RecoverActiveReservations(ctx)
	case "full":
		err = h.recovery.FullRecovery(ctx)
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid recovery type, must be 'reservations' or 'full'")
	}

	if err != nil {
		logger.ErrorContext(ctx, "recovery failed",
			zap.String("recovery_type", req.RecoveryType),
			zap.Error(err),
		)
		return &stockv1.TriggerRecoveryResponse{
			Success: false,
			Message: fmt.Sprintf("recovery failed: %v", err),
		}, nil
	}

	logger.InfoContext(ctx, "recovery completed successfully",
		zap.String("recovery_type", req.RecoveryType),
	)

	return &stockv1.TriggerRecoveryResponse{
		Success:               true,
		Message:               "recovery completed successfully",
		ReservationsRecovered: int32(count),
	}, nil
}

// logError logs error based on gRPC code classification
func logError(ctx context.Context, grpcErr error, msg string, fields ...zap.Field) {
	code := status.Code(grpcErr)
	allFields := append(fields, zap.String("grpc_code", code.String()))

	if isSystemError(code) {
		logger.ErrorContext(ctx, msg, allFields...)
	} else if isBusinessError(code) {
		logger.WarnContext(ctx, msg, allFields...)
	} else {
		logger.DebugContext(ctx, msg, allFields...)
	}
}
