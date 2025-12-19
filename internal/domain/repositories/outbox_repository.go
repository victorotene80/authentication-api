package repositories

import (
	"context"
	"database/sql"
	"time"
	"fmt"
	"authentication/internal/application/contracts/persistence"
)

type PostgresOutbox struct {
	db *sql.DB
}

func NewPostgresOutbox(db *sql.DB) *PostgresOutbox {
	return &PostgresOutbox{db: db}
}

// Save a new outbox message (must be called within a UoW transaction)
func (p *PostgresOutbox) Save(ctx context.Context, msg *persistence.OutboxMessage) error {
	if msg.CreatedAt == 0 {
		msg.CreatedAt = time.Now().UnixMilli()
	}
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO outbox (id, event_type, payload, created_at) VALUES ($1, $2, $3, to_timestamp($4 / 1000.0))`,
		msg.ID, msg.EventType, msg.Payload, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save outbox message: %w", err)
	}
	return nil
}

// MarkAsSent marks a message as successfully dispatched
func (p *PostgresOutbox) MarkAsSent(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	_, err := p.db.ExecContext(ctx,
		`UPDATE outbox SET sent_at = to_timestamp($1 / 1000.0) WHERE id = $2`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to mark outbox message as sent: %w", err)
	}
	return nil
}

// FetchPending returns unsent messages for background processing
func (p *PostgresOutbox) FetchPending(ctx context.Context, limit int) ([]*persistence.OutboxMessage, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, event_type, payload, extract(epoch from created_at)*1000 as created_at, 
		COALESCE(extract(epoch from sent_at)*1000, NULL) as sent_at 
		FROM outbox WHERE sent_at IS NULL ORDER BY created_at ASC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending outbox messages: %w", err)
	}
	defer rows.Close()

	var msgs []*persistence.OutboxMessage
	for rows.Next() {
		var m persistence.OutboxMessage
		var sentAt sql.NullFloat64
		if err := rows.Scan(&m.ID, &m.EventType, &m.Payload, &m.CreatedAt, &sentAt); err != nil {
			return nil, fmt.Errorf("failed to scan outbox row: %w", err)
		}
		if sentAt.Valid {
			v := int64(sentAt.Float64)
			m.SentAt = &v
		}
		msgs = append(msgs, &m)
	}

	return msgs, nil
}
