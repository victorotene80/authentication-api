package event

import (
	"authentication/internal/application/contracts/messaging"
	"context"
	"fmt"
	"sync"
)

type CommandBus struct {
	handlers map[string]messaging.CommandHandler
	mu       sync.RWMutex
	chain    []messaging.CommandMiddleware
}

func NewCommandBus() *CommandBus {
	return &CommandBus{
		handlers: make(map[string]messaging.CommandHandler),
		chain:    make([]messaging.CommandMiddleware, 0),
	}
}

func (cb *CommandBus) Register(cmd messaging.Command, handler messaging.CommandHandler) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cmdName := cmd.CommandName()
	if _, exists := cb.handlers[cmdName]; exists {
		return fmt.Errorf("command handler already registered for %s", cmdName)
	}

	cb.handlers[cmdName] = handler
	return nil
}

func (cb *CommandBus) RegisterFunc(cmd messaging.Command, handler messaging.CommandHandlerFunc) error {
	return cb.Register(cmd, handler)
}

func (cb *CommandBus) Use(middleware messaging.CommandMiddleware) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.chain = append(cb.chain, middleware)
}

func (cb *CommandBus) Execute(ctx context.Context, cmd messaging.Command) error {
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


/*func (cb *CommandBus) executeWithMiddleware(ctx context.Context, cmd contracts.Command, handler contracts.CommandHandler, chain []contracts.CommandMiddleware) error {
	if len(chain) == 0 {
		return handler.Handle(ctx, cmd)
	}

	return chain[0].Execute(ctx, cmd, CommandHandlerFunc(func(c context.Context, command contracts.Command) error {
		return cb.executeWithMiddleware(c, command, handler, chain[1:])
	}))
}*/
