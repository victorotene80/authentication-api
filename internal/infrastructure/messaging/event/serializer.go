package event

import (
	"encoding/json"
	"fmt"

	"authentication/internal/domain/events"
)

// EventSerializer handles serialization of domain events to outbox format
type EventSerializer struct{}

func NewEventSerializer() *EventSerializer {
	return &EventSerializer{}
}

// SerializeEvent converts a domain event to JSON bytes for outbox storage
func (s *EventSerializer) SerializeEvent(event events.DomainEvent) ([]byte, error) {
	// Create envelope with full event metadata
	envelope := map[string]interface{}{
		"event_id":     event.EventID().String(),
		"event_name":   event.EventName(),
		"occurred_at":  event.OccurredAt().UTC().Format("2006-01-02T15:04:05.999Z"),
		"aggregate_id": event.AggregateID(),
		"payload":      event.Payload(),
		"metadata":     event.Metadata(),
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize event %s: %w", event.EventName(), err)
	}

	return data, nil
}

// DeserializeEvent reconstructs a BaseDomainEvent from JSON bytes
// (useful for background processors that need to republish)
func (s *EventSerializer) DeserializeEvent(data []byte) (*events.BaseDomainEvent, error) {
	var envelope struct {
		EventID     string            `json:"event_id"`
		EventName   string            `json:"event_name"`
		OccurredAt  string            `json:"occurred_at"`
		AggregateID string            `json:"aggregate_id"`
		Payload     json.RawMessage   `json:"payload"`
		Metadata    map[string]string `json:"metadata"`
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("failed to deserialize event: %w", err)
	}

	// Parse UUID and timestamp would go here if needed
	// For now, return basic structure
	return &events.BaseDomainEvent{
		Name:        envelope.EventName,
		AggregateId: envelope.AggregateID,
		Data:        envelope.Payload,
		Meta:        envelope.Metadata,
	}, nil
}
