package wallet

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
)

type PostgresWalletRepository struct {
	db *sql.DB
}

func NewPostgresWalletRepository(db *sql.DB) *PostgresWalletRepository {
	return &PostgresWalletRepository{db: db}
}

// Save saves a wallet aggregate to the database
func (r *PostgresWalletRepository) Save(ctx context.Context, aggregate *wallet.WalletAggregate) error {
	// Start a transaction to ensure atomic save operation
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return WrapRepositoryError("begin_transaction", "", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO wallets (id, user_id, balance, status, created_at, updated_at, transactions) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			balance = EXCLUDED.balance,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at,
			transactions = EXCLUDED.transactions
	`

	// Serialize transactions to JSON
	transactionsJSON, err := json.Marshal(aggregate.Transactions)
	if err != nil {
		return WrapRepositoryError("marshal_transactions", aggregate.UserID, err)
	}

	// Generate ID if not set
	if aggregate.ID == "" {
		aggregate.ID = fmt.Sprintf("wallet_%s", aggregate.UserID)
	}

	_, err = tx.ExecContext(ctx, query,
		aggregate.ID,
		aggregate.UserID,
		aggregate.Balance,
		int(aggregate.Status),
		aggregate.CreatedAt,
		aggregate.UpdatedAt,
		transactionsJSON,
	)

	if err != nil {
		return WrapRepositoryError("save_wallet", aggregate.UserID, err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return WrapRepositoryError("commit_transaction", aggregate.UserID, err)
	}

	return nil
}

// GetByUserID retrieves a wallet aggregate by user ID
func (r *PostgresWalletRepository) GetByUserID(ctx context.Context, userID string) (*wallet.WalletAggregate, error) {
	query := `SELECT id, user_id, balance, status, created_at, updated_at, transactions FROM wallets WHERE user_id = $1`
	row := r.db.QueryRowContext(ctx, query, userID)

	var id, userIDDB string
	var balance float64
	var status int
	var createdAt, updatedAt time.Time
	var transactionsJSON []byte

	err := row.Scan(&id, &userIDDB, &balance, &status, &createdAt, &updatedAt, &transactionsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, WrapWalletNotFoundError(userID)
		}
		return nil, WrapRepositoryError("scan_wallet", userID, err)
	}

	// Deserialize transactions from JSON
	var transactions []wallet.Transaction
	if len(transactionsJSON) > 0 {
		err = json.Unmarshal(transactionsJSON, &transactions)
		if err != nil {
			return nil, WrapRepositoryError("unmarshal_transactions", userID, err)
		}
	}

	// Reconstruct aggregate from stored data (no events should be triggered)
	aggregate := wallet.ReconstructWalletAggregate(
		id, userIDDB, balance, wallet.WalletStatus(status),
		createdAt, updatedAt, transactions,
	)

	return aggregate, nil
}

// CreateWallet creates a new wallet aggregate for the given user
func (r *PostgresWalletRepository) CreateWallet(ctx context.Context, userID string) (*wallet.WalletAggregate, error) {
	// Create new wallet aggregate
	aggregate := wallet.CreateNewWallet(userID)

	// Save the aggregate
	err := r.Save(ctx, aggregate)
	if err != nil {
		return nil, err
	}

	return aggregate, nil
}
