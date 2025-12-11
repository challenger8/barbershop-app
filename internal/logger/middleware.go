// internal/logger/middleware.go
package logger

import (
	"time"

	"github.com/gin-gonic/gin"
)

// ========================================================================
// LOGGING MIDDLEWARE - Injects logger into request context
// ========================================================================

// Middleware creates a gin middleware that:
// 1. Creates a request-scoped logger with request metadata
// 2. Injects it into the context
// 3. Logs request completion with timing
func Middleware(baseLogger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get or generate request ID
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = c.GetHeader("X-Request-ID")
		}

		// Create request-scoped logger with metadata
		reqLogger := baseLogger.With().
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Logger()

		// Add user info if available
		if userID, exists := c.Get("user_id"); exists {
			reqLogger = reqLogger.With().Int("user_id", userID.(int)).Logger()
		}

		// Store logger in context
		ToGinContext(c, reqLogger)

		// Also store in request context for services
		ctx := ToContext(c.Request.Context(), reqLogger)
		c.Request = c.Request.WithContext(ctx)

		// Log request start (debug level)
		reqLogger.Debug("Request started").
			Str("user_agent", c.Request.UserAgent()).
			Str("query", c.Request.URL.RawQuery).
			Send()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Determine log level based on status code
		status := c.Writer.Status()
		event := reqLogger.Info("Request completed")

		if status >= 500 {
			event = reqLogger.Error(nil).Str("level", "error")
		} else if status >= 400 {
			event = reqLogger.Warn("Request completed")
		}

		// Log request completion
		event.
			Int("status", status).
			Dur("duration", duration).
			Int("size", c.Writer.Size()).
			Send()

			// Log errors if any
			// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				reqLogger.Error(e.Err).
					Int("error_type", int(e.Type)).
					Msg("Request error")
			}
		}
	}
}

// SkipPaths returns a middleware that skips logging for certain paths
func SkipPaths(paths []string, next gin.HandlerFunc) gin.HandlerFunc {
	skipSet := make(map[string]bool)
	for _, p := range paths {
		skipSet[p] = true
	}

	return func(c *gin.Context) {
		if skipSet[c.Request.URL.Path] {
			c.Next()
			return
		}
		next(c)
	}
}
