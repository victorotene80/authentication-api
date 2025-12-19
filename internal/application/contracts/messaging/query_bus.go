package messaging

import "context"

type Query interface {
	QueryName() string
}

// Generic typed interfaces
type QueryHandler[Q Query, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

type QueryHandlerFunc[Q Query, R any] func(ctx context.Context, query Q) (R, error)

func (f QueryHandlerFunc[Q, R]) Handle(ctx context.Context, query Q) (R, error) {
	return f(ctx, query)
}

type QueryMiddleware[Q Query, R any] interface {
	Execute(ctx context.Context, query Q, next QueryHandler[Q, R]) (R, error)
}

// Untyped interfaces for runtime handling
type UntypedQueryHandler interface {
	HandleUntyped(ctx context.Context, query Query) (any, error)
}

type UntypedQueryMiddleware interface {
	ExecuteUntyped(ctx context.Context, query Query, next UntypedQueryHandler) (any, error)
}

// Adapter from typed to untyped handler
type typedQueryHandlerAdapter[Q Query, R any] struct {
	inner QueryHandler[Q, R]
}

func (h typedQueryHandlerAdapter[Q, R]) HandleUntyped(ctx context.Context, query Query) (any, error) {
	return h.inner.Handle(ctx, query.(Q))
}

// Adapter from typed to untyped middleware
type typedQueryMiddlewareAdapter[Q Query, R any] struct {
	inner QueryMiddleware[Q, R]
}

func (m typedQueryMiddlewareAdapter[Q, R]) ExecuteUntyped(
	ctx context.Context,
	query Query,
	next UntypedQueryHandler,
) (any, error) {
	handler := QueryHandlerFunc[Q, R](func(c context.Context, qq Q) (R, error) {
		result, err := next.HandleUntyped(c, qq)
		if err != nil {
			var zero R
			return zero, err
		}
		return result.(R), nil
	})
	return m.inner.Execute(ctx, query.(Q), handler)
}

// QueryBus interface (optional - you may not need this if using concrete type)
type QueryBus interface {
	ExecuteUntyped(ctx context.Context, query Query) (any, error)
	RegisterUntyped(query Query, handler UntypedQueryHandler) error
	UseUntyped(middleware UntypedQueryMiddleware)
	GetRegisteredQueries() []string
}