// internal/logger/context.go
package logger

import (
	"context"

	"github.com/gin-gonic/gin"
)

// ========================================================================
// CONTEXT HELPERS - Propagate logger through request context
// ========================================================================

type contextKey string

const loggerKey contextKey = "logger"

// ToContext adds logger to context
func ToContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext retrieves logger from context
// Returns global logger if not found
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		return l
	}
	return Global()
}

// FromGinContext retrieves logger from gin.Context
func FromGinContext(c *gin.Context) *Logger {
	if l, exists := c.Get(string(loggerKey)); exists {
		if logger, ok := l.(*Logger); ok {
			return logger
		}
	}
	return Global()
}

// ToGinContext adds logger to gin.Context
func ToGinContext(c *gin.Context, l *Logger) {
	c.Set(string(loggerKey), l)
}
