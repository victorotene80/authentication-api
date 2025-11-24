package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"authentication/internal/application/contracts/messaging"
	"authentication/internal/domain/events"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/logging"
	"authentication/shared/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// CompositeEventDispatcher implements messaging.EventDispatcher with observability.
type CompositeEventDispatcher struct {
	handlers []messaging.EventHandler
	async    bool
	mu       sync.RWMutex
	logger   logging.Logger
	tracer   tracing.Tracer
	metrics  *metrics.MetricsRecorder
}

// NewCompositeEventDispatcher creates a new dispatcher.
func NewCompositeEventDispatcher(
	async bool,
	logger logging.Logger,
	tracer tracing.Tracer,
	metricsRecorder *metrics.MetricsRecorder,
) messaging.EventDispatcher {
	return &CompositeEventDispatcher{
		handlers: make([]messaging.EventHandler, 0),
		async:    async,
		logger:   logger.With(zap.String("component", "event_dispatcher")),
		tracer:   tracer,
		metrics:  metricsRecorder,
	}
}

// RegisterHandler adds a new event handler.
func (d *CompositeEventDispatcher) RegisterHandler(handler messaging.EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers = append(d.handlers, handler)

	d.logger.Debug(context.Background(), "Event handler registered",
		zap.Int("total_handlers", len(d.handlers)),
	)
}

// Dispatch sends event to all handlers (legacy method).
func (d *CompositeEventDispatcher) Dispatch(event events.DomainEvent) error {
	return d.DispatchWithContext(context.Background(), event)
}

// DispatchWithContext dispatches event with context, tracing, logging, and metrics.
func (d *CompositeEventDispatcher) DispatchWithContext(ctx context.Context, event events.DomainEvent) error {
	eventName := event.EventName()
	ctx, span := d.tracer.StartSpan(ctx, "event.dispatch."+eventName,
		trace.WithSpanKind(trace.SpanKindProducer),
	)
	defer span.End()

	d.tracer.AddAttributes(span,
		attribute.String("event.name", eventName),
		attribute.String("event.id", event.EventID().String()),
		attribute.String("aggregate.id", event.AggregateID().String()),
		attribute.Bool("async", d.async),
	)

	d.logger.Info(ctx, "Dispatching event",
		zap.String("event_name", eventName),
		zap.String("event_id", event.EventID().String()),
		zap.String("aggregate_id", event.AggregateID().String()),
		zap.Bool("async", d.async),
	)

	start := time.Now()

	d.mu.RLock()
	handlers := make([]messaging.EventHandler, len(d.handlers))
	copy(handlers, d.handlers)
	d.mu.RUnlock()

	// Count handlers
	var handlerCount int
	for _, handler := range handlers {
		if handler.CanHandle(eventName) {
			handlerCount++
		}
	}
	d.tracer.AddAttributes(span, attribute.Int("handlers.count", handlerCount))

	// Dispatch
	var err error
	if d.async {
		err = d.dispatchAsync(ctx, event, handlers)
	} else {
		err = d.dispatchSync(ctx, event, handlers)
	}

	duration := time.Since(start)
	status := "success"
	if err != nil {
		status = "error"
		d.logger.Error(ctx, "Event dispatch failed",
			zap.Error(err),
			zap.String("event_name", eventName),
			zap.Duration("duration", duration),
		)
		d.tracer.RecordError(span, err)
		span.SetStatus(codes.Error, "dispatch failed")
	} else {
		d.logger.Debug(ctx, "Event dispatched successfully",
			zap.String("event_name", eventName),
			zap.Int("handlers_executed", handlerCount),
			zap.Duration("duration", duration),
		)
		span.SetStatus(codes.Ok, "success")
	}

	if d.metrics != nil {
		d.metrics.RecordCommand(ctx, eventName, status, duration)
	}

	d.tracer.AddEvent(span, "event.dispatched", attribute.Int("handlers_count", handlerCount))
	d.tracer.AddAttributes(span, attribute.Float64("duration_ms", duration.Seconds()*1000), attribute.String("status", status))

	return err
}

