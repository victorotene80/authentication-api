package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/utils"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := utils.NowUTC()
		path := c.FullPath()
		
		// If path is empty (404), use the raw path
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		metrics.HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)
	}
}