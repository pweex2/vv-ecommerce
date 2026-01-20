package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger middleware logs request details with TraceID
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		traceID := GetTraceID(c)
		latency := time.Since(start)

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[HTTP] %s | %3d | %13v | %-7s | %s | TraceID: %s",
			start.Format("2006/01/02 - 15:04:05"),
			c.Writer.Status(),
			latency,
			c.Request.Method,
			path,
			traceID,
		)
	}
}
