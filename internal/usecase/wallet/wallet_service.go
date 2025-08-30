package wallet

import (
	"context"
	"errors"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
)

type WalletService interface {
	EnsureWalletExists(ctx context.Context, userId string) (*WalletInfo, error)
	CreateWallet(ctx context.Context, userId string) (*WalletInfo, error)
	ProcessPaymentWithSufficientFunds(ctx context.Context, userId, orderId string, amount float64) (*WalletInfo, error)
	ProcessRefundSafely(ctx context.Context, userId, orderId string, amount float64) (*WalletInfo, error)
}

type walletService struct {
	walletRepo     wallet.WalletRepository
	eventPublisher wallet.EventPublisher
}

// NewWalletService creates a new wallet service instance
func NewWalletService(walletRepo wallet.WalletRepository, eventPublisher wallet.EventPublisher) WalletService {
	return &walletService{
		walletRepo:     walletRepo,
		eventPublisher: eventPublisher,
	}
}

func (s *walletService) EnsureWalletExists(ctx context.Context, userId string) (*WalletInfo, error) {
	// Try to get existing wallet
	walletAgg, err := s.walletRepo.GetByUserId(ctx, userId)
	if err == nil {
		return s.aggregateToInfo(walletAgg), nil
	}

	// If wallet doesn't exist, create it
	if errors.Is(err, wallet.ErrWalletNotFound) {
		return s.CreateWallet(ctx, userId)
	}

	return nil, &wallet.RepositoryError{
		Operation: "check_wallet_existence",
		UserId:    userId,
		Wrapped:   err,
	}
}

func (s *walletService) CreateWallet(ctx context.Context, userId string) (*WalletInfo, error) {
	// Check if wallet already exists
	_, err := s.walletRepo.GetByUserId(ctx, userId)
	if err == nil {
		return nil, &wallet.WalletAlreadyExistsError{
			UserId: userId,
		}
	}

	if !errors.Is(err, wallet.ErrWalletNotFound) {
		return nil, &wallet.RepositoryError{
			Operation: "check_wallet_existence",
			UserId:    userId,
			Wrapped:   err,
		}
	}

	// Create new wallet aggregate (this is a business operation, so use CreateNewWallet)
	walletAgg := wallet.CreateNewWallet(userId)

	// Save to repository
	if err := s.walletRepo.Save(ctx, walletAgg); err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "save_wallet",
			UserId:    userId,
			Wrapped:   err,
		}
	}

	// Publish events
	for _, event := range walletAgg.GetUncommittedEvents() {
		if err := s.eventPublisher.Publish(ctx, event); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish event: %v\n", err)
		}
	}

	walletAgg.MarkEventsAsCommitted()

	return s.aggregateToInfo(walletAgg), nil
}

func (s *walletService) ProcessPaymentWithSufficientFunds(ctx context.Context, userId, orderId string, amount float64) (*WalletInfo, error) {
	// Get wallet
	walletAgg, err := s.walletRepo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "get_wallet",
			UserId:    userId,
			Wrapped:   err,
		}
	}

	// Process payment
	if err := walletAgg.ProcessPayment(amount, orderId); err != nil {
		return nil, err // Domain errors are already properly typed
	}

	// Save updated aggregate
	if err := s.walletRepo.Save(ctx, walletAgg); err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "save_wallet_after_payment",
			UserId:    userId,
			Wrapped:   err,
		}
	}

	// Publish events
	for _, event := range walletAgg.GetUncommittedEvents() {
		if err := s.eventPublisher.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish event: %v\n", err)
		}
	}

	walletAgg.MarkEventsAsCommitted()

	return s.aggregateToInfo(walletAgg), nil
}

func (s *walletService) ProcessRefundSafely(ctx context.Context, userId, orderId string, amount float64) (*WalletInfo, error) {
	// Get wallet
	walletAgg, err := s.walletRepo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "get_wallet",
			UserId:    userId,
			Wrapped:   err,
		}
	}

	// Process refund
	if err := walletAgg.ProcessRefund(amount, orderId); err != nil {
		return nil, err // Domain errors are already properly typed
	}

	// Save updated aggregate
	if err := s.walletRepo.Save(ctx, walletAgg); err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "save_wallet_after_refund",
			UserId:    userId,
			Wrapped:   err,
		}
	}

	// Publish events
	for _, event := range walletAgg.GetUncommittedEvents() {
		if err := s.eventPublisher.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish event: %v\n", err)
		}
	}

	walletAgg.MarkEventsAsCommitted()

	return s.aggregateToInfo(walletAgg), nil
}

func (s *walletService) aggregateToInfo(agg *wallet.WalletAggregate) *WalletInfo {
	return &WalletInfo{
		Balance:   agg.Balance,
		UpdatedAt: agg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
