package observability

import (
	"context"
	"time"
)

type CommandMetrics interface {
	RecordCommand(
		ctx context.Context,
		commandName string,
		status string,
		duration time.Duration,
	)
}
