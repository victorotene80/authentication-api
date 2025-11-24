package event

import (
	"context"
	"fmt"
	"time"

	"authentication/internal/application/contracts/messaging"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/logging"
	"authentication/shared/tracing"
	"authentication/shared/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// EventBus decorates a clean EventBus with logging, tracing, and metrics.
type EventBus struct {
	bus     messaging.EventBus
	logger  logging.Logger
	tracer  tracing.Tracer
	metrics *metrics.MetricsRecorder
}

// NewEventBus creates a new instrumented event bus.
func NewEventBus(
	bus messaging.EventBus,
	logger logging.Logger,
	tracer tracing.Tracer,
	metricsRecorder *metrics.MetricsRecorder,
) messaging.EventBus {
	return &EventBus{
		bus:     bus,
		logger:  logger.With(zap.String("component", "instrumented_event_bus")),
		tracer:  tracer,
		metrics: metricsRecorder,
	}
}

// Publish wraps the clean bus Publish with observability.
func (b *EventBus) Publish(event messaging.Event) error {
	eventName := event.Name
	start := utils.NowUTC()

	ctx := context.Background()
	ctx, span := b.tracer.StartSpan(ctx, fmt.Sprintf("event.publish.%s", eventName),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	b.tracer.AddAttributes(span,
		attribute.String("event.name", eventName),
		attribute.Int("payload.size", len(event.Payload)),
	)

	b.logger.Debug(ctx, "publishing event",
		zap.String("event_name", eventName),
		zap.Int("payload_size", len(event.Payload)),
	)

	err := b.bus.Publish(event)
	duration := time.Since(start)
	status := "success"

	if err != nil {
		status = "error"
		span.SetStatus(codes.Error, err.Error())
		b.tracer.RecordError(span, err)
		b.logger.Error(ctx, "event publishing failed",
			zap.String("event_name", eventName),
			zap.Error(err),
			zap.Float64("duration_seconds", duration.Seconds()),
		)
	} else {
		b.logger.Debug(ctx, "event published successfully",
			zap.String("event_name", eventName),
			zap.Float64("duration_seconds", duration.Seconds()),
		)
	}

	// Record metrics if recorder is provided
	if b.metrics != nil {
		b.metrics.RecordCommand(ctx, eventName, status, duration)
	}

	return err
}

// Subscribe wraps the clean bus Subscribe with tracing and logging.
func (b *EventBus) Subscribe(topic string, handler func(messaging.Event)) error {
	ctx := context.Background()
	ctx, span := b.tracer.StartSpan(ctx, fmt.Sprintf("event.subscribe.%s", topic))
	defer span.End()

	b.tracer.AddAttributes(span, attribute.String("topic", topic))
	b.logger.Info(ctx, "subscribing to topic", zap.String("topic", topic))

	// Wrap handler to inject logging and observability
	wrappedHandler := func(e messaging.Event) {
		handlerCtx := context.Background()
		b.logger.Debug(handlerCtx, "received event from subscription",
			zap.String("topic", topic),
			zap.String("event_name", e.Name),
		)
		handler(e)
	}

	err := b.bus.Subscribe(topic, wrappedHandler)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		b.tracer.RecordError(span, err)
		b.logger.Error(ctx, "subscription failed",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return err
	}

	b.logger.Info(ctx, "successfully subscribed to topic", zap.String("topic", topic))
	return nil
}

// Close wraps the clean bus Close with tracing and logging.
func (b *EventBus) Close() error {
	ctx := context.Background()
	ctx, span := b.tracer.StartSpan(ctx, "event.bus.close")
	defer span.End()

	b.logger.Info(ctx, "closing event bus")

	err := b.bus.Close()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		b.tracer.RecordError(span, err)
		b.logger.Error(ctx, "error closing event bus", zap.Error(err))
		return err
	}

	b.logger.Info(ctx, "event bus closed successfully")
	return nil
}
