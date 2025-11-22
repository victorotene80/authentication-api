package messaging

import "context"

type Command interface {
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
}
