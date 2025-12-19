package middleware

import (
	"context"
	"time"

	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/observability"
)

func MetricsMiddleware(recorder observability.CommandMetrics) messaging.Middleware {
	return func(next messaging.HandlerFunc) messaging.HandlerFunc {
		return func(ctx context.Context, cmd messaging.Command) (any, error) {
			if recorder == nil {
				return next(ctx, cmd)
			}

			start := time.Now()
			result, err := next(ctx, cmd)

			status := "success"
			if err != nil {
				status = "failure"
			}

			commandName := getCommandName(cmd)

			recorder.RecordCommand(
				ctx,
				commandName,
				status,
				time.Since(start),
			)

			return result, err
		}
	}
}
