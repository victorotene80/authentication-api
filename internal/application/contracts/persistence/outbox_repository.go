package persistence

import "context"

// OutboxMessage represents a generic message to be dispatched
type OutboxMessage struct {
	ID          string
	EventType   string
	AggregateID string
	Payload     []byte
	Metadata    map[string]string
	OccurredAt  int64
	SentAt      *int64
}


// OutboxRepository defines operations for persisting and fetching outbox messages
type OutboxRepository interface {
	Save(ctx context.Context, msg *OutboxMessage) error
	MarkAsSent(ctx context.Context, id string) error
	FetchPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
}
