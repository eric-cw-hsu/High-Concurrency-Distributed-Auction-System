package message

import "context"

// Handler is a generic interface for handling any message envelope (DomainEvent, LogMessage, etc.)
type Handler interface {
	Handle(ctx context.Context, envelope MessageEnvelopeRaw) error
}
