package middleware

import (
	"time"

	"github.com/PeterM45/perfolio-api/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get or generate request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Request.Header.Set("X-Request-ID", requestID)
		}

		// Set request ID in context
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		// Process request
		c.Next()

		// After request
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path = path + "?" + query
		}
		userID, _ := c.Get("userID")

		// Log the request
		event := log.Info()

		// Add failure level for errors
		if statusCode >= 400 {
			event = log.Warn()
		}
		if statusCode >= 500 {
			event = log.Error()
		}

		// Include all fields
		event = event.
			Str("requestID", requestID).
			Str("method", method).
			Str("path", path).
			Str("clientIP", clientIP).
			Int("statusCode", statusCode).
			Dur("latency", latency)

		// Add user ID if available
		if userID != nil {
			event = event.Str("userID", userID.(string))
		}

		event.Msg("HTTP Request")
	}
}

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
