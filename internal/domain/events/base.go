package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventID() uuid.UUID
	EventName() string
	OccurredAt() time.Time
	AggregateID() string
	Payload() json.RawMessage
	Metadata() map[string]string
}

type BaseDomainEvent struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Timestamp   time.Time         `json:"timestamp"`
	AggregateId string            `json:"aggregate_id"`
	Data        json.RawMessage   `json:"data"`
	Meta        map[string]string `json:"meta"`
}

func (e BaseDomainEvent) EventID() uuid.UUID          { return e.ID }
func (e BaseDomainEvent) EventName() string           { return e.Name }
func (e BaseDomainEvent) OccurredAt() time.Time       { return e.Timestamp }
func (e BaseDomainEvent) AggregateID() string         { return e.AggregateId }
func (e BaseDomainEvent) Payload() json.RawMessage    { return e.Data }
func (e BaseDomainEvent) Metadata() map[string]string { return e.Meta }

func newEvent(name string, aggregateID string, payload interface{}, meta map[string]string) DomainEvent {
	raw, _ := json.Marshal(payload)

	return BaseDomainEvent{
		ID:          uuid.New(),
		Name:        name,
		Timestamp:   time.Now().UTC(),
		AggregateId: aggregateID,
		Data:        raw,
		Meta:        meta,
	}
}
