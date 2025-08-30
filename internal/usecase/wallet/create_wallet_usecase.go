package wallet

import (
	"context"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
)

type CreateWalletUsecase struct {
	walletRepository wallet.WalletRepository
	eventPublisher   wallet.EventPublisher
}

func NewCreateWalletUsecase(walletRepository wallet.WalletRepository, eventPublisher wallet.EventPublisher) *CreateWalletUsecase {
	return &CreateWalletUsecase{
		walletRepository: walletRepository,
		eventPublisher:   eventPublisher,
	}
}

func (uc *CreateWalletUsecase) Execute(ctx context.Context, userId string) (*wallet.WalletAggregate, error) {
	// Check if wallet already exists
	existingAggregate, err := uc.walletRepository.GetByUserId(ctx, userId)
	if err == nil && existingAggregate != nil {
		return nil, &wallet.WalletAlreadyExistsError{
			UserId: userId,
		}
	}

	// Create new wallet aggregate
	aggregate := wallet.CreateNewWallet(userId)

	// Save the new aggregate
	if err := uc.walletRepository.Save(ctx, aggregate); err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "save_wallet",
			UserId:    userId,
			Wrapped:   err,
		}
	}

	// Publish events
	events := aggregate.GetEvents()
	for _, event := range events {
		if err := uc.eventPublisher.Publish(ctx, event); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish event: %v\n", err)
		}
	}

	// Return the aggregate
	return aggregate, nil
}
