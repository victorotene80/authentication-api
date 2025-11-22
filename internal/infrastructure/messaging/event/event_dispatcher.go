// internal/infrastructure/messaging/event/event_dispatcher.go
package event

import (
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/domain/events"
	"authentication/shared/logging"
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// CompositeEventDispatcher implements contracts.EventDispatcher
type CompositeEventDispatcher struct {
	handlers []messaging.EventHandler
	async    bool // Dispatch asynchronously
	mu       sync.RWMutex
}

// NewCompositeEventDispatcher creates a new multi-handler dispatcher
func NewCompositeEventDispatcher(async bool) messaging.EventDispatcher {
	return &CompositeEventDispatcher{
		handlers: make([] messaging.EventHandler, 0),
		async:    async,
	}
}

// RegisterHandler adds a new event handler
func (d *CompositeEventDispatcher) RegisterHandler(handler messaging.EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = append(d.handlers, handler)
}

// Dispatch sends event to all registered handlers (legacy method for compatibility)
func (d *CompositeEventDispatcher) Dispatch(event events.DomainEvent) error {
	return d.DispatchWithContext(context.Background(), event)
}

// DispatchWithContext dispatches with context support (implements interface)
func (d *CompositeEventDispatcher) DispatchWithContext(ctx context.Context, event events.DomainEvent) error {
	logging.InfoCtx(ctx, "Dispatching event",
		zap.String("event_name", event.EventName()),
		zap.String("event_id", event.EventID().String()),
		zap.String("aggregate_id", event.AggregateID().String()),
	)

	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.async {
		return d.dispatchAsync(ctx, event)
	}
	return d.dispatchSync(ctx, event)
}

// DispatchAll dispatches multiple events (implements interface)
func (d *CompositeEventDispatcher) DispatchAll(ctx context.Context, domainEvents []events.DomainEvent) error {
	for _, event := range domainEvents {
		if err := d.DispatchWithContext(ctx, event); err != nil {
			return fmt.Errorf("failed to dispatch event %s: %w", event.EventName(), err)
		}
	}
	return nil
}

func (d *CompositeEventDispatcher) dispatchSync(ctx context.Context, event events.DomainEvent) error {
	var errs []error
	eventName := event.EventName()

	for _, handler := range d.handlers {
		if handler.CanHandle(eventName) {
			if err := handler.Handle(ctx, event); err != nil {
				errs = append(errs, fmt.Errorf("handler error: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("dispatch errors: %v", errs)
	}
	return nil
}

func (d *CompositeEventDispatcher) dispatchAsync(ctx context.Context, event events.DomainEvent) error {
	eventName := event.EventName()
	var wg sync.WaitGroup

	for _, handler := range d.handlers {
		if handler.CanHandle(eventName) {
			wg.Add(1)
			go func(h messaging.EventHandler) {
				defer wg.Done()
				_ = h.Handle(ctx, event) // Log errors internally in handlers
			}(handler)
		}
	}

	wg.Wait()
	return nil
}