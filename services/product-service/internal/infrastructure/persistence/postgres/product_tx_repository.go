package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service/internal/domain/product"
	"github.com/jmoiron/sqlx"
	"github.com/samborkent/uuidv7"
)

// ProductTxRepository handles transactional writes for products with events
type ProductTxRepository struct {
	db *sqlx.DB
}

// NewProductWriter creates a new ProductTxRepository
func NewProductWriter(db *sqlx.DB) *ProductTxRepository {
	return &ProductTxRepository{db: db}
}

// Save saves a product and publishes events in a transaction
func (w *ProductTxRepository) Save(ctx context.Context, p *product.Product) error {
	// Begin transaction
	tx, err := w.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Save product
	model := DomainToModel(p)

	productQuery := `
		INSERT INTO products (
			id, seller_id, name, description,
			regular_price, flash_sale_price, currency,
			status, stock_status, created_at, updated_at
		) VALUES (
			:id, :seller_id, :name, :description,
			:regular_price, :flash_sale_price, :currency,
			:status, :stock_status, :created_at, :updated_at
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			regular_price = EXCLUDED.regular_price,
			flash_sale_price = EXCLUDED.flash_sale_price,
			currency = EXCLUDED.currency,
			status = EXCLUDED.status,
			stock_status = EXCLUDED.stock_status,
			updated_at = EXCLUDED.updated_at
	`

	_, err = tx.NamedExecContext(ctx, productQuery, model)
	if err != nil {
		return fmt.Errorf("failed to save product: %w", err)
	}

	// 2. Insert outbox events
	events := p.DomainEvents()
	for _, event := range events {
		if err := w.insertOutboxEvent(ctx, tx, p, event); err != nil {
			return fmt.Errorf("failed to insert outbox event: %w", err)
		}
	}

	// 3. Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 4. Clear events after successful commit
	p.ClearEvents()

	return nil
}

// insertOutboxEvent inserts a single outbox event within transaction
func (w *ProductTxRepository) insertOutboxEvent(
	ctx context.Context,
	tx *sqlx.Tx,
	p *product.Product,
	event product.DomainEvent,
) error {
	// Create outbox event
	eventID := uuidv7.New().String()
	id := uuidv7.New().String()

	payload := w.domainEventToPayload(event)
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Insert to outbox_events table
	query := `
		INSERT INTO outbox_events (
			id, aggregate_type, aggregate_id, event_type, event_id,
			payload, status, created_at, retry_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = tx.ExecContext(
		ctx, query,
		id,
		"product",
		p.ID().String(),
		event.EventType(),
		eventID,
		payloadJSON,
		"PENDING",
		time.Now(),
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}

	return nil
}

// domainEventToPayload converts domain event to payload map
func (w *ProductTxRepository) domainEventToPayload(event product.DomainEvent) map[string]interface{} {
	payload := map[string]interface{}{
		"occurred_at": event.OccurredAt().Format(time.RFC3339),
	}

	switch e := event.(type) {
	case product.ProductCreatedEvent:
		payload["product_id"] = e.ProductID.String()
		payload["seller_id"] = e.SellerID.String()

	case product.ProductPublishedEvent:
		payload["product_id"] = e.ProductID.String()
		payload["price"] = e.Money.Amount()
		payload["currency"] = e.Money.Currency()

	case product.ProductDeactivatedEvent:
		payload["product_id"] = e.ProductID.String()

	case product.ProductSoldOutEvent:
		payload["product_id"] = e.ProductID.String()
	}

	return payload
}
