package observability

type CommandMetrics interface {
	RecordCommandDuration(commandName string, durationSeconds float64)
	IncrementCommandCount(commandName string, status string)
}

type QueryMetrics interface {
	RecordQueryDuration(queryName string, durationSeconds float64)
	IncrementQueryCount(queryName string, status string)
}