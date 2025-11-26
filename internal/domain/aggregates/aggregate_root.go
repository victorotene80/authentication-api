package aggregates

import (
	"authentication/internal/domain/events"
	"authentication/shared/utils"
	"time"
)

type AggregateRoot struct {
	id        string
	version   int
	events    []events.DomainEvent
	createdAt time.Time
	updatedAt time.Time
}

func NewAggregateRoot(id string) *AggregateRoot {
	now := utils.NowUTC()
	return &AggregateRoot{
		id:        id,
		version:   1,
		events:    make([]events.DomainEvent, 0),
		createdAt: now,
		updatedAt: now,
	}
}

func (a *AggregateRoot) ID() string {
	return a.id
}

func (a *AggregateRoot) Version() int {
	return a.version
}

func (a *AggregateRoot) GetEvents() []events.DomainEvent {
	return a.events
}

func (a *AggregateRoot) IncrementVersion() {
	a.version++
	a.updatedAt = time.Now()
}

func (a *AggregateRoot) AddEvent(event events.DomainEvent) {
    a.events = append(a.events, event)
}

// DomainEvents returns a copy of all events (immutable)
func (a *AggregateRoot) DomainEvents() []events.DomainEvent {
	eventsCopy := make([]events.DomainEvent, len(a.events))
	copy(eventsCopy, a.events)
	return eventsCopy
}

func (a *AggregateRoot) ClearEvents() {
	a.events = make([]events.DomainEvent, 0)
}
