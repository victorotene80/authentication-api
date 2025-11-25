package aggregates

import "authentication/internal/domain/events"

// AggregateRoot provides reusable event management
type AggregateRoot struct {
	version int64 // Add this for optimistic locking
	events  []events.DomainEvent
}

// AddEvent registers a new event
func (a *AggregateRoot) AddEvent(event events.DomainEvent) {
	a.events = append(a.events, event)
}

// DomainEvents returns a copy of all events (immutable)
func (a *AggregateRoot) DomainEvents() []events.DomainEvent {
	eventsCopy := make([]events.DomainEvent, len(a.events))
	copy(eventsCopy, a.events)
	return eventsCopy
}

// ClearEvents clears all registered events
func (a *AggregateRoot) ClearEvents() {
	a.events = []events.DomainEvent{}
}

func (a *AggregateRoot) Version() int64    { return a.version }
func (a *AggregateRoot) IncrementVersion() { a.version++ }
