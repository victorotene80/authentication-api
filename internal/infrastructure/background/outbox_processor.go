package background

import (
	"authentication/internal/application/contracts"
	"authentication/internal/application/contracts/messaging"
	"authentication/shared/logging"
	"context"

	"go.uber.org/zap"
)

// OutboxProcessor handles processing of outbox events
type OutboxProcessor struct {
	outboxRepo contracts.OutboxRepository
	serializer messaging.EventSerializer
	logger     logging.Logger
	batchSize  int
	dispatcher messaging.EventDispatcher
}

func NewOutboxProcessor(
	outboxRepo contracts.OutboxRepository,
	serializer messaging.EventSerializer,
	logger logging.Logger,
	dispatcher messaging.EventDispatcher,
) *OutboxProcessor {
	return &OutboxProcessor{
		outboxRepo: outboxRepo,
		serializer: serializer,
		logger:     logger,
		dispatcher: dispatcher,
		batchSize:  100,
	}
}

// Process fetches and processes pending outbox events
func (p *OutboxProcessor) Process(ctx context.Context) error {
	events, err := p.outboxRepo.FetchPending(ctx, p.batchSize)
	if err != nil {
		p.logger.Error(ctx, "Failed to fetch pending outbox events", zap.Error(err))
		return err
	}

	if len(events) == 0 {
		p.logger.Debug(ctx, "No pending outbox events to process")
		return nil
	}

	p.logger.Info(ctx, "Processing outbox events", zap.Int("count", len(events)))

	var processed, failed int

	for _, outboxEvent := range events {
		if err := p.Process(ctx); err != nil {
			p.logger.Error(ctx, "Failed to process outbox event",
				zap.Error(err),
				zap.String("event_id", outboxEvent.EventId.String()),
				zap.String("event_name", outboxEvent.EventName),
			)

			if markErr := p.outboxRepo.MarkFailed(ctx, outboxEvent.EventId, err.Error()); markErr != nil {
				p.logger.Error(ctx, "Failed to mark event as failed", zap.Error(markErr))
			}
			failed++
		} else {
			if markErr := p.outboxRepo.MarkSent(ctx, outboxEvent.EventId); markErr != nil {
				p.logger.Error(ctx, "Failed to mark event as sent", zap.Error(markErr))
			}
			processed++
		}
	}

	p.logger.Info(ctx, "Outbox processing completed",
		zap.Int("processed", processed),
		zap.Int("failed", failed),
	)

	return nil
}

