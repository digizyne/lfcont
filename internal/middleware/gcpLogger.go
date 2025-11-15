package middleware

import (
	"time"

	"cloud.google.com/go/logging"
	"github.com/gin-gonic/gin"
)

func GcpLogger(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Format timestamp like Gin default
		timestamp := start.Format("2006/01/02 - 15:04:05")

		logger.Log(logging.Entry{
			Severity: logging.Info,
			Payload: map[string]interface{}{
				"timestamp": timestamp,
				"status":    c.Writer.Status(),
				"duration":  duration.String(),
				"client_ip": clientIP,
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
				"query":     c.Request.URL.RawQuery,
			},
		})
	}
}
