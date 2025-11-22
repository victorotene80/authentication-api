package event

import (
	"authentication/internal/application/contracts/messaging"
	"context"
	"fmt"
	"sync"
)

type QueryBus struct {
	handlers map[string]messaging.QueryHandler
	mu       sync.RWMutex
	chain    []messaging.QueryMiddleware
}

func NewQueryBus() *QueryBus {
	return &QueryBus{
		handlers: make(map[string]messaging.QueryHandler),
		chain:    make([]messaging.QueryMiddleware, 0),
	}
}

func (qb *QueryBus) Register(query messaging.Query, handler messaging.QueryHandler) error {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	queryName := query.QueryName()
	if _, exists := qb.handlers[queryName]; exists {
		return fmt.Errorf("query handler already registered for %s", queryName)
	}

	qb.handlers[queryName] = handler
	return nil
}

func (qb *QueryBus) RegisterFunc(query messaging.Query, handler messaging.QueryHandlerFunc) error {
	return qb.Register(query, handler)
}

func (qb *QueryBus) Use(middleware messaging.QueryMiddleware) {
	qb.mu.Lock()
	defer qb.mu.Unlock()
	qb.chain = append(qb.chain, middleware)
}

func (qb *QueryBus) Execute(ctx context.Context, query messaging.Query) (interface{}, error) {
	qb.mu.RLock()
	handler, exists := qb.handlers[query.QueryName()]
	chain := make([]messaging.QueryMiddleware, len(qb.chain))
	copy(chain, qb.chain)
	qb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no handler registered for query: %s", query.QueryName())
	}

	if len(chain) == 0 {
		return handler.Handle(ctx, query)
	}

	return qb.executeWithMiddleware(ctx, query, handler, chain)
}

func (qb *QueryBus) executeWithMiddleware(ctx context.Context, query messaging.Query, handler messaging.QueryHandler, chain []messaging.QueryMiddleware) (interface{}, error) {
	if len(chain) == 0 {
		return handler.Handle(ctx, query)
	}

	return chain[0].Execute(ctx, query, messaging.QueryHandlerFunc(func(c context.Context, q messaging.Query) (interface{}, error) {
		return qb.executeWithMiddleware(c, q, handler, chain[1:])
	}))
}
