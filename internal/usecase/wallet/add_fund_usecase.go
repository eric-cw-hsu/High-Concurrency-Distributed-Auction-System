package wallet

import (
	"context"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
)

type AddFundUsecase struct {
	walletRepository wallet.WalletRepository
	eventPublisher   wallet.EventPublisher
}

func NewAddFundUsecase(walletRepository wallet.WalletRepository, eventPublisher wallet.EventPublisher) *AddFundUsecase {
	return &AddFundUsecase{
		walletRepository: walletRepository,
		eventPublisher:   eventPublisher,
	}
}

func (uc *AddFundUsecase) Execute(ctx context.Context, command *wallet.AddFundCommand) (*wallet.WalletAggregate, error) {
	// Get wallet aggregate
	aggregate, err := uc.walletRepository.GetByUserId(ctx, command.UserId)
	if err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "get_wallet",
			UserId:    command.UserId,
			Wrapped:   err,
		}
	}

	if aggregate == nil {
		return nil, &wallet.WalletNotFoundError{
			UserId: command.UserId,
		}
	}

	// Add fund to aggregate
	description := command.Description
	if description == "" {
		description = "Fund added"
	}

	if err := aggregate.AddFund(command.Amount, description); err != nil {
		return nil, err // Domain errors are already properly typed
	}

	// Save the updated aggregate
	if err := uc.walletRepository.Save(ctx, aggregate); err != nil {
		return nil, &wallet.RepositoryError{
			Operation: "save_wallet",
			UserId:    command.UserId,
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
