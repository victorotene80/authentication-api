package metrics

import (
	"context"
	"log"
	"runtime"
	"time"
)

func CollectSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)

	go func() {
		defer ticker.Stop()

		var mem runtime.MemStats

		for {
			select {
			case <-ctx.Done():
				log.Println("System metrics collector shutting down...")
				return

			case <-ticker.C:
				runtime.ReadMemStats(&mem)

				GoroutineCount.Set(float64(runtime.NumGoroutine()))

				MemoryUsage.WithLabelValues("alloc").Set(float64(mem.Alloc))
				MemoryUsage.WithLabelValues("total_alloc").Set(float64(mem.TotalAlloc))
				MemoryUsage.WithLabelValues("heap_alloc").Set(float64(mem.HeapAlloc))
				MemoryUsage.WithLabelValues("heap_sys").Set(float64(mem.HeapSys))
			}
		}
	}()
}
