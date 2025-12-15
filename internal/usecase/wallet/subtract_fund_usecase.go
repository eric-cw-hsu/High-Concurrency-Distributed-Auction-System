package wallet

import (
	"context"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/producer"
)

type SubtractFundUsecase struct {
	walletRepository    wallet.WalletRepository
	walletEventProducer producer.EventProducer
}

func NewSubtractFundUsecase(walletRepository wallet.WalletRepository, walletEventProducer producer.EventProducer) *SubtractFundUsecase {
	return &SubtractFundUsecase{
		walletRepository:    walletRepository,
		walletEventProducer: walletEventProducer,
	}
}

func (uc *SubtractFundUsecase) Execute(ctx context.Context, command *wallet.SubtractFundCommand) (*wallet.WalletAggregate, error) {
	// Get wallet aggregate
	aggregate, err := uc.walletRepository.GetByUserID(ctx, command.UserID)
	if err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "get_wallet",
			UserID:    command.UserID,
			Wrapped:   err,
		}
	}

	if aggregate == nil {
		return nil, &wallet.WalletNotFoundError{
			UserID: command.UserID,
		}
	}

	// Subtract fund from aggregate
	description := command.Description
	if description == "" {
		description = "Fund subtracted"
	}

	if err := aggregate.SubtractFund(command.Amount, description); err != nil {
		return nil, err // Domain errors are already properly typed
	}

	// Save the updated aggregate
	if err := uc.walletRepository.Save(ctx, aggregate); err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "save_wallet",
			UserID:    command.UserID,
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
