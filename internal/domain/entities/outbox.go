package entities

import (
	"time"
	"github.com/google/uuid"
)

type OutboxEvent struct {
    ID uuid.UUID
    AggregateID uuid.UUID
	EventId uuid.UUID
    EventName string
    Payload []byte
    Headers map[string]string
    CreatedAt time.Time
    Status string
    Attempts int
    LastError *string
    LastAttemptAt *time.Time
}