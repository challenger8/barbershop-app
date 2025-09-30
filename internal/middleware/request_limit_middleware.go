// internal/middleware/request_limit_middleware.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestBodyLimitConfig defines configuration for request body limits
type RequestBodyLimitConfig struct {
	MaxSize       int64
	SkipPaths     []string
	SkipMethods   []string
	CustomMessage string
}

// DefaultRequestBodyLimitConfig returns default configuration
func DefaultRequestBodyLimitConfig(maxSize int64) RequestBodyLimitConfig {
	return RequestBodyLimitConfig{
		MaxSize:       maxSize,
		SkipPaths:     []string{"/health", "/metrics"},
		SkipMethods:   []string{http.MethodGet, http.MethodHead, http.MethodOptions},
		CustomMessage: "Request body exceeds maximum allowed size",
	}
}

// DefaultRequestBodyLimit creates a middleware with default settings
func DefaultRequestBodyLimit(maxSize int64) gin.HandlerFunc {
	return LimitRequestBody(DefaultRequestBodyLimitConfig(maxSize))
}

// LimitRequestBody creates a middleware that limits request body size
func LimitRequestBody(config RequestBodyLimitConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	skipMethods := make(map[string]bool)
	for _, method := range config.SkipMethods {
		skipMethods[method] = true
	}

	return func(c *gin.Context) {
		if skipPaths[c.Request.URL.Path] || skipMethods[c.Request.Method] {
			c.Next()
			return
		}

		if c.Request.ContentLength > config.MaxSize {
			RespondWithError(c, &AppError{
				StatusCode: http.StatusRequestEntityTooLarge,
				Code:       "REQUEST_TOO_LARGE",
				Message:    config.CustomMessage,
				Details: map[string]interface{}{
					"max_size_bytes": config.MaxSize,
					"max_size_mb":    float64(config.MaxSize) / (1024 * 1024),
					"received_bytes": c.Request.ContentLength,
				},
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxSize)
		c.Next()
	}
}
