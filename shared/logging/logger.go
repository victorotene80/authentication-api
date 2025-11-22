package logging

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const loggerKey contextKey = "logger"

var globalLogger *zapLogger

// zapLogger wraps zap.Logger
type zapLogger struct {
	zap *zap.Logger
}

// Ensure interface
var _ Logger = (*zapLogger)(nil)

// Initialize creates a production-ready logger writing to stdout
func Initialize() error {
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			MessageKey:     "msg",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			LineEnding:     zapcore.DefaultLineEnding,
		},
	}

	logger, err := config.Build(zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return err
	}

	globalLogger = &zapLogger{zap: logger}
	return nil
}

// Get returns the global logger
func Get() Logger {
	if globalLogger == nil {
		globalLogger = &zapLogger{zap: zap.NewNop()}
	}
	return globalLogger
}

// WithContext stores logger in context
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves logger or returns global
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return Get()
}

// Adds trace_id if present
func WithTraceContext(ctx context.Context) Logger {
	logger := FromContext(ctx)
	if span := trace.SpanFromContext(ctx); span != nil {
		sc := span.SpanContext()
		if sc.IsValid() {
			return logger.With(zap.String("trace_id", sc.TraceID().String()))
		}
	}
	return logger
}

// --- Logger Implementation ---

func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.withTrace(ctx).Debug(msg, fields...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.withTrace(ctx).Info(msg, fields...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.withTrace(ctx).Warn(msg, fields...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.withTrace(ctx).Error(msg, fields...)
}

func (l *zapLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.withTrace(ctx).Fatal(msg, fields...)
}

func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{zap: l.zap.With(fields...)}
}

func (l *zapLogger) withTrace(ctx context.Context) *zap.Logger {
	if span := trace.SpanFromContext(ctx); span != nil {
		sc := span.SpanContext()
		if sc.IsValid() {
			return l.zap.With(zap.String("trace_id", sc.TraceID().String()))
		}
	}
	return l.zap
}

// --- Convenience functions for global logger ---

func Debug(msg string, fields ...zap.Field) { globalLogger.zap.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { globalLogger.zap.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { globalLogger.zap.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { globalLogger.zap.Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { globalLogger.zap.Fatal(msg, fields...) }

// Context-aware global logging
func DebugCtx(ctx context.Context, msg string, fields ...zap.Field) { WithTraceContext(ctx).Debug(ctx, msg, fields...) }
func InfoCtx(ctx context.Context, msg string, fields ...zap.Field)  { WithTraceContext(ctx).Info(ctx, msg, fields...) }
func WarnCtx(ctx context.Context, msg string, fields ...zap.Field)  { WithTraceContext(ctx).Warn(ctx, msg, fields...) }
func ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) { WithTraceContext(ctx).Error(ctx, msg, fields...) }

// Sync flushes logs
func Sync() error {
	if globalLogger != nil && globalLogger.zap != nil {
		return globalLogger.zap.Sync()
	}
	return nil
}