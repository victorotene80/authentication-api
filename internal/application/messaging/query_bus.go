package messaging

import (
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/logging"
	"authentication/internal/application/contracts/observability"
	"authentication/shared/utils"
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type QueryBus struct {
	handlers map[string]messaging.UntypedQueryHandler
	mu       sync.RWMutex
	chain    []messaging.UntypedQueryMiddleware
	logger   logging.Logger
	tracer   observability.Tracer
	metrics  *metrics.MetricsRecorder
}

// NewQueryBus creates a new instrumented QueryBus with logger, tracer, and metrics
func NewQueryBus(
	logger logging.Logger,
	tracer observability.Tracer,
	metricsRecorder *metrics.MetricsRecorder,
) *QueryBus {
	return &QueryBus{
		handlers: make(map[string]messaging.UntypedQueryHandler),
		chain:    make([]messaging.UntypedQueryMiddleware, 0),
		logger:   logger.With(zap.String("component", "query_bus")),
		tracer:   tracer,
		metrics:  metricsRecorder,
	}
}

// Register registers a typed query handler
func RegisterQuery[Q messaging.Query, R any](
	qb *QueryBus,
	query Q,
	handler messaging.QueryHandler[Q, R],
) error {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	queryName := query.QueryName()
	if _, exists := qb.handlers[queryName]; exists {
		qb.logger.Warn(context.Background(), "Attempted to register duplicate query handler",
			zap.String("query", queryName),
		)
		return fmt.Errorf("query handler already registered for %s", queryName)
	}

	// Wrap the typed handler in an adapter
	qb.handlers[queryName] = &typedQueryHandlerAdapter[Q, R]{inner: handler}
	qb.logger.Debug(context.Background(), "Query handler registered",
		zap.String("query", queryName),
	)

	return nil
}

// RegisterQueryFunc registers a typed query handler function
func RegisterQueryFunc[Q messaging.Query, R any](
	qb *QueryBus,
	query Q,
	handler messaging.QueryHandlerFunc[Q, R],
) error {
	return RegisterQuery(qb, query, handler)
}

// UseQuery adds a typed middleware to the chain
func UseQuery[Q messaging.Query, R any](
	qb *QueryBus,
	middleware messaging.QueryMiddleware[Q, R],
) {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	qb.logger.Debug(context.Background(), "Query middleware added",
		zap.Int("total_middleware", len(qb.chain)),
	)

	qb.chain = append(qb.chain, &typedQueryMiddlewareAdapter[Q, R]{inner: middleware})
}

// ExecuteQuery executes a query and returns the result
func ExecuteQuery[Q messaging.Query, R any](
	qb *QueryBus,
	ctx context.Context,
	query Q,
) (R, error) {
	var zero R
	queryName := query.QueryName()
	const slowQueryThreshold = 1.0 // seconds

	// Start tracing span
	ctx, span := qb.tracer.StartSpan(
		ctx,
		"query."+queryName,
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	qb.tracer.AddAttributes(span,
		attribute.String("query.name", queryName),
		attribute.String("query.type", "query"),
	)

	qb.logger.Debug(ctx, "Executing query",
		zap.String("query", queryName),
	)

	start := utils.NowUTC()

	// Get handler and middleware chain
	qb.mu.RLock()
	handler, exists := qb.handlers[queryName]
	chain := make([]messaging.UntypedQueryMiddleware, len(qb.chain))
	copy(chain, qb.chain)
	qb.mu.RUnlock()

	if !exists {
		err := fmt.Errorf("no handler registered for query: %s", queryName)
		qb.logger.Error(ctx, "Query handler not found",
			zap.Error(err),
			zap.String("query", queryName),
		)

		qb.tracer.RecordError(span, err,
			attribute.String("error.type", "handler_not_found"),
		)
		span.SetStatus(codes.Error, "handler not found")

		if qb.metrics != nil {
			qb.metrics.RecordQuery(ctx, queryName, "handler_not_found", time.Since(start))
		}

		return zero, err
	}

	// Execute query through middleware
	var result any
	var err error
	if len(chain) == 0 {
		result, err = handler.HandleUntyped(ctx, query)
	} else {
		result, err = qb.executeWithMiddleware(ctx, query, handler, chain)
	}

	// Calculate duration
	durationSecs := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
	}

	if qb.metrics != nil {
		qb.metrics.RecordQuery(ctx, queryName, status, time.Since(start))
	}

	// Logging + Tracing
	if err != nil {
		qb.logger.Error(ctx, "Query execution failed",
			zap.Error(err),
			zap.String("query", queryName),
			zap.Float64("duration_seconds", durationSecs),
		)
		span.SetStatus(codes.Error, "execution failed")
		qb.tracer.RecordError(span, err)
		qb.tracer.AddEvent(span, "query.error",
			attribute.String("query.name", queryName),
			attribute.String("error", err.Error()),
		)
		return zero, err
	}

	qb.logger.Debug(ctx, "Query executed successfully",
		zap.String("query", queryName),
		zap.Float64("duration_seconds", durationSecs),
	)
	span.SetStatus(codes.Ok, "success")
	qb.tracer.AddEvent(span, "query.success",
		attribute.String("query.name", queryName),
	)

	qb.tracer.AddAttributes(span,
		attribute.Float64("duration_seconds", durationSecs),
		attribute.Float64("duration_ms", durationSecs*1000),
		attribute.String("status", status),
	)

	// Detect slow queries
	if durationSecs > slowQueryThreshold {
		qb.logger.Warn(ctx, "Slow query execution detected",
			zap.String("query", queryName),
			zap.Float64("duration_seconds", durationSecs),
		)
		qb.tracer.AddEvent(span, "query.slow",
			attribute.Float64("duration_seconds", durationSecs),
			attribute.String("query.name", queryName),
		)
	}

	// Type assert result to expected type
	if result == nil {
		return zero, nil
	}

	typedResult, ok := result.(R)
	if !ok {
		err := fmt.Errorf("result type mismatch for query %s", queryName)
		qb.logger.Error(ctx, "Type assertion failed",
			zap.Error(err),
			zap.String("query", queryName),
		)
		return zero, err
	}

	return typedResult, nil
}

func (qb *QueryBus) executeWithMiddleware(
	ctx context.Context,
	query messaging.Query,
	handler messaging.UntypedQueryHandler,
	chain []messaging.UntypedQueryMiddleware,
) (any, error) {
	final := handler
	for i := len(chain) - 1; i >= 0; i-- {
		next := final
		m := chain[i]
		final = untypedQueryHandlerFunc(func(c context.Context, q messaging.Query) (any, error) {
			return m.ExecuteUntyped(c, q, next)
		})
	}
	return final.HandleUntyped(ctx, query)
}

func (qb *QueryBus) GetRegisteredQueries() []string {
	qb.mu.RLock()
	defer qb.mu.RUnlock()

	queries := make([]string, 0, len(qb.handlers))
	for queryName := range qb.handlers {
		queries = append(queries, queryName)
	}
	return queries
}

// Helper types
type typedQueryHandlerAdapter[Q messaging.Query, R any] struct {
	inner messaging.QueryHandler[Q, R]
}

func (h *typedQueryHandlerAdapter[Q, R]) HandleUntyped(ctx context.Context, query messaging.Query) (any, error) {
	return h.inner.Handle(ctx, query.(Q))
}

type typedQueryMiddlewareAdapter[Q messaging.Query, R any] struct {
	inner messaging.QueryMiddleware[Q, R]
}

func (m *typedQueryMiddlewareAdapter[Q, R]) ExecuteUntyped(
	ctx context.Context,
	query messaging.Query,
	next messaging.UntypedQueryHandler,
) (any, error) {
	handler := messaging.QueryHandlerFunc[Q, R](func(c context.Context, qq Q) (R, error) {
		result, err := next.HandleUntyped(c, qq)
		if err != nil {
			var zero R
			return zero, err
		}
		return result.(R), nil
	})
	return m.inner.Execute(ctx, query.(Q), handler)
}

type untypedQueryHandlerFunc func(ctx context.Context, query messaging.Query) (any, error)

func (f untypedQueryHandlerFunc) HandleUntyped(ctx context.Context, query messaging.Query) (any, error) {
	return f(ctx, query)
}