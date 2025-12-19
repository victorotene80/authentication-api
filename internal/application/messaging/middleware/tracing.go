package middleware

import (
	"context"
	"time"
	"fmt"

	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/observability"
	"reflect"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func TracingMiddleware(tracer observability.Tracer) messaging.Middleware {
	return func(next messaging.HandlerFunc) messaging.HandlerFunc {
		return func(ctx context.Context, cmd messaging.Command) (any, error) {
			cmdType := reflect.TypeOf(cmd).String()
			cmdName := getCommandName(cmd)

			// Start a new span for this command execution
			ctx, span := tracer.StartSpan(ctx, fmt.Sprintf("CommandBus.Execute: %s", cmdName),
				trace.WithSpanKind(trace.SpanKindInternal),
				trace.WithAttributes(
					attribute.String("command.type", cmdType),
					attribute.String("command.name", cmdName),
					attribute.String("component", "command_bus"),
				),
			)
			defer span.End()

			// Add command metadata as span attributes
			addCommandMetadata(span, cmd)

			start := time.Now()
			result, err := next(ctx, cmd)
			duration := time.Since(start)

			// Record execution duration
			span.SetAttributes(
				attribute.Int64("command.duration_ms", duration.Milliseconds()),
			)

			if err != nil {
				// Record error details
				span.RecordError(err, trace.WithAttributes(
					attribute.String("error.type", reflect.TypeOf(err).String()),
				))
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "success")
			}

			return result, err
		}
	}
}
