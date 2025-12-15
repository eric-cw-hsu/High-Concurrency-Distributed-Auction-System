package wallet

import (
	"context"
	"errors"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/producer"
)

type WalletService interface {
	EnsureWalletExists(ctx context.Context, userID string) (*WalletInfo, error)
	CreateWallet(ctx context.Context, userID string) (*WalletInfo, error)
	ProcessPaymentWithSufficientFunds(ctx context.Context, userID, orderId string, amount float64) (*WalletInfo, error)
	ProcessRefundSafely(ctx context.Context, userID, orderId string, amount float64) (*WalletInfo, error)
}

type walletService struct {
	walletRepo          wallet.WalletRepository
	walletEventProducer producer.EventProducer
}

// NewWalletService creates a new wallet service instance
func NewWalletService(walletRepo wallet.WalletRepository, walletEventProducer producer.EventProducer) WalletService {
	return &walletService{
		walletRepo:          walletRepo,
		walletEventProducer: walletEventProducer,
	}
}

func (s *walletService) EnsureWalletExists(ctx context.Context, userID string) (*WalletInfo, error) {
	// Try to get existing wallet
	walletAgg, err := s.walletRepo.GetByUserID(ctx, userID)
	if err == nil {
		return s.aggregateToInfo(walletAgg), nil
	}

	// If wallet doesn't exist, create it
	if errors.Is(err, wallet.ErrWalletNotFound) {
		return s.CreateWallet(ctx, userID)
	}

	return nil, &wallet.RepositoryError{
		Operation: "check_wallet_existence",
		UserID:    userID,
		Wrapped:   err,
	}
}

func (s *walletService) CreateWallet(ctx context.Context, userID string) (*WalletInfo, error) {
	// Check if wallet already exists
	_, err := s.walletRepo.GetByUserID(ctx, userID)
	if err == nil {
		return nil, &wallet.WalletAlreadyExistsError{
			UserID: userID,
		}
	}

	if !errors.Is(err, wallet.ErrWalletNotFound) {
		return nil, &wallet.RepositoryError{
			Operation: "check_wallet_existence",
			UserID:    userID,
			Wrapped:   err,
		}
	}

	// Create new wallet aggregate (this is a business operation, so use CreateNewWallet)
	walletAgg := wallet.CreateNewWallet(userID)

	// Save to repository
	if err := s.walletRepo.Save(ctx, walletAgg); err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "save_wallet",
			UserID:    userID,
			Wrapped:   err,
		}
	}

	// Publish events
	for eventPayload := range walletAgg.PopEventPayloads() {
		domainEvent, ok := buildWalletDomainEvent(eventPayload)
		if !ok {
			continue
		}

		if err := s.walletEventProducer.PublishEvent(ctx, domainEvent); err != nil {
			fmt.Printf("Failed to publish event: %v\n", err)
		}
	}

	return s.aggregateToInfo(walletAgg), nil
}

func (s *walletService) ProcessPaymentWithSufficientFunds(ctx context.Context, userID, orderId string, amount float64) (*WalletInfo, error) {
	// Get wallet
	walletAgg, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "get_wallet",
			UserID:    userID,
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
			UserID:    userID,
			Wrapped:   err,
		}
	}

	// Publish events
	for eventPayload := range walletAgg.PopEventPayloads() {
		domainEvent, ok := buildWalletDomainEvent(eventPayload)
		if !ok {
			continue
		}

		if err := s.walletEventProducer.PublishEvent(ctx, domainEvent); err != nil {
			fmt.Printf("Failed to publish event: %v\n", err)
		}
	}

	return s.aggregateToInfo(walletAgg), nil
}

func (s *walletService) ProcessRefundSafely(ctx context.Context, userID, orderId string, amount float64) (*WalletInfo, error) {
	// Get wallet
	walletAgg, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "get_wallet",
			UserID:    userID,
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
			UserID:    userID,
			Wrapped:   err,
		}
	}

	// Publish events
	for eventPayload := range walletAgg.PopEventPayloads() {
		domainEvent, ok := buildWalletDomainEvent(eventPayload)
		if !ok {
			continue
		}

		if err := s.walletEventProducer.PublishEvent(ctx, domainEvent); err != nil {
			fmt.Printf("Failed to publish event: %v\n", err)
		}
	}

	return s.aggregateToInfo(walletAgg), nil
}

func (s *walletService) aggregateToInfo(agg *wallet.WalletAggregate) *WalletInfo {
	return &WalletInfo{
		Balance:   agg.Balance,
		UpdatedAt: agg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
