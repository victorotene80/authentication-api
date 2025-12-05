package messaging

import "context"

type Command interface {
	CommandName() string
}

// Typed Handler
type CommandHandler[C Command, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

// Typed Functional Handler
type CommandHandlerFunc[C Command, R any] func(ctx context.Context, cmd C) (R, error)

func (f CommandHandlerFunc[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	return f(ctx, cmd)
}

// Typed Middleware
type CommandMiddleware[C Command, R any] interface {
	Execute(ctx context.Context, cmd C, next CommandHandler[C, R]) (R, error)
}

// A wrapper type to hide generics inside the bus registry
type UntypedHandler interface {
	HandleUntyped(ctx context.Context, cmd Command) (any, error)
}

type UntypedMiddleware interface {
	ExecuteUntyped(ctx context.Context, cmd Command, next UntypedHandler) (any, error)
}

// Adapters so the bus can store typed handlers but call them untyped
type typedHandlerAdapter[C Command, R any] struct {
	inner CommandHandler[C, R]
}

func (h typedHandlerAdapter[C, R]) HandleUntyped(ctx context.Context, cmd Command) (any, error) {
	return h.inner.Handle(ctx, cmd.(C))
}

type typedMiddlewareAdapter[C Command, R any] struct {
	inner CommandMiddleware[C, R]
}

func (m typedMiddlewareAdapter[C, R]) ExecuteUntyped(
	ctx context.Context,
	cmd Command,
	next UntypedHandler,
) (any, error) {

	handler := CommandHandlerFunc[C, R](func(c context.Context, cc C) (R, error) {
		result, err := next.HandleUntyped(c, cc)
		if err != nil {
			var zero R
			return zero, err
		}
		return result.(R), nil
	})

	return m.inner.Execute(ctx, cmd.(C), handler)
}




























/*type Command interface {
	CommandName() string
}

type CommandHandler interface {
	Handle(ctx context.Context, cmd Command) error
}

type CommandHandlerFunc func(ctx context.Context, cmd Command) error

func (f CommandHandlerFunc) Handle(ctx context.Context, cmd Command) error {
	return f(ctx, cmd)
}

type CommandMiddleware interface {
	Execute(ctx context.Context, cmd Command, next CommandHandler) error
}

type CommandBus interface {
	Execute(ctx context.Context, cmd Command) error
	Register(cmd Command, handler CommandHandler) error
	RegisterFunc(cmd Command, handler CommandHandlerFunc) error
	Use(middleware CommandMiddleware)
}*/
