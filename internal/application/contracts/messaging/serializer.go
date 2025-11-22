package messaging

import(
    "authentication/internal/domain/events"
)

type EventSerializer interface {
    Serialize(event events.DomainEvent) ([]byte, error)
    Deserialize(data []byte, eventName string) (events.DomainEvent, error)
}