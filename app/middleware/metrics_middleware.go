package middleware

import (
	"strconv"
	"time"

	"group1-userservice/app/metrics"

	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		method := c.Request.Method
		elapsed := time.Since(start).Seconds()

		metrics.HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HttpRequestDurationSeconds.WithLabelValues(method, path, status).Observe(elapsed)
	}
}
