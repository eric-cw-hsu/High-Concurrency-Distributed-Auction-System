package wallet

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/test/mocks"
)

func TestAddFundUsecase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockWalletRepository)
	mockPublisher := new(mocks.MockEventPublisher)
	usecase := NewAddFundUsecase(mockRepo, mockPublisher)

	userId := "user-123"
	amount := 100.0
	description := "Test deposit"

	command := &wallet.AddFundCommand{
		UserId:      userId,
		Amount:      amount,
		Description: description,
	}

	// Create mock wallet aggregate
	aggregate := wallet.CreateNewWallet(userId)
	aggregate.MarkEventsAsCommitted() // Clear creation events

	// Mock expectations
	mockRepo.On("GetByUserId", mock.Anything, userId).Return(aggregate, nil)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*wallet.WalletAggregate")).Return(nil)
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	// Act
	result, err := usecase.Execute(context.Background(), command)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userId, result.UserId)
	assert.Equal(t, amount, result.Balance)

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestAddFundUsecase_Execute_WalletNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockWalletRepository)
	mockPublisher := new(mocks.MockEventPublisher)
	usecase := NewAddFundUsecase(mockRepo, mockPublisher)

	userId := "non-existent-user"
	command := &wallet.AddFundCommand{
		UserId:      userId,
		Amount:      100.0,
		Description: "Test deposit",
	}

	// Mock expectations
	mockRepo.On("GetByUserId", mock.Anything, userId).Return(nil, wallet.ErrWalletNotFound)

	// Act
	result, err := usecase.Execute(context.Background(), command)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, wallet.ErrWalletNotFound))

	mockRepo.AssertExpectations(t)
}

func TestAddFundUsecase_Execute_InvalidAmount(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockWalletRepository)
	mockPublisher := new(mocks.MockEventPublisher)
	usecase := NewAddFundUsecase(mockRepo, mockPublisher)

	userId := "user-123"
	aggregate := wallet.CreateNewWallet(userId)

	command := &wallet.AddFundCommand{
		UserId:      userId,
		Amount:      -50.0, // Invalid amount
		Description: "Invalid deposit",
	}

	// Mock expectations
	mockRepo.On("GetByUserId", mock.Anything, userId).Return(aggregate, nil)

	// Act
	result, err := usecase.Execute(context.Background(), command)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, wallet.ErrInvalidAmount))

	mockRepo.AssertExpectations(t)
}

func TestAddFundUsecase_Execute_SaveFailure(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockWalletRepository)
	mockPublisher := new(mocks.MockEventPublisher)
	usecase := NewAddFundUsecase(mockRepo, mockPublisher)

	userId := "user-123"
	aggregate := wallet.CreateNewWallet(userId)
	aggregate.MarkEventsAsCommitted()

	command := &wallet.AddFundCommand{
		UserId:      userId,
		Amount:      100.0,
		Description: "Test deposit",
	}

	saveError := errors.New("database error")

	// Mock expectations
	mockRepo.On("GetByUserId", mock.Anything, userId).Return(aggregate, nil)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*wallet.WalletAggregate")).Return(saveError)

	// Act
	result, err := usecase.Execute(context.Background(), command)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	// Check if it's a repository error wrapping the save error
	var repoErr *wallet.RepositoryError
	assert.True(t, errors.As(err, &repoErr))
	assert.Equal(t, saveError, repoErr.Wrapped)

	mockRepo.AssertExpectations(t)
}

