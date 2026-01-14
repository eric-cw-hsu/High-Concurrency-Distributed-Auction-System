package stock

import (
	"time"
)

const (
	// LowStockThresholdPercentage defines when stock is considered low (10% of initial)
	LowStockThresholdPercentage = 0.1

	// MaxDeductQuantity is the maximum quantity that can be deducted at once
	MaxDeductQuantity = 10
)

// Stock represents the inventory for a product
type Stock struct {
	productID         ProductID
	quantity          int     // current available quantity
	initialQuantity   int     // initial quantity (for low stock calculation)
	lowStockThreshold float64 // percentage threshold
	updatedAt         time.Time
	domainEvents      []DomainEvent
}

// NewStock creates a new stock
func NewStock(productID ProductID, quantity int) (*Stock, error) {
	if productID.IsEmpty() {
		return nil, ErrProductIDRequired
	}
	if quantity < 0 {
		return nil, ErrNegativeQuantity
	}

	return &Stock{
		productID:         productID,
		quantity:          quantity,
		initialQuantity:   quantity,
		lowStockThreshold: LowStockThresholdPercentage,
		updatedAt:         time.Now(),
	}, nil
}

// ReconstructStock reconstructs stock from persistence
func ReconstructStock(
	productID ProductID,
	quantity int,
	initialQuantity int,
	updatedAt time.Time,
) *Stock {
	return &Stock{
		productID:         productID,
		quantity:          quantity,
		initialQuantity:   initialQuantity,
		lowStockThreshold: LowStockThresholdPercentage,
		updatedAt:         updatedAt,
	}
}

// Getters
func (s *Stock) ProductID() ProductID {
	return s.productID
}

func (s *Stock) Quantity() int {
	return s.quantity
}

func (s *Stock) InitialQuantity() int {
	return s.initialQuantity
}

func (s *Stock) UpdatedAt() time.Time {
	return s.updatedAt
}

// Deduct deducts quantity from stock
func (s *Stock) Deduct(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	if quantity > MaxDeductQuantity {
		return ErrExceedsMaxQuantity
	}
	if s.quantity < quantity {
		return ErrInsufficientStock
	}

	s.quantity -= quantity
	s.updatedAt = time.Now()

	return nil
}

// Add adds quantity to stock
func (s *Stock) Add(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	s.quantity += quantity
	s.updatedAt = time.Now()

	return nil
}

// IsDepleted checks if stock is depleted
func (s *Stock) IsDepleted() bool {
	return s.quantity == 0
}

// IsLowStock checks if stock is below threshold (percentage-based)
func (s *Stock) IsLowStock() bool {
	if s.initialQuantity == 0 {
		return false
	}
	threshold := float64(s.initialQuantity) * s.lowStockThreshold
	return float64(s.quantity) < threshold
}

// GetLowStockThreshold returns the low stock threshold quantity
func (s *Stock) GetLowStockThreshold() int {
	return int(float64(s.initialQuantity) * s.lowStockThreshold)
}

// SetQuantity sets the stock quantity (for initial stock or replenishment)
func (s *Stock) SetQuantity(quantity int) error {
	if quantity < 0 {
		return ErrNegativeQuantity
	}

	s.quantity = quantity
	s.initialQuantity = quantity
	s.updatedAt = time.Now()

	return nil
}

// Domain events
func (s *Stock) recordEvent(event DomainEvent) {
	s.domainEvents = append(s.domainEvents, event)
}

func (s *Stock) DomainEvents() []DomainEvent {
	events := make([]DomainEvent, len(s.domainEvents))
	copy(events, s.domainEvents)
	return events
}

func (s *Stock) ClearEvents() {
	s.domainEvents = nil
}
