package product

import (
	"time"
)

// DomainEvent is the interface for all domain events
type DomainEvent interface {
	OccurredAt() time.Time
	EventType() string
}

// ProductCreatedEvent is emitted when a product is created
type ProductCreatedEvent struct {
	ProductID  ProductID
	SellerID   SellerID
	occurredAt time.Time
}

func NewProductCreatedEvent(productID ProductID, sellerID SellerID, occurredAt time.Time) ProductCreatedEvent {
	return ProductCreatedEvent{
		ProductID:  productID,
		SellerID:   sellerID,
		occurredAt: occurredAt,
	}
}

func (e ProductCreatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ProductCreatedEvent) EventType() string {
	return "product.created"
}

// ProductPublishedEvent is emitted when a product is published
type ProductPublishedEvent struct {
	ProductID  ProductID
	occurredAt time.Time
}

func NewProductPublishedEvent(productID ProductID, occurredAt time.Time) ProductPublishedEvent {
	return ProductPublishedEvent{
		ProductID:  productID,
		occurredAt: occurredAt,
	}
}

func (e ProductPublishedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ProductPublishedEvent) EventType() string {
	return "product.published"
}

// ProductDeactivatedEvent is emitted when a product is deactivated
type ProductDeactivatedEvent struct {
	ProductID  ProductID
	occurredAt time.Time
}

func NewProductDeactivatedEvent(productID ProductID, occurredAt time.Time) ProductDeactivatedEvent {
	return ProductDeactivatedEvent{
		ProductID:  productID,
		occurredAt: occurredAt,
	}
}

func (e ProductDeactivatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ProductDeactivatedEvent) EventType() string {
	return "product.deactivated"
}

// ProductSoldOutEvent is emitted when a product is sold out
type ProductSoldOutEvent struct {
	ProductID  ProductID
	occurredAt time.Time
}

func NewProductSoldOutEvent(productID ProductID, occurredAt time.Time) ProductSoldOutEvent {
	return ProductSoldOutEvent{
		ProductID:  productID,
		occurredAt: occurredAt,
	}
}

func (e ProductSoldOutEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ProductSoldOutEvent) EventType() string {
	return "product.sold_out"
}

// ProductSnapshotEvent represents a complete snapshot of active products
type ProductSnapshotEvent struct {
	GeneratedAt      time.Time
	ActiveProducts   []string
	PartitionOffsets map[int]int64 // partition_id -> offset at snapshot time
	Total            int
	occurredAt       time.Time
}

// NewProductSnapshotEvent creates a new snapshot event
func NewProductSnapshotEvent(
	activeProductIDs []string,
	partitionOffsets map[int]int64,
	occurredAt time.Time,
) *ProductSnapshotEvent {
	return &ProductSnapshotEvent{
		ActiveProducts:   activeProductIDs,
		PartitionOffsets: partitionOffsets,
		Total:            len(activeProductIDs),
		occurredAt:       occurredAt,
	}
}

// OccurredAt returns when the event occurred
func (e ProductSnapshotEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// EventType returns the event type
func (e ProductSnapshotEvent) EventType() string {
	return "product.snapshot"
}
