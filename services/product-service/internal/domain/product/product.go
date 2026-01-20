package product

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var PRODUCT_NAME_MAX_LENGTH = 200

// Product is the aggregate root for the product domain
type Product struct {
	id           ProductID
	sellerID     SellerID
	name         string
	description  string
	pricing      Pricing
	status       ProductStatus
	stockStatus  StockStatus
	createdAt    time.Time
	updatedAt    time.Time
	domainEvents []DomainEvent
}

// NewProduct creates a new product (factory method)
func NewProduct(
	sellerID SellerID,
	name string,
	description string,
	pricing Pricing,
) (*Product, error) {
	// Validate Name
	if len(name) == 0 {
		return nil, errors.New("product name cannot be empty")
	}
	if len(name) > PRODUCT_NAME_MAX_LENGTH {
		return nil, errors.New(fmt.Sprintf("product name too long (max %d characters)", PRODUCT_NAME_MAX_LENGTH))
	}

	// Validate description
	description = strings.TrimSpace(description)
	if len(description) > 5000 {
		return nil, errors.New("description too long (max 5000 characters)")
	}

	if sellerID.IsEmpty() {
		return nil, errors.New("seller id is required")
	}

	now := time.Now()
	productID := NewProductID()

	p := &Product{
		id:          productID,
		sellerID:    sellerID,
		name:        name,
		description: description,
		pricing:     pricing,
		status:      ProductStatusDraft,
		stockStatus: StockStatusUnknown,
		createdAt:   now,
		updatedAt:   now,
	}

	p.recordEvent(NewProductCreatedEvent(productID, sellerID, now))

	return p, nil
}

// ReconstructProduct reconstructs a product from persistence (for Repository use)
func ReconstructProduct(
	id ProductID,
	sellerID SellerID,
	name string,
	description string,
	pricing Pricing,
	status ProductStatus,
	stockStatus StockStatus,
	createdAt time.Time,
	updatedAt time.Time,
) *Product {
	return &Product{
		id:          id,
		sellerID:    sellerID,
		name:        name,
		description: description,
		pricing:     pricing,
		status:      status,
		stockStatus: stockStatus,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Getters - aggregate root exposes read-only access
func (p *Product) ID() ProductID {
	return p.id
}

func (p *Product) SellerID() SellerID {
	return p.sellerID
}

func (p *Product) Name() string {
	return p.name
}

func (p *Product) Description() string {
	return p.description
}

func (p *Product) Pricing() Pricing {
	return p.pricing
}

func (p *Product) Status() ProductStatus {
	return p.status
}

func (p *Product) StockStatus() StockStatus {
	return p.stockStatus
}

func (p *Product) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Product) UpdatedAt() time.Time {
	return p.updatedAt
}

// Publish publishes the product (makes it available for sale)
func (p *Product) Publish() error {
	if !p.status.CanPublish() {
		return ErrCannotPublishProduct
	}

	p.status = ProductStatusActive
	p.updatedAt = time.Now()

	var money Money
	if p.pricing.flashSalePrice != nil {
		money = *p.pricing.flashSalePrice
	} else {
		money = p.pricing.regularPrice
	}

	p.recordEvent(NewProductPublishedEvent(p.id, money, p.updatedAt))

	return nil
}

// Deactivate deactivates the product (removes from sale)
func (p *Product) Deactivate() error {
	if !p.status.CanDeactivate() {
		return ErrCannotDeactivateProduct
	}

	p.status = ProductStatusInactive
	p.updatedAt = time.Now()

	p.recordEvent(NewProductDeactivatedEvent(p.id, p.updatedAt))

	return nil
}

// MarkAsSoldOut marks the product as sold out (triggered by stock.depleted event)
func (p *Product) MarkAsSoldOut() error {
	if !p.status.CanMarkAsSoldOut() {
		return ErrCannotMarkAsSoldOut
	}

	p.status = ProductStatusSoldOut
	p.stockStatus = StockStatusOutOfStock
	p.updatedAt = time.Now()

	p.recordEvent(NewProductSoldOutEvent(p.id, p.updatedAt))

	return nil
}

// UpdateInfo updates product name and description
func (p *Product) UpdateInfo(name string, description string) error {
	if !p.status.CanUpdate() {
		return ErrCannotUpdateActiveProduct
	}

	// Validate name
	name = strings.TrimSpace(name)
	if len(name) == 0 || len(name) > 200 {
		return errors.New("invalid product name")
	}

	// Validate description
	description = strings.TrimSpace(description)
	if len(description) > 5000 {
		return errors.New("description too long")
	}

	p.name = name
	p.description = description
	p.updatedAt = time.Now()

	return nil
}

// UpdatePricing updates product pricing
func (p *Product) UpdatePricing(newPricing Pricing) error {
	if !p.status.CanUpdate() {
		return ErrCannotUpdatePricingForActiveProduct
	}

	p.pricing = newPricing
	p.updatedAt = time.Now()

	return nil
}

// UpdateStockStatus updates stock status (called when consuming stock events)
func (p *Product) UpdateStockStatus(newStatus StockStatus) {
	p.stockStatus = newStatus
	p.updatedAt = time.Now()
}

// CanBeDeletedBy checks if the product can be deleted by the given seller
func (p *Product) CanBeDeletedBy(sellerID SellerID) bool {
	// Business rule: only the seller can delete
	// Business rule: only draft or inactive products can be deleted
	return p.sellerID.Equals(sellerID) &&
		(p.status == ProductStatusDraft || p.status == ProductStatusInactive)
}

// Domain event management
func (p *Product) recordEvent(event DomainEvent) {
	p.domainEvents = append(p.domainEvents, event)
}

// DomainEvents returns a copy of domain events
func (p *Product) DomainEvents() []DomainEvent {
	// Return a copy to prevent external modification
	events := make([]DomainEvent, len(p.domainEvents))
	copy(events, p.domainEvents)
	return events
}

// ClearEvents clears all domain events (called after publishing)
func (p *Product) ClearEvents() {
	p.domainEvents = nil
}
