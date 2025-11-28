package event

import (
	"authentication/internal/domain/events"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JSONEventSerializer struct{}

func NewJSONEventSerializer() *JSONEventSerializer {
	return &JSONEventSerializer{}
}

func (s *JSONEventSerializer) Serialize(event events.DomainEvent) ([]byte, error) {
	data := map[string]interface{}{
		"event_id":     event.EventID().String(),
		"event_name":   event.EventName(),
		"occurred_at":  event.OccurredAt().Unix(),
		"aggregate_id": event.AggregateID(),
		"payload":      event.Payload(),
	}
	return json.Marshal(data)
}

func (s *JSONEventSerializer) Deserialize(data []byte, eventName string) (events.DomainEvent, error) {
	var raw struct {
		EventID     string                 `json:"event_id"`
		EventName   string                 `json:"event_name"`
		OccurredAt  int64                  `json:"occurred_at"`
		AggregateID string                 `json:"aggregate_id"`
		Payload     map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	return events.BaseDomainEvent{
		ID:          uuid.MustParse(raw.EventID),
		Name:        raw.EventName,
		Timestamp:   time.Unix(raw.OccurredAt, 0),
		AggregateId: raw.AggregateID,
		Data:        raw.Payload,
	}, nil
}

