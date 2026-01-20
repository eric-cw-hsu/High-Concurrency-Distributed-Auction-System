package postgres

import (
	"context"
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/domain/order"
	"github.com/jmoiron/sqlx"
)

type RepositoryProvider interface {
	Orders() order.Repository
	Outbox() *OutboxRepository
}

// txProvider implements RepositoryProvider within a transaction
type txProvider struct {
	tx         *sqlx.Tx
	orderRepo  order.Repository
	outboxRepo *OutboxRepository
}

func (p *txProvider) Orders() order.Repository  { return p.orderRepo }
func (p *txProvider) Outbox() *OutboxRepository { return p.outboxRepo }

// TxManager coordinates database transactions and repository decoration
type TxManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

// Execute runs a function within a transaction block
func (m *TxManager) Execute(ctx context.Context, fn func(RepositoryProvider) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Decorate repositories with the transaction
	provider := &txProvider{
		tx:         tx,
		orderRepo:  NewOrderRepositoryWithTx(tx),
		outboxRepo: NewOutboxRepositoryWithTx(tx),
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	if err := fn(provider); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
