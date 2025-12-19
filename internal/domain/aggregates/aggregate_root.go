package aggregates

import (
	"authentication/internal/domain/events"
	"authentication/shared/utils"
	"time"
)

type AggregateRoot struct {
	id                string
	version           int
	events            []events.DomainEvent
	createdAt         time.Time
	updatedAt         time.Time
	//uncommittedEvents []events.DomainEvent
}

func NewAggregateRoot(id string) *AggregateRoot {
	now := utils.NowUTC()
	return &AggregateRoot{
		id:                id,
		version:           1,
		events:            make([]events.DomainEvent, 0),
		createdAt:         now,
		updatedAt:         now,
		//uncommittedEvents: make([]events.DomainEvent, 0),
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

/*
// GetUncommittedEvents returns all events that haven't been persisted yet
func (a *AggregateRoot) GetUncommittedEvents() []events.DomainEvent {
	return a.uncommittedEvents
}

// ClearUncommittedEvents clears the event list after successful persistence
func (a *AggregateRoot) ClearUncommittedEvents() {
	a.uncommittedEvents = make([]events.DomainEvent, 0)
}

// HasUncommittedEvents checks if there are any unpersisted events
func (a *AggregateRoot) HasUncommittedEvents() bool {
	return len(a.uncommittedEvents) > 0
}*/
