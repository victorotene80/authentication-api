package middleware

import(
	"authentication/shared/logging"
	"authentication/internal/application/contracts/messaging"
	"context"
	"reflect"
	"go.uber.org/zap"
	"time"
	"go.opentelemetry.io/otel/trace"
)


func LoggingMiddleware(logger logging.Logger) messaging.Middleware {
	return func(next messaging.HandlerFunc) messaging.HandlerFunc {
		return func(ctx context.Context, cmd messaging.Command) (any, error) {
			cmdType := reflect.TypeOf(cmd).String()
			cmdName := getCommandName(cmd)

			// Extract trace context if available
			span := trace.SpanFromContext(ctx)
			traceID := span.SpanContext().TraceID().String()
			spanID := span.SpanContext().SpanID().String()

			logger.Info(ctx, "Command execution started",
				zap.String("command.type", cmdType),
				zap.String("command.name", cmdName),
				zap.String("trace.id", traceID),
				zap.String("span.id", spanID),
			)

			start := time.Now()
			result, err := next(ctx, cmd)
			duration := time.Since(start)

			if err != nil {
				logger.Error(ctx, "Command execution failed",
					zap.String("command.type", cmdType),
					zap.String("command.name", cmdName),
					zap.Duration("duration", duration),
					zap.Error(err),
					zap.String("trace.id", traceID),
					zap.String("span.id", spanID),
				)
			} else {
				logger.Info(ctx, "Command execution succeeded",
					zap.String("command.type", cmdType),
					zap.String("command.name", cmdName),
					zap.Duration("duration", duration),
					zap.String("trace.id", traceID),
					zap.String("span.id", spanID),
				)
			}

			return result, err
		}
	}
}