// DispatchAll dispatches multiple events.
func (d *CompositeEventDispatcher) DispatchAll(ctx context.Context, events []events.DomainEvent) error {
	ctx, span := d.tracer.StartSpan(ctx, "event.dispatch_all")
	defer span.End()

	d.tracer.AddAttributes(span, attribute.Int("events.count", len(events)))
	d.logger.Info(ctx, "Dispatching multiple events", zap.Int("count", len(events)))

	for i, e := range events {
		if err := d.DispatchWithContext(ctx, e); err != nil {
			d.logger.Error(ctx, "Failed to dispatch event in batch",
				zap.Error(err),
				zap.String("event_name", e.EventName()),
				zap.Int("index", i),
			)
			d.tracer.RecordError(span, err, attribute.Int("failed_at_index", i))
			span.SetStatus(codes.Error, "batch dispatch failed")
			return fmt.Errorf("failed to dispatch event %s at index %d: %w", e.EventName(), i, err)
		}
	}

	span.SetStatus(codes.Ok, "all dispatched")
	return nil
}

// dispatchSync dispatches event synchronously.
func (d *CompositeEventDispatcher) dispatchSync(ctx context.Context, event events.DomainEvent, handlers []messaging.EventHandler) error {
	eventName := event.EventName()
	var errs []error

	for i, h := range handlers {
		if !h.CanHandle(eventName) {
			continue
		}

		handlerCtx, handlerSpan := d.tracer.StartSpan(ctx, "event.handle."+eventName)
		d.tracer.AddAttributes(handlerSpan, attribute.Int("handler.index", i))

		start := time.Now()
		err := h.Handle(handlerCtx, event)
		duration := time.Since(start)

		if err != nil {
			d.logger.Error(handlerCtx, "Event handler failed",
				zap.Error(err),
				zap.String("event_name", eventName),
				zap.Int("handler_index", i),
			)
			d.tracer.RecordError(handlerSpan, err)
			handlerSpan.SetStatus(codes.Error, "handler failed")
			errs = append(errs, fmt.Errorf("handler %d error: %w", i, err))
			if d.metrics != nil {
				d.metrics.RecordCommand(handlerCtx, eventName, "error", duration)
			}
		} else {
			handlerSpan.SetStatus(codes.Ok, "success")
			if d.metrics != nil {
				d.metrics.RecordCommand(handlerCtx, eventName, "success", duration)
			}
		}

		handlerSpan.End()
	}

	if len(errs) > 0 {
		return fmt.Errorf("dispatch errors: %v", errs)
	}
	return nil
}

// dispatchAsync dispatches event asynchronously.
func (d *CompositeEventDispatcher) dispatchAsync(ctx context.Context, event events.DomainEvent, handlers []messaging.EventHandler) error {
	eventName := event.EventName()
	var wg sync.WaitGroup
	errChan := make(chan error, len(handlers))

	for i, h := range handlers {
		if !h.CanHandle(eventName) {
			continue
		}

		wg.Add(1)
		go func(handler messaging.EventHandler, index int) {
			defer wg.Done()

			handlerCtx, handlerSpan := d.tracer.StartSpan(ctx, "event.handle_async."+eventName)
			defer handlerSpan.End()

			d.tracer.AddAttributes(handlerSpan,
				attribute.Int("handler.index", index),
				attribute.Bool("async", true),
			)

			start := time.Now()
			err := handler.Handle(handlerCtx, event)
			duration := time.Since(start)

			if err != nil {
				d.logger.Error(handlerCtx, "Async event handler failed",
					zap.Error(err),
					zap.String("event_name", eventName),
					zap.Int("handler_index", index),
				)
				d.tracer.RecordError(handlerSpan, err)
				handlerSpan.SetStatus(codes.Error, "handler failed")
				errChan <- fmt.Errorf("async handler %d error: %w", index, err)
				if d.metrics != nil {
					d.metrics.RecordCommand(handlerCtx, eventName, "error", duration)
				}
			} else {
				handlerSpan.SetStatus(codes.Ok, "success")
				if d.metrics != nil {
					d.metrics.RecordCommand(handlerCtx, eventName, "success", duration)
				}
			}
		}(h, i)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("async dispatch errors: %v", errs)
	}
	return nil
}

// GetHandlerCount returns number of registered handlers.
func (d *CompositeEventDispatcher) GetHandlerCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers)
}
