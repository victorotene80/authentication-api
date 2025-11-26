package messaging

import (
	"authentication/internal/domain/events"
	"context"
)

type EventDispatcher interface {
	RegisterHandler(handler EventHandler)
	Dispatch(event events.DomainEvent) error
	DispatchWithContext(ctx context.Context, event events.DomainEvent) error
	DispatchAll(ctx context.Context, domainEvents []events.DomainEvent) error
}
