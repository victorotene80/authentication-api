package metrics

import (
	"context"
	"database/sql"
	"log"
	"time"
)

func CollectDBPoolMetrics(ctx context.Context, db *sql.DB, dbName string, env string) {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("DB metrics collector panic: %v", r)
			}
			ticker.Stop()
			log.Println("DB metrics collector stopped.")
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("DB metrics collector shutting down...")
				return

			case <-ticker.C:
				stats := db.Stats()

				DatabaseConnectionPoolSize.WithLabelValues("idle", dbName, env).Set(float64(stats.Idle))
				DatabaseConnectionPoolSize.WithLabelValues("in_use", dbName, env).Set(float64(stats.InUse))
				DatabaseConnectionPoolSize.WithLabelValues("open", dbName, env).Set(float64(stats.OpenConnections))
				DatabaseConnectionPoolSize.WithLabelValues("wait_count", dbName, env).Set(float64(stats.WaitCount))
				DatabaseConnectionPoolSize.WithLabelValues("wait_duration", dbName, env).Set(stats.WaitDuration.Seconds())
			}
		}
	}()
}


 func MeasureDBQueryWithContext(
	ctx context.Context,
	operation string,
	table string,
	dbName string,
	env string,
	fn func(context.Context) error,
) error {

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	DatabaseQueryDuration.WithLabelValues(operation, table, dbName, env).Observe(duration)
	DatabaseQueryTotal.WithLabelValues(operation, table, status, dbName, env).Inc()

	return err
}

