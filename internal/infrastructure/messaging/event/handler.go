package event

import (
	"context"
	"fmt"
	"time"

	"authentication/internal/application/contracts/messaging"
	domainEvents "authentication/internal/domain/events"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/logging"
	"authentication/shared/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// EventHandler wraps event handlers with observability
type EventHandler struct {
	handler messaging.EventHandler
	logger  logging.Logger
	tracer  tracing.Tracer
	metrics *metrics.MetricsRecorder
}

// NewEventHandler creates an instrumented event handler
func NewEventHandler(
	handler messaging.EventHandler,
	logger logging.Logger,
	tracer tracing.Tracer,
	metricsRecorder *metrics.MetricsRecorder,
) messaging.EventHandler {
	return &EventHandler{
		handler: handler,
		logger:  logger.With(zap.String("component", "instrumented_event_handler")),
		tracer:  tracer,
		metrics: metricsRecorder,
	}
}

func (h *EventHandler) Handle(ctx context.Context, event domainEvents.DomainEvent) error {
	eventName := event.EventName()
	start := time.Now()

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
		h.tracer.AddAttributes(span, attribute.String("error.type", fmt.Sprintf("%T", err)))

		h.logger.Error(ctx, "event handling failed",
			zap.String("event_name", eventName),
			zap.String("event_id", event.EventID().String()),
			zap.Error(err),
			zap.Float64("duration_ms", duration*1000),
		)
	} else {
		span.SetStatus(codes.Ok, "success")
		h.logger.Debug(ctx, "event handled successfully",
			zap.String("event_name", eventName),
			zap.Float64("duration_ms", duration*1000),
		)
	}

	// Record metrics if recorder is provided
	if h.metrics != nil {
		h.metrics.RecordCommand(ctx, eventName, status, time.Duration(duration*float64(time.Second)))
	}

	return err
}

func (h *EventHandler) CanHandle(eventName string) bool {
	return h.handler.CanHandle(eventName)
}
