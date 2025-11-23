package event

import (
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/logging"
	"authentication/shared/tracing"
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

type CommandBus struct {
	handlers map[string]messaging.CommandHandler
	mu       sync.RWMutex
	chain    []messaging.CommandMiddleware
	logger   logging.Logger
	tracingr tracing.Tracer
}

func NewCommandBus(logger logging.Logger, tracingr tracing.Tracer) *CommandBus {
	return &CommandBus{
		handlers: make(map[string]messaging.CommandHandler),
		chain:    make([]messaging.CommandMiddleware, 0),
		logger:   logger.With(zap.String("component", "command_bus")),
		tracingr: tracingr,
	}
}

func (cb *CommandBus) Register(cmd messaging.Command, handler messaging.CommandHandler) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cmdName := cmd.CommandName()
	if _, exists := cb.handlers[cmdName]; exists {
		cb.logger.Warn(context.Background(), "Attempted to register duplicate command handler",
			zap.String("command", cmdName),
		)
		return fmt.Errorf("command handler already registered for %s", cmdName)
	}

	cb.handlers[cmdName] = handler

	cb.logger.Debug(context.Background(), "Command handler registered",
		zap.String("command", cmdName),
	)

	return nil
}

func (cb *CommandBus) RegisterFunc(cmd messaging.Command, handler messaging.CommandHandlerFunc) error {
	return cb.Register(cmd, handler)
}

func (cb *CommandBus) Use(middleware messaging.CommandMiddleware) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.logger.Debug(context.Background(), "Command middleware added",
		zap.Int("total_middleware", len(cb.chain)),
	)

	cb.chain = append(cb.chain, middleware)
}

func (cb *CommandBus) Execute(ctx context.Context, cmd messaging.Command) error {
	cmdName := cmd.CommandName()
	const slowCommandThreshold = 1.0 // seconds

	// Start tracing span
	ctx, span := cb.tracingr.StartSpan(
		ctx,
		"command."+cmdName,
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	cb.tracingr.AddAttributes(span,
		attribute.String("command.name", cmdName),
		attribute.String("command.type", "command"),
	)

	cb.logger.Debug(ctx, "Executing command",
		zap.String("command", cmdName),
	)

	start := utils.NowUTC()

	// Get handler
	cb.mu.RLock()
	handler, exists := cb.handlers[cmdName]
	chain := make([]messaging.CommandMiddleware, len(cb.chain))
	copy(chain, cb.chain)
	cb.mu.RUnlock()

	if !exists {
		err := fmt.Errorf("no handler registered for command: %s", cmdName)

		cb.logger.Error(ctx, "Command handler not found",
			zap.Error(err),
			zap.String("command", cmdName),
		)

		cb.tracingr.RecordError(span, err,
			attribute.String("error.type", "handler_not_found"),
		)
		span.SetStatus(codes.Error, "handler not found")

		metrics.CommandExecutionTotal.WithLabelValues(cmdName, "handler_not_found").Inc()
		return err
	}

	// --- Execute command ---
	var err error
	if len(chain) == 0 {
		err = handler.Handle(ctx, cmd)
	} else {
		err = cb.executeWithMiddleware(ctx, cmd, handler, chain)
	}

	// Calculate duration
	durationSecs := time.Since(start).Seconds()

	// Metrics
	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.CommandExecutionTotal.WithLabelValues(cmdName, status).Inc()
	metrics.CommandExecutionDuration.WithLabelValues(cmdName).Observe(durationSecs)

	// Logging + Tracing
	if err != nil {
		cb.logger.Error(ctx, "Command execution failed",
			zap.Error(err),
			zap.String("command", cmdName),
			zap.Float64("duration_seconds", durationSecs),
		)

		span.SetStatus(codes.Error, "execution failed")
		cb.tracingr.RecordError(span, err)
		cb.tracingr.AddEvent(span, "command.error",
			attribute.String("command.name", cmdName),
			attribute.String("error", err.Error()),
		)
	} else {
		cb.logger.Debug(ctx, "Command executed successfully",
			zap.String("command", cmdName),
			zap.Float64("duration_seconds", durationSecs),
		)

		span.SetStatus(codes.Ok, "success")
		cb.tracingr.AddEvent(span, "command.success",
			attribute.String("command.name", cmdName),
		)
	}

	// Attach final attributes
	cb.tracingr.AddAttributes(span,
		attribute.Float64("duration_seconds", durationSecs),
		attribute.Float64("duration_ms", durationSecs*1000),
		attribute.String("status", status),
	)

	// --- Slow Command Detection ---
	if durationSecs > slowCommandThreshold {
		cb.logger.Warn(ctx, "Slow command execution detected",
			zap.String("command", cmdName),
			zap.Float64("duration_seconds", durationSecs),
		)

		cb.tracingr.AddEvent(span, "command.slow",
			attribute.Float64("duration_seconds", durationSecs),
			attribute.String("command.name", cmdName),
		)
	}

	return err
}

func (cb *CommandBus) executeWithMiddleware(ctx context.Context, cmd messaging.Command, handler messaging.CommandHandler, chain []messaging.CommandMiddleware) error {
	final := handler
	for i := len(chain) - 1; i >= 0; i-- {
		next := final
		m := chain[i]
		final = messaging.CommandHandlerFunc(func(c context.Context, command messaging.Command) error {
			return m.Execute(c, command, next)
		})
	}
	return final.Handle(ctx, cmd)
}

func (cb *CommandBus) GetRegisteredCommands() []string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	commands := make([]string, 0, len(cb.handlers))
	for cmdName := range cb.handlers {
		commands = append(commands, cmdName)
	}

	return commands
}

/*func (cb *CommandBus) executeWithMiddleware(ctx context.Context, cmd contracts.Command, handler contracts.CommandHandler, chain []contracts.CommandMiddleware) error {
	if len(chain) == 0 {
		return handler.Handle(ctx, cmd)
	}

	return chain[0].Execute(ctx, cmd, CommandHandlerFunc(func(c context.Context, command contracts.Command) error {
		return cb.executeWithMiddleware(c, command, handler, chain[1:])
	}))
}*/


/*func (cb *CommandBus) Execute(ctx context.Context, cmd messaging.Command) error {
	cb.mu.RLock()
	handler, exists := cb.handlers[cmd.CommandName()]
	chain := make([]messaging.CommandMiddleware, len(cb.chain))
	copy(chain, cb.chain)
	cb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for command: %s", cmd.CommandName())
	}

	if len(chain) == 0 {
		return handler.Handle(ctx, cmd)
	}

	return cb.executeWithMiddleware(ctx, cmd, handler, chain)
}*/