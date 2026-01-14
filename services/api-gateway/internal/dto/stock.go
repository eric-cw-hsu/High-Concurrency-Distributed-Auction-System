package dto

import "time"

// Stock DTOs

type SetStockRequest struct {
	Quantity int32 `json:"quantity" binding:"required,min=0"`
}

type StockResponse struct {
	ProductID       string    `json:"product_id"`
	Quantity        int32     `json:"quantity"`
	InitialQuantity int32     `json:"initial_quantity"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Reservation DTOs

type ReserveStockRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int32  `json:"quantity" binding:"required,min=1,max=10"`
}

type ReserveStockResponse struct {
	Reservation    ReservationResponse `json:"reservation"`
	RemainingStock int32               `json:"remaining_stock"`
}

type ReleaseReservationRequest struct {
	ReservationID string `json:"reservation_id" binding:"required"`
}

type ReleaseReservationResponse struct {
	Success  bool  `json:"success"`
	NewStock int32 `json:"new_stock"`
}

type GetReservationRequest struct {
	ReservationID string `json:"reservation_id" binding:"required"`
}

type ReservationResponse struct {
	ID         string    `json:"id"`
	ProductID  string    `json:"product_id"`
	UserID     string    `json:"user_id"`
	Quantity   int32     `json:"quantity"`
	Status     string    `json:"status"`
	ReservedAt time.Time `json:"reserved_at"`
	ExpiredAt  time.Time `json:"expired_at"`
	OrderID    *string   `json:"order_id,omitempty"`
}

// Admin DTOs

type TriggerRecoveryRequest struct {
	RecoveryType string `json:"recovery_type" binding:"required,oneof=reservations stock full"`
}

type TriggerRecoveryResponse struct {
	Success               bool   `json:"success"`
	Message               string `json:"message"`
	ReservationsRecovered int32  `json:"reservations_recovered"`
}
