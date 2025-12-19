package messaging

import "context"


/*type Command interface {
	CommandName() string 
}

type CommandHandler[TCommand Command, TResult any] interface {
	Handle(ctx context.Context, cmd TCommand) (TResult, error)
}

type Middleware func(next ExecuteFunc) ExecuteFunc

type ExecuteFunc func(ctx context.Context, cmd Command) (any, error)*/

// Command is a marker interface for all commands
type Command interface{}

// CommandHandler processes a command and returns a result
type CommandHandler[TCommand Command, TResult any] interface {
	Handle(ctx context.Context, cmd TCommand) (TResult, error)
}

// Middleware wraps command execution with cross-cutting concerns
type Middleware func(next HandlerFunc) HandlerFunc

// HandlerFunc is the function signature for command execution
type HandlerFunc func(ctx context.Context, cmd Command) (any, error)
