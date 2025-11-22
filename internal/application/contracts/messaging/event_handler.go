package messaging

import (
	"authentication/internal/domain/events"
	"context"
)

type EventHandler interface {
	Handle(ctx context.Context, event events.DomainEvent) error
	CanHandle(eventName string) bool
}