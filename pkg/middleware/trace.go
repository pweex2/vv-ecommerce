package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	TraceIDHeader = "X-Trace-ID"
	TraceIDKey    = "trace_id"
)

// TraceID middleware checks for X-Trace-ID header or generates a new one
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Set in Gin context
		c.Set(TraceIDKey, traceID)

		// Set in Response Header
		c.Writer.Header().Set(TraceIDHeader, traceID)

		c.Next()
	}
}

// GetTraceID extracts TraceID from gin.Context
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}
