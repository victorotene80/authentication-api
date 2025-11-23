package middleware

import (
	"time"
	"fmt"
	"authentication/internal/infrastructure/observability/metrics"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		method := c.Request.Method
		endpoint := c.FullPath()

		// Capture request size
		reqSize := c.Request.ContentLength
		if reqSize > 0 {
			metrics.HTTPRequestSize.WithLabelValues(method, endpoint).Observe(float64(reqSize))
		}

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start).Seconds()

		// Increment request counter
		metrics.HTTPRequestsTotal.
			WithLabelValues(method, endpoint, formatStatus(status)).
			Inc()

		// Observe duration
		metrics.HTTPRequestDuration.
			WithLabelValues(method, endpoint).
			Observe(duration)

		// Observe response size
		resSize := c.Writer.Size()
		if resSize > 0 {
			metrics.HTTPResponseSize.WithLabelValues(method, endpoint).Observe(float64(resSize))
		}
	}
}

func formatStatus(s int) string {
	return fmt.Sprintf("%d", s)
}
