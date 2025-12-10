// internal/middleware/request_limit_middleware.go
package middleware

import (
	"net/http"

	"barber-booking-system/internal/config"
	"barber-booking-system/internal/utils"

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
		SkipPaths:     config.DefaultSkipPaths,
		SkipMethods:   []string{http.MethodGet, http.MethodHead, http.MethodOptions},
		CustomMessage: "Request body exceeds maximum allowed size",
	}
}

// DefaultRequestBodyLimit creates a middleware with default settings
func DefaultRequestBodyLimit(maxSize int64) gin.HandlerFunc {
	return LimitRequestBody(DefaultRequestBodyLimitConfig(maxSize))
}

// LimitRequestBody creates a middleware that limits request body size
func LimitRequestBody(cfg RequestBodyLimitConfig) gin.HandlerFunc {
	skipPaths := utils.BuildStringSet(cfg.SkipPaths)
	skipMethods := utils.BuildStringSet(cfg.SkipMethods)

	return func(c *gin.Context) {
		if skipPaths[c.Request.URL.Path] || skipMethods[c.Request.Method] {
			c.Next()
			return
		}

		if c.Request.ContentLength > cfg.MaxSize {
			RespondWithError(c, &AppError{
				StatusCode: http.StatusRequestEntityTooLarge,
				Code:       "REQUEST_TOO_LARGE",
				Message:    cfg.CustomMessage,
				Details: map[string]interface{}{
					"max_size_bytes": cfg.MaxSize,
					"max_size_mb":    float64(cfg.MaxSize) / (1024 * 1024),
					"received_bytes": c.Request.ContentLength,
				},
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.MaxSize)
		c.Next()
	}
}
