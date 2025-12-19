package messaging

import (
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/observability"
	"authentication/internal/application/messaging/middleware"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/logging"
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	//"go.opentelemetry.io/otel/attribute"
	//"go.opentelemetry.io/otel/codes"
	//"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

//const slowCommandThresholdSeconds = 1.0

var (
	ErrHandlerNotFound = errors.New("handler not found for command")
	ErrHandlerExists   = errors.New("handler already registered for command")
	ErrNilCommand      = errors.New("command cannot be nil")
)

type CommandBus struct {
	mu              sync.RWMutex
	handlers        map[reflect.Type]handlerWrapper
	middleware      []messaging.Middleware
	logger          logging.Logger
	tracer          observability.Tracer
	metricsRecorder *metrics.MetricsRecorder
}

type handlerWrapper func(ctx context.Context, cmd any) (any, error)

// New creates a new CommandBus instance with default middleware
func New(logger logging.Logger, tracer observability.Tracer, metricsRecorder *metrics.MetricsRecorder) *CommandBus {
	bus := &CommandBus{
		handlers:        make(map[reflect.Type]handlerWrapper),
		middleware:      make([]messaging.Middleware, 0),
		logger:          logger.With(zap.String("component", "command_bus")),
		tracer:          tracer,
		metricsRecorder: metricsRecorder,
	}

	// Register default middleware in order (last added executes first in chain)
	bus.Use(middleware.TracingMiddleware(tracer))
	bus.Use(middleware.MetricsMiddleware(metricsRecorder))
	bus.Use(middleware.LoggingMiddleware(logger))

	return bus
}

// Register registers a command handler
func Register[TCommand messaging.Command, TResult any](
	bus *CommandBus,
	handler messaging.CommandHandler[TCommand, TResult],
) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	var cmd TCommand
	cmdType := reflect.TypeOf(cmd)

	if cmdType.Kind() == reflect.Ptr {
		cmdType = cmdType.Elem()
	}

	if _, exists := bus.handlers[cmdType]; exists {
		return fmt.Errorf("%w: %v", ErrHandlerExists, cmdType)
	}

	// Wrap the typed handler to match handlerWrapper signature
	wrapper := func(ctx context.Context, cmdAny any) (any, error) {
		typedCmd, ok := cmdAny.(TCommand)
		if !ok {
			return nil, fmt.Errorf("invalid command type: expected %T, got %T", cmd, cmdAny)
		}
		return handler.Handle(ctx, typedCmd)
	}

	bus.handlers[cmdType] = wrapper
	bus.logger.Info(context.Background(), "Handler registered",
		zap.String("command_type", cmdType.String()))

	return nil
}

// MustRegister is like Register but panics on error
func MustRegister[TCommand messaging.Command, TResult any](
	bus *CommandBus,
	handler messaging.CommandHandler[TCommand, TResult],
) {
	if err := Register(bus, handler); err != nil {
		panic(err)
	}
}

// Execute executes a command and returns the result
func Execute[TCommand messaging.Command, TResult any](
	bus *CommandBus,
	ctx context.Context,
	cmd TCommand,
) (TResult, error) {
	var zero TResult

	v := reflect.ValueOf(cmd)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return zero, ErrNilCommand
	}

	bus.mu.RLock()
	cmdType := reflect.TypeOf(cmd)
	if cmdType.Kind() == reflect.Ptr {
		cmdType = cmdType.Elem()
	}

	wrapper, exists := bus.handlers[cmdType]
	bus.mu.RUnlock()

	if !exists {
		return zero, fmt.Errorf("%w: %v", ErrHandlerNotFound, cmdType)
	}

	// Build middleware chain
	finalHandler := func(ctx context.Context, cmdAny messaging.Command) (any, error) {
		return wrapper(ctx, cmdAny)
	}

	// Apply middleware in reverse order (last registered executes first)
	for i := len(bus.middleware) - 1; i >= 0; i-- {
		finalHandler = bus.middleware[i](finalHandler)
	}

	// Execute with middleware chain
	result, err := finalHandler(ctx, cmd)
	if err != nil {
		return zero, err
	}

	// Type assert the result
	typedResult, ok := result.(TResult)
	if !ok {
		return zero, fmt.Errorf("invalid result type: expected %T, got %T", zero, result)
	}

	return typedResult, nil
}

// Use adds middleware to the command bus
func (bus *CommandBus) Use(middleware messaging.Middleware) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.middleware = append(bus.middleware, middleware)
}