func TestCreateWalletUsecase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockWalletRepository)
	mockPublisher := new(mocks.MockEventPublisher)
	usecase := NewCreateWalletUsecase(mockRepo, mockPublisher)

	userId := "new-user-123"

	// Mock expectations
	mockRepo.On("GetByUserId", mock.Anything, userId).Return(nil, wallet.ErrWalletNotFound)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*wallet.WalletAggregate")).Return(nil)
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	// Act
	result, err := usecase.Execute(context.Background(), userId)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userId, result.UserId)
	assert.Equal(t, 0.0, result.Balance)

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestCreateWalletUsecase_Execute_WalletAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockWalletRepository)
	mockPublisher := new(mocks.MockEventPublisher)
	usecase := NewCreateWalletUsecase(mockRepo, mockPublisher)

	userId := "existing-user-123"
	existingAggregate := wallet.CreateNewWallet(userId)

	// Mock expectations
	mockRepo.On("GetByUserId", mock.Anything, userId).Return(existingAggregate, nil)

	// Act
	result, err := usecase.Execute(context.Background(), userId)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, wallet.ErrWalletAlreadyExists))

	mockRepo.AssertExpectations(t)
}

func TestSubtractFundUsecase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockWalletRepository)
	mockPublisher := new(mocks.MockEventPublisher)
	usecase := NewSubtractFundUsecase(mockRepo, mockPublisher)

	userId := "user-123"
	aggregate := wallet.CreateNewWallet(userId)
	_ = aggregate.AddFund(100.0, "Initial deposit")
	aggregate.MarkEventsAsCommitted()

	command := &wallet.SubtractFundCommand{
		UserId:      userId,
		Amount:      50.0,
		Description: "Test withdrawal",
	}

	// Mock expectations
	mockRepo.On("GetByUserId", mock.Anything, userId).Return(aggregate, nil)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*wallet.WalletAggregate")).Return(nil)
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	// Act
	result, err := usecase.Execute(context.Background(), command)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userId, result.UserId)
	assert.Equal(t, 50.0, result.Balance) // 100 - 50

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestSubtractFundUsecase_Execute_InsufficientFunds(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockWalletRepository)
	mockPublisher := new(mocks.MockEventPublisher)
	usecase := NewSubtractFundUsecase(mockRepo, mockPublisher)

	userId := "user-123"
	aggregate := wallet.CreateNewWallet(userId)
	_ = aggregate.AddFund(30.0, "Small deposit")
	aggregate.MarkEventsAsCommitted()

	command := &wallet.SubtractFundCommand{
		UserId:      userId,
		Amount:      50.0, // More than available
		Description: "Test withdrawal",
	}

	// Mock expectations
	mockRepo.On("GetByUserId", mock.Anything, userId).Return(aggregate, nil)

	// Act
	result, err := usecase.Execute(context.Background(), command)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, wallet.ErrInsufficientBalance))

	mockRepo.AssertExpectations(t)
}

func TestWalletService_ProcessPaymentWithSufficientFunds_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockWalletRepository)
	mockPublisher := new(mocks.MockEventPublisher)
	service := NewWalletService(mockRepo, mockPublisher)

	userId := "user-123"
	orderId := "order-456"
	amount := 50.0

	aggregate := wallet.CreateNewWallet(userId)
	_ = aggregate.AddFund(100.0, "Initial deposit")
	aggregate.MarkEventsAsCommitted()

	// Mock expectations
	mockRepo.On("GetByUserId", mock.Anything, userId).Return(aggregate, nil)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*wallet.WalletAggregate")).Return(nil)
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	// Act
	result, err := service.ProcessPaymentWithSufficientFunds(context.Background(), userId, orderId, amount)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 50.0, result.Balance) // Updated balance

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestWalletService_ProcessRefundSafely_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockWalletRepository)
	mockPublisher := new(mocks.MockEventPublisher)
	service := NewWalletService(mockRepo, mockPublisher)

	userId := "user-123"
	orderId := "order-456"
	refundAmount := 30.0

	aggregate := wallet.CreateNewWallet(userId)
	_ = aggregate.AddFund(100.0, "Initial deposit")
	_ = aggregate.ProcessPayment(50.0, orderId)
	aggregate.MarkEventsAsCommitted()

	// Mock expectations
	mockRepo.On("GetByUserId", mock.Anything, userId).Return(aggregate, nil)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*wallet.WalletAggregate")).Return(nil)
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	// Act
	result, err := service.ProcessRefundSafely(context.Background(), userId, orderId, refundAmount)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 80.0, result.Balance) // 50 + 30 refund

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}
