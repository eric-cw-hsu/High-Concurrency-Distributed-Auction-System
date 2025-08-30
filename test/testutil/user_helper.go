package testutil

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/samborkent/uuidv7"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
	walletUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/wallet"
)

// UserTestHelper provides utilities for creating test users and wallets
type UserTestHelper struct {
	db             *sql.DB
	walletService  walletUsecase.WalletService
	addFundUsecase *walletUsecase.AddFundUsecase
	ctx            context.Context
}

// NewUserTestHelper creates a new user test helper
func NewUserTestHelper(ctx context.Context, db *sql.DB, walletService walletUsecase.WalletService, addFundUsecase *walletUsecase.AddFundUsecase) *UserTestHelper {
	return &UserTestHelper{
		db:             db,
		walletService:  walletService,
		addFundUsecase: addFundUsecase,
		ctx:            ctx,
	}
}

// CreateTestUsersWithWallets creates test users with wallets and initial balance
func (h *UserTestHelper) CreateTestUsersWithWallets(count int, initialBalance float64) ([]string, error) {
	userIDs := make([]string, count)

	for i := 0; i < count; i++ {
		userID := uuidv7.New().String()
		userIDs[i] = userID

		// Create user in users table
		_, err := h.db.ExecContext(h.ctx,
			"INSERT INTO users (id, email, password_hash, name) VALUES ($1, $2, $3, $4)",
			userID, fmt.Sprintf("user%d@test.com", i), "test_hash", fmt.Sprintf("Test User %d", i))
		if err != nil {
			return nil, fmt.Errorf("failed to create user %s: %w", userID, err)
		}

		// Create wallet
		_, err = h.walletService.CreateWallet(h.ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to create wallet for user %s: %w", userID, err)
		}

		// Add initial balance if specified
		if initialBalance > 0 {
			addFundCmd := &wallet.AddFundCommand{
				UserId:      userID,
				Amount:      initialBalance,
				Description: "Initial test funding",
			}
			_, err = h.addFundUsecase.Execute(h.ctx, addFundCmd)
			if err != nil {
				return nil, fmt.Errorf("failed to add initial balance to wallet for user %s: %w", userID, err)
			}
		}
	}

	return userIDs, nil
}

// CreateTestUsersOnly creates test users without wallets
func (h *UserTestHelper) CreateTestUsersOnly(count int) ([]string, error) {
	userIDs := make([]string, count)

	for i := 0; i < count; i++ {
		userID := uuidv7.New().String()
		userIDs[i] = userID

		// Create user in users table only
		_, err := h.db.ExecContext(h.ctx,
			"INSERT INTO users (id, email, password_hash, name) VALUES ($1, $2, $3, $4)",
			userID, fmt.Sprintf("user%d@test.com", i), "test_hash", fmt.Sprintf("Test User %d", i))
		if err != nil {
			return nil, fmt.Errorf("failed to create user %s: %w", userID, err)
		}
	}

	return userIDs, nil
}

// CreateSingleTestUser creates a single test user without wallet
func (h *UserTestHelper) CreateSingleTestUser(email, name string) (string, error) {
	userID := uuidv7.New().String()

	_, err := h.db.ExecContext(h.ctx,
		"INSERT INTO users (id, email, password_hash, name) VALUES ($1, $2, $3, $4)",
		userID, email, "test_hash", name)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

// CleanupTestUsers removes all test users and their associated data
func (h *UserTestHelper) CleanupTestUsers(userIDs []string) error {
	for _, userID := range userIDs {
		// Delete wallet (cascade should handle related data)
		_, err := h.db.ExecContext(h.ctx, "DELETE FROM wallets WHERE user_id = $1", userID)
		if err != nil {
			return fmt.Errorf("failed to delete wallet for user %s: %w", userID, err)
		}

		// Delete user
		_, err = h.db.ExecContext(h.ctx, "DELETE FROM users WHERE id = $1", userID)
		if err != nil {
			return fmt.Errorf("failed to delete user %s: %w", userID, err)
		}
	}
	return nil
}
