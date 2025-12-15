package wallet

import (
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/message"
)

// Helper: Construct DomainEvent based on payload type (message package)
func buildWalletDomainEvent(payload interface{}) (*message.DomainEvent, bool) {
	switch p := payload.(type) {
	case *wallet.FundAddedPayload:
		return &message.DomainEvent{
			Name:        "wallet.fund_added",
			AggregateID: p.UserID,
			OccurredAt:  p.CreatedAt,
			Payload:     p,
			Version:     1,
		}, true
	case *wallet.WalletCreatedPayload:
		return &message.DomainEvent{
			Name:        "wallet.created",
			AggregateID: p.UserID,
			OccurredAt:  p.CreatedAt,
			Payload:     p,
			Version:     1,
		}, true
	case *wallet.FundSubtractedPayload:
		return &message.DomainEvent{
			Name:        "wallet.fund_subtracted",
			AggregateID: p.UserID,
			OccurredAt:  p.CreatedAt,
			Payload:     p,
			Version:     1,
		}, true
	case *wallet.RefundProcessedPayload:
		return &message.DomainEvent{
			Name:        "wallet.refund_processed",
			AggregateID: p.UserID,
			OccurredAt:  p.CreatedAt,
			Payload:     p,
			Version:     1,
		}, true
	case *wallet.WalletSuspendedPayload:
		return &message.DomainEvent{
			Name:        "wallet.suspended",
			AggregateID: p.UserID,
			OccurredAt:  p.SuspendedAt,
			Payload:     p,
			Version:     1,
		}, true
	case *wallet.WalletActivatedPayload:
		return &message.DomainEvent{
			Name:        "wallet.activated",
			AggregateID: p.UserID,
			OccurredAt:  p.ActivatedAt,
			Payload:     p,
			Version:     1,
		}, true
	default:
		return nil, false
	}
}
