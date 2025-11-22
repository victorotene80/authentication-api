package messaging

import "context"

type Query interface {
    QueryName() string
}

type QueryHandler interface {
    Handle(ctx context.Context, query Query) (interface{}, error)
}

type QueryHandlerFunc func(ctx context.Context, query Query) (interface{}, error)

func (f QueryHandlerFunc) Handle(ctx context.Context, query Query) (interface{}, error) {
    return f(ctx, query)
}

type QueryMiddleware interface {
    Execute(ctx context.Context, query Query, next QueryHandler) (interface{}, error)
}

type QueryBus interface {
    Execute(ctx context.Context, query Query) (interface{}, error)
    Register(query Query, handler QueryHandler) error
    RegisterFunc(query Query, handler QueryHandlerFunc) error
    Use(middleware QueryMiddleware)
}