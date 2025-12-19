package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
		`INSERT INTO outbox (id, event_type, payload, created_at) 
		 VALUES ($1, $2, $3, to_timestamp($4 / 1000.0))`,
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
	
	result, err := p.db.ExecContext(ctx,
		`UPDATE outbox SET sent_at = to_timestamp($1 / 1000.0) WHERE id = $2 AND sent_at IS NULL`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to mark outbox message as sent: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("outbox message %s not found or already sent", id)
	}

	return nil
}

// FetchPending returns unsent messages for background processing
// IMPROVED: Use int64 directly to avoid float precision issues
func (p *PostgresOutbox) FetchPending(ctx context.Context, limit int) ([]*persistence.OutboxMessage, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, event_type, payload, 
		        EXTRACT(EPOCH FROM created_at)::bigint * 1000 as created_at,
		        EXTRACT(EPOCH FROM sent_at)::bigint * 1000 as sent_at
		 FROM outbox 
		 WHERE sent_at IS NULL 
		 ORDER BY created_at ASC 
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending outbox messages: %w", err)
	}
	defer rows.Close()

	var msgs []*persistence.OutboxMessage
	for rows.Next() {
		var m persistence.OutboxMessage
		var sentAt sql.NullInt64

		if err := rows.Scan(&m.ID, &m.EventType, &m.Payload, &m.CreatedAt, &sentAt); err != nil {
			return nil, fmt.Errorf("failed to scan outbox row: %w", err)
		}

		if sentAt.Valid {
			m.SentAt = &sentAt.Int64
		}

		msgs = append(msgs, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating outbox rows: %w", err)
	}

	return msgs, nil
}

// Delete removes old sent messages (for cleanup jobs)
func (p *PostgresOutbox) DeleteOldSent(ctx context.Context, olderThanDays int) (int64, error) {
	result, err := p.db.ExecContext(ctx,
		`DELETE FROM outbox 
		 WHERE sent_at IS NOT NULL 
		 AND sent_at < NOW() - INTERVAL '1 day' * $1`,
		olderThanDays,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old outbox messages: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get deleted count: %w", err)
	}

	return count, nil
}