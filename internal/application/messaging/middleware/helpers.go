package middleware

import (
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/infrastructure/observability/metrics"
	"context"
	"fmt"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// getCommandName extracts a readable command name
func getCommandName(cmd messaging.Command) string {
	cmdType := reflect.TypeOf(cmd)
	if cmdType.Kind() == reflect.Ptr {
		cmdType = cmdType.Elem()
	}
	return cmdType.Name()
}

// addCommandMetadata adds command-specific fields as span attributes
func addCommandMetadata(span trace.Span, cmd messaging.Command) {
	cmdValue := reflect.ValueOf(cmd)
	if cmdValue.Kind() == reflect.Ptr {
		cmdValue = cmdValue.Elem()
	}

	if cmdValue.Kind() != reflect.Struct {
		return
	}

	// Add selected fields as attributes (limit to avoid too many attributes)
	cmdType := cmdValue.Type()
	for i := 0; i < cmdValue.NumField() && i < 10; i++ {
		field := cmdType.Field(i)
		value := cmdValue.Field(i)

		// Skip sensitive fields
		fieldName := field.Name
		if isSensitiveField(fieldName) {
			continue
		}

		// Add safe fields as attributes
		switch value.Kind() {
		case reflect.String:
			if str := value.String(); str != "" && len(str) < 100 {
				span.SetAttributes(attribute.String(fmt.Sprintf("command.%s", fieldName), str))
			}
		case reflect.Int, reflect.Int64:
			span.SetAttributes(attribute.Int64(fmt.Sprintf("command.%s", fieldName), value.Int()))
		case reflect.Bool:
			span.SetAttributes(attribute.Bool(fmt.Sprintf("command.%s", fieldName), value.Bool()))
		}
	}
}

// isSensitiveField checks if a field should be excluded from tracing
func isSensitiveField(fieldName string) bool {
	sensitive := []string{"Password", "Token", "Secret", "APIKey", "AccessToken", "RefreshToken"}
	for _, s := range sensitive {
		if fieldName == s {
			return true
		}
	}
	return false
}

// getMetricsRecorder retrieves metrics recorder from context (if available)
func getMetricsRecorder(ctx context.Context) *metrics.MetricsRecorder {
	if recorder, ok := ctx.Value(metricsRecorderKey{}).(*metrics.MetricsRecorder); ok {
		return recorder
	}
	return nil
}

type metricsRecorderKey struct{}

// WithMetricsRecorder adds metrics recorder to context
func WithMetricsRecorder(ctx context.Context, recorder *metrics.MetricsRecorder) context.Context {
	return context.WithValue(ctx, metricsRecorderKey{}, recorder)
}
