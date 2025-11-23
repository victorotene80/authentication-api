package event

import (
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/domain/events"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/logging"
	"authentication/shared/tracing"
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// CompositeEventDispatcher implements contracts.EventDispatcher
type CompositeEventDispatcher struct {
	handlers []messaging.EventHandler
	async    bool
	mu       sync.RWMutex
	logger   logging.Logger
	tracer   tracing.Tracer
}

// NewCompositeEventDispatcher creates a new multi-handler dispatcher
func NewCompositeEventDispatcher(async bool, logger logging.Logger, tracer tracing.Tracer) messaging.EventDispatcher {
	return &CompositeEventDispatcher{
		handlers: make([]messaging.EventHandler, 0),
		async:    async,
		logger:   logger.With(zap.String("component", "event_dispatcher")),
		tracer:   tracer,
	}
}

// RegisterHandler adds a new event handler
func (d *CompositeEventDispatcher) RegisterHandler(handler messaging.EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers = append(d.handlers, handler)

	d.logger.Debug(context.Background(), "Event handler registered",
		zap.Int("total_handlers", len(d.handlers)),
	)
}

// Dispatch sends event to all registered handlers (legacy method for compatibility)
func (d *CompositeEventDispatcher) Dispatch(event events.DomainEvent) error {
	return d.DispatchWithContext(context.Background(), event)
}

// DispatchWithContext dispatches event with context and full observability
func (d *CompositeEventDispatcher) DispatchWithContext(ctx context.Context, event events.DomainEvent) error {
	eventName := event.EventName()

	// Start trace span
	ctx, span := d.tracer.StartSpan(
		ctx,
		"event.dispatch."+eventName,
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
	handlers := d.handlers
	d.mu.RUnlock()

	// Count handlers that will process this event
	var handlerCount int
	for _, handler := range handlers {
		if handler.CanHandle(eventName) {
			handlerCount++
		}
	}

	d.tracer.AddAttributes(span,
		attribute.Int("handlers.count", handlerCount),
	)

	var err error
	if d.async {
		err = d.dispatchAsync(ctx, event, handlers)
	} else {
		err = d.dispatchSync(ctx, event, handlers)
	}

	duration := time.Since(start).Seconds()

	// Record metrics
	status := "success"
	if err != nil {
		status = "error"
		d.logger.Error(ctx, "Event dispatch failed",
			zap.Error(err),
			zap.String("event_name", eventName),
			zap.Duration("duration", time.Duration(duration*float64(time.Second))),
		)

		d.tracer.RecordError(span, err)
		span.SetStatus(codes.Error, "dispatch failed")
	} else {
		d.logger.Debug(ctx, "Event dispatched successfully",
			zap.String("event_name", eventName),
			zap.Int("handlers_executed", handlerCount),
			zap.Duration("duration", time.Duration(duration*float64(time.Second))),
		)

		span.SetStatus(codes.Ok, "success")
	}

	metrics.EventsPublished.WithLabelValues(eventName, status).Inc()

	d.tracer.AddAttributes(span,
		attribute.Float64("duration_ms", duration*1000),
		attribute.String("status", status),
	)

	d.tracer.AddEvent(span, "event.dispatched",
		attribute.Int("handlers_count", handlerCount),
	)

	return err
}

// DispatchAll dispatches multiple events
func (d *CompositeEventDispatcher) DispatchAll(ctx context.Context, domainEvents []events.DomainEvent) error {
	ctx, span := d.tracer.StartSpan(ctx, "event.dispatch_all")
	defer span.End()

	d.tracer.AddAttributes(span,
		attribute.Int("events.count", len(domainEvents)),
	)

	d.logger.Info(ctx, "Dispatching multiple events",
		zap.Int("count", len(domainEvents)),
	)

	for i, event := range domainEvents {
		if err := d.DispatchWithContext(ctx, event); err != nil {
			d.logger.Error(ctx, "Failed to dispatch event in batch",
				zap.Error(err),
				zap.String("event_name", event.EventName()),
				zap.Int("index", i),
			)

			d.tracer.RecordError(span, err,
				attribute.Int("failed_at_index", i),
			)
			span.SetStatus(codes.Error, "batch dispatch failed")

			return fmt.Errorf("failed to dispatch event %s at index %d: %w", event.EventName(), i, err)
		}
	}

	span.SetStatus(codes.Ok, "all dispatched")
	return nil
}

// dispatchSync dispatches event synchronously to all handlers
func (d *CompositeEventDispatcher) dispatchSync(ctx context.Context, event events.DomainEvent, handlers []messaging.EventHandler) error {
	eventName := event.EventName()
	var errs []error

	for i, handler := range handlers {
		if !handler.CanHandle(eventName) {
			continue
		}

		// Create span for each handler
		handlerCtx, handlerSpan := d.tracer.StartSpan(ctx, "event.handle."+eventName)
		d.tracer.AddAttributes(handlerSpan,
			attribute.String("handler.index", fmt.Sprintf("%d", i)),
		)

		start := time.Now()
		err := handler.Handle(handlerCtx, event)
		duration := time.Since(start).Seconds()

		if err != nil {
			d.logger.Error(handlerCtx, "Event handler failed",
				zap.Error(err),
				zap.String("event_name", eventName),
				zap.Int("handler_index", i),
			)

			d.tracer.RecordError(handlerSpan, err)
			handlerSpan.SetStatus(codes.Error, "handler failed")
			errs = append(errs, fmt.Errorf("handler %d error: %w", i, err))

			metrics.EventsProcessed.WithLabelValues(eventName, "error").Inc()
		} else {
			handlerSpan.SetStatus(codes.Ok, "success")
			metrics.EventsProcessed.WithLabelValues(eventName, "success").Inc()
		}

		metrics.EventProcessingDuration.WithLabelValues(eventName).Observe(duration)
		handlerSpan.End()
	}

	if len(errs) > 0 {
		return fmt.Errorf("dispatch errors: %v", errs)
	}

	return nil
}

// dispatchAsync dispatches event asynchronously to all handlers
func (d *CompositeEventDispatcher) dispatchAsync(ctx context.Context, event events.DomainEvent, handlers []messaging.EventHandler) error {
	eventName := event.EventName()
	var wg sync.WaitGroup
	errChan := make(chan error, len(handlers))

	for i, handler := range handlers {
		if !handler.CanHandle(eventName) {
			continue
		}

		wg.Add(1)
		go func(h messaging.EventHandler, index int) {
			defer wg.Done()

			// Create span for async handler
			handlerCtx, handlerSpan := d.tracer.StartSpan(ctx, "event.handle_async."+eventName)
			defer handlerSpan.End()

			d.tracer.AddAttributes(handlerSpan,
				attribute.String("handler.index", fmt.Sprintf("%d", index)),
				attribute.Bool("async", true),
			)

			start := time.Now()
			err := h.Handle(handlerCtx, event)
			duration := time.Since(start).Seconds()

			if err != nil {
				d.logger.Error(handlerCtx, "Async event handler failed",
					zap.Error(err),
					zap.String("event_name", eventName),
					zap.Int("handler_index", index),
				)

				d.tracer.RecordError(handlerSpan, err)
				handlerSpan.SetStatus(codes.Error, "handler failed")
				errChan <- fmt.Errorf("async handler %d error: %w", index, err)

				metrics.EventsProcessed.WithLabelValues(eventName, "error").Inc()
			} else {
				handlerSpan.SetStatus(codes.Ok, "success")
				metrics.EventsProcessed.WithLabelValues(eventName, "success").Inc()
			}

			metrics.EventProcessingDuration.WithLabelValues(eventName).Observe(duration)
		}(handler, i)
	}

	// Wait for all handlers to complete
	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("async dispatch errors: %v", errs)
	}

	return nil
}

func (d *CompositeEventDispatcher) GetHandlerCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers)
}
