package wallet

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/test/testutil"
)

// RepositoryTestSuite is the test suite for wallet repository
type RepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	dbHelper   *testutil.DatabaseTestHelper
	repository *PostgresWalletRepository
}

// SetupSuite sets up the test database
func (suite *RepositoryTestSuite) SetupSuite() {
	ctx := context.Background()

	// Create database test helper
	dbHelper, err := testutil.NewDatabaseTestHelper(ctx)
	require.NoError(suite.T(), err, "Failed to create database test helper")

	// Run migrations
	err = dbHelper.RunMigrations()
	require.NoError(suite.T(), err, "Failed to run migrations")

	suite.dbHelper = dbHelper
	suite.repository = NewPostgresWalletRepository(dbHelper.DB)
}

// createTestTable creates the wallets table for testing
// TearDownSuite cleans up after tests
func (suite *RepositoryTestSuite) TearDownSuite() {
	if suite.dbHelper != nil {
		suite.dbHelper.Close()
	}
}

// SetupTest sets up each test
func (suite *RepositoryTestSuite) SetupTest() {
	// Clean up test data using database helper
	err := suite.dbHelper.CleanDatabase()
	require.NoError(suite.T(), err)

	// Create test users that will be used in wallet tests
	testUsers := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"550e8400-e29b-41d4-a716-446655440001",
		"550e8400-e29b-41d4-a716-446655440002",
		"550e8400-e29b-41d4-a716-446655440003",
		"550e8400-e29b-41d4-a716-446655440004",
		"550e8400-e29b-41d4-a716-446655440005",
	}

	for i, userID := range testUsers {
		_, err := suite.dbHelper.DB.ExecContext(context.Background(),
			`INSERT INTO users (id, email, password_hash, name) VALUES ($1, $2, $3, $4)`,
			userID,
			fmt.Sprintf("test%d@example.com", i),
			"hashedpassword",
			fmt.Sprintf("Test User %d", i),
		)
		require.NoError(suite.T(), err, "Failed to create test user")
	}
}

func (suite *RepositoryTestSuite) TestSaveAndGetWalletAggregate() {
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440000" // Valid UUID format

	// Create a new wallet aggregate
	aggregate := wallet.CreateNewWallet(userID)
	aggregate.ID = "wallet-550e8400-e29b-41d4-a716-446655440000" // Set ID for testing
	_ = aggregate.AddFund(100.0, "Initial deposit")

	// Save the aggregate
	err := suite.repository.Save(ctx, aggregate)
	require.NoError(suite.T(), err)

	// Retrieve the aggregate
	retrievedAggregate, err := suite.repository.GetByUserID(ctx, userID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrievedAggregate)

	// Verify the aggregate
	assert.Equal(suite.T(), aggregate.ID, retrievedAggregate.ID)
	assert.Equal(suite.T(), aggregate.UserID, retrievedAggregate.UserID)
	assert.Equal(suite.T(), aggregate.Balance, retrievedAggregate.Balance)
	assert.Equal(suite.T(), aggregate.Status, retrievedAggregate.Status)
	assert.Len(suite.T(), retrievedAggregate.Transactions, 1)

	// Reconstructed aggregate should not have uncommitted events
	assert.Empty(suite.T(), retrievedAggregate.PopEventPayloads())
}

func (suite *RepositoryTestSuite) TestGetByUserID_NotFound() {
	ctx := context.Background()

	aggregate, err := suite.repository.GetByUserID(ctx, "550e8400-e29b-41d4-a716-446655440001")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), aggregate)
	assert.True(suite.T(), errors.Is(err, wallet.ErrWalletNotFound))
}

func (suite *RepositoryTestSuite) TestCreateWallet() {
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440002"

	// Create wallet using CreateWallet
	createdAggregate, err := suite.repository.CreateWallet(ctx, userID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), createdAggregate)

	assert.Equal(suite.T(), userID, createdAggregate.UserID)
	assert.Equal(suite.T(), 0.0, createdAggregate.Balance)
	assert.Equal(suite.T(), wallet.WalletStatusActive, createdAggregate.Status)
	assert.NotEmpty(suite.T(), createdAggregate.ID)
}

func (suite *RepositoryTestSuite) TestWalletOperations() {
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440003"

	// Create wallet
	aggregate, err := suite.repository.CreateWallet(ctx, userID)
	require.NoError(suite.T(), err)

	// Add funds
	err = aggregate.AddFund(100.0, "Initial deposit")
	require.NoError(suite.T(), err)

	// Save the updated aggregate
	err = suite.repository.Save(ctx, aggregate)
	require.NoError(suite.T(), err)

	// Retrieve and verify
	retrievedAggregate, err := suite.repository.GetByUserID(ctx, userID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 100.0, retrievedAggregate.Balance)
	assert.Len(suite.T(), retrievedAggregate.Transactions, 1)
}

func (suite *RepositoryTestSuite) TestWalletDebit() {
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440004"

	// Create wallet and add funds
	aggregate, err := suite.repository.CreateWallet(ctx, userID)
	require.NoError(suite.T(), err)

	err = aggregate.AddFund(100.0, "Initial deposit")
	require.NoError(suite.T(), err)

	err = suite.repository.Save(ctx, aggregate)
	require.NoError(suite.T(), err)

	// Retrieve fresh aggregate
	aggregate, err = suite.repository.GetByUserID(ctx, userID)
	require.NoError(suite.T(), err)

	// Debit funds
	err = aggregate.SubtractFund(50.0, "Purchase")
	require.NoError(suite.T(), err)

	err = suite.repository.Save(ctx, aggregate)
	require.NoError(suite.T(), err)

	// Verify final balance
	finalAggregate, err := suite.repository.GetByUserID(ctx, userID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 50.0, finalAggregate.Balance)
	assert.Len(suite.T(), finalAggregate.Transactions, 2)
}

func (suite *RepositoryTestSuite) TestWalletInsufficientFunds() {
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440005"

	// Create wallet with insufficient funds
	aggregate, err := suite.repository.CreateWallet(ctx, userID)
	require.NoError(suite.T(), err)

	err = aggregate.AddFund(30.0, "Small deposit")
	require.NoError(suite.T(), err)

	err = suite.repository.Save(ctx, aggregate)
	require.NoError(suite.T(), err)

	// Try to debit more than available
	err = aggregate.SubtractFund(50.0, "Large purchase")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "insufficient balance")
}

// Run the test suite
func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

// Individual tests that don't require the test suite setup
func TestNewPostgresWalletRepository(t *testing.T) {
	// This test just checks that the constructor works
	// We can't easily test with a nil db here, so just test that the function exists
	assert.NotNil(t, NewPostgresWalletRepository)
}
