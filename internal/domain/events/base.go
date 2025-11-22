package events

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventID() uuid.UUID
	EventName() string
	OccurredAt() time.Time
	AggregateID() uuid.UUID
	Payload() map[string]interface{}
}

type BaseDomainEvent struct {
	ID          uuid.UUID
	Name        string
	Timestamp   time.Time
	AggregateId uuid.UUID
	Data        map[string]interface{}
}

func (e BaseDomainEvent) EventID() uuid.UUID     { return e.ID }
func (e BaseDomainEvent) EventName() string      { return e.Name }
func (e BaseDomainEvent) OccurredAt() time.Time  { return e.Timestamp }
func (e BaseDomainEvent) AggregateID() uuid.UUID { return e.AggregateId }
func (e BaseDomainEvent) Payload() map[string]interface{} { return e.Data }
