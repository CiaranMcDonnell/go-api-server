package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}
		method := c.Request.Method

		HttpRequestsTotal.WithLabelValues(method, route, status).Inc()
		HttpRequestDuration.WithLabelValues(method, route).Observe(duration)
	}
}
