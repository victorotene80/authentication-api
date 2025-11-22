package repositories

import (
	"authentication/internal/domain/entities"
	"authentication/internal/application/contracts"
	uow "authentication/shared/persistence"
	"context"
	"time"

	"github.com/google/uuid"
)

type postgresOutboxRepository struct {
	db uow.DB
}

func NewPostgresOutboxRepository(db uow.DB) contracts.OutboxRepository {
	return &postgresOutboxRepository{db: db}
}

func (r *postgresOutboxRepository) SetDB(db uow.DB) {
	r.db = db
}

func (r *postgresOutboxRepository) InsertTx(ctx context.Context, tx uow.DB, event entities.OutboxEvent) error {
	query := `
		INSERT INTO outbox_events (
			id, event_id, event_name, aggregate_id, payload, created_at, status, attempts
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		event.ID,
		event.EventId,
		event.EventName,
		event.AggregateID,
		event.Payload,
		event.CreatedAt,
		"pending",
		0,
	)
	return err
}

func (r *postgresOutboxRepository) FetchPending(ctx context.Context, limit int) ([]entities.OutboxEvent, error) {
	query := `
		SELECT id, event_id, event_name, aggregate_id, payload, created_at, status, attempts, last_error, last_attempt_at
		FROM outbox_events
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []entities.OutboxEvent
	for rows.Next() {
		var e entities.OutboxEvent
		err := rows.Scan(
			&e.ID,
			&e.EventId,
			&e.EventName,
			&e.AggregateID,
			&e.Payload,
			&e.CreatedAt,
			&e.Status,
			&e.Attempts,
			&e.LastError,
			&e.LastAttemptAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *postgresOutboxRepository) MarkSent(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE outbox_events
		SET status = 'sent', last_attempt_at = $2
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	return err
}

func (r *postgresOutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `
		UPDATE outbox_events
		SET status = 'failed', attempts = attempts + 1, last_error = $2, last_attempt_at = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, errMsg, time.Now())
	return err
}

func (r *postgresOutboxRepository) DeleteProcessed(ctx context.Context, olderThanDays int) error {
	query := `
		DELETE FROM outbox_events
		WHERE status IN ('sent', 'failed')
		AND last_attempt_at < NOW() - INTERVAL '1 day' * $1
	`
	_, err := r.db.ExecContext(ctx, query, olderThanDays)
	return err
}
