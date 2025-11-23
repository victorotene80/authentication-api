// internal/infrastructure/messaging/event/instrumented.go
package event

import (
	"authentication/shared/logging"
	"authentication/internal/application/contracts/messaging"
	"authentication/shared/tracing"
	domainEvents "authentication/internal/domain/events"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/utils"
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// InstrumentedEventHandler wraps event handlers with observability
type InstrumentedEventHandler struct {
	handler messaging.EventHandler
	logger  logging.Logger
	tracer  tracing.Tracer
}

func NewInstrumentedEventHandler(
	handler messaging.EventHandler,
	logger logging.Logger,
	tracer tracing.Tracer,
) messaging.EventHandler {
	return &InstrumentedEventHandler{
		handler: handler,
		logger:  logger.With(zap.String("component", "instrumented_event_handler")),
		tracer:  tracer,
	}
}

func (h *InstrumentedEventHandler) Handle(ctx context.Context, event domainEvents.DomainEvent) error {
	eventName := event.EventName()
	start := utils.NowUTC()

	ctx, span := h.tracer.StartSpan(ctx, fmt.Sprintf("event.handle.%s", eventName),
		trace.WithSpanKind(trace.SpanKindConsumer),
	)
	defer span.End()

	h.tracer.AddAttributes(span,
		attribute.String("event.name", eventName),
		attribute.String("event.id", event.EventID().String()),
	)

	h.logger.Debug(ctx, "handling event",
		zap.String("event_name", eventName),
		zap.String("event_id", event.EventID().String()),
	)

	err := h.handler.Handle(ctx, event)

	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
		span.SetStatus(codes.Error, err.Error())
		h.tracer.RecordError(span, err)
		h.logger.Error(ctx, "event handling failed",
			zap.String("event_name", eventName),
			zap.String("event_id", event.EventID().String()),
			zap.Error(err),
			zap.Float64("duration_seconds", duration),
		)
	} else {
		h.logger.Debug(ctx, "event handled successfully",
			zap.String("event_name", eventName),
			zap.Float64("duration_seconds", duration),
		)
	}

	// Record metrics only at the handler level to avoid duplication
	metrics.EventsProcessed.WithLabelValues(eventName, status).Inc()
	metrics.EventProcessingDuration.WithLabelValues(eventName).Observe(duration)

	return err
}

func (h *InstrumentedEventHandler) CanHandle(eventName string) bool {
	return h.handler.CanHandle(eventName)
}