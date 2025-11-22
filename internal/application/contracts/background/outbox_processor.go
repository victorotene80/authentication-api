package background

import "context"

type OutboxProcessor interface {
	//Start(ctx context.Context)
	//Stop()
	Process(ctx context.Context) error
}
