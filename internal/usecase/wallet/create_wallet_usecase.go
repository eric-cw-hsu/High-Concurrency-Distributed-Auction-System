package wallet

import (
	"context"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/producer"
)

type CreateWalletUsecase struct {
	walletRepository    wallet.WalletRepository
	walletEventProducer producer.EventProducer
}

func NewCreateWalletUsecase(walletRepository wallet.WalletRepository, walletEventProducer producer.EventProducer) *CreateWalletUsecase {
	return &CreateWalletUsecase{
		walletRepository:    walletRepository,
		walletEventProducer: walletEventProducer,
	}
}

func (uc *CreateWalletUsecase) Execute(ctx context.Context, userID string) (*wallet.WalletAggregate, error) {
	// Check if wallet already exists
	existingAggregate, err := uc.walletRepository.GetByUserID(ctx, userID)
	if err == nil && existingAggregate != nil {
		return nil, &wallet.WalletAlreadyExistsError{
			UserID: userID,
		}
	}

	// Create new wallet aggregate
	aggregate := wallet.CreateNewWallet(userID)

	// Save the new aggregate
	if err := uc.walletRepository.Save(ctx, aggregate); err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "save_wallet",
			UserID:    userID,
			Wrapped:   err,
		}
	}

	// Publish events
	for eventPayload := range aggregate.PopEventPayloads() {
		domainEvent, ok := buildWalletDomainEvent(eventPayload)
		if !ok {
			continue
		}

		if err := uc.walletEventProducer.PublishEvent(ctx, domainEvent); err != nil {
			fmt.Printf("Failed to publish event: %v\n", err)
		}
	}

	// Return the aggregate
	return aggregate, nil
}
