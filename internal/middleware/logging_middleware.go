// internal/middleware/logging_middleware.go
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"barber-booking-system/internal/config"
	"barber-booking-system/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LogFormat defines the log output format
type LogFormat string

const (
	JSONFormat LogFormat = "json"
	TextFormat LogFormat = "text"
)

// LoggerConfig defines the configuration for the logger middleware
type LoggerConfig struct {
	Format          LogFormat
	SkipPaths       []string
	LogRequestBody  bool
	LogResponseBody bool
	MaxBodySize     int // Maximum body size to log (in bytes)
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Format:          JSONFormat,
		SkipPaths:       config.DefaultSkipPaths,
		LogRequestBody:  false,
		LogResponseBody: false,
		MaxBodySize:     1024, // 1KB
	}
}

// LogEntry represents a single log entry
type LogEntry struct {
	RequestID    string                 `json:"request_id"`
	Timestamp    string                 `json:"timestamp"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	Query        string                 `json:"query,omitempty"`
	StatusCode   int                    `json:"status_code"`
	Latency      string                 `json:"latency"`
	ClientIP     string                 `json:"client_ip"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	RequestBody  string                 `json:"request_body,omitempty"`
	ResponseBody string                 `json:"response_body,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// responseBodyWriter wraps gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logger creates a logger middleware with custom configuration
func Logger(cfg LoggerConfig) gin.HandlerFunc {
	skipPaths := utils.BuildStringSet(cfg.SkipPaths)

	return func(c *gin.Context) {
		// Skip logging for certain paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Use existing request ID or generate a new one
		// (RequestIDMiddleware may have already set this)
		requestID := GetRequestID(c)
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("request_id", requestID)
			c.Header("X-Request-ID", requestID)
		}

		// Start timer
		start := time.Now()

		// Capture request body if configured
		var requestBody string
		if cfg.LogRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) <= cfg.MaxBodySize {
				requestBody = string(bodyBytes)
				// Restore body for further reading
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Capture response body if configured
		var responseBody string
		var blw *responseBodyWriter
		if cfg.LogResponseBody {
			blw = &responseBodyWriter{
				ResponseWriter: c.Writer,
				body:           bytes.NewBufferString(""),
			}
			c.Writer = blw
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get response body if captured
		if cfg.LogResponseBody && blw != nil {
			responseBytes := blw.body.Bytes()
			if len(responseBytes) <= cfg.MaxBodySize {
				responseBody = blw.body.String()
			}
		}

		// Get error message if any
		var errorMessage string
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String()
		}

		// Get additional metadata from context
		metadata := make(map[string]interface{})
		if userID, exists := c.Get("user_id"); exists {
			metadata["user_id"] = userID
		}
		if userType, exists := c.Get("user_type"); exists {
			metadata["user_type"] = userType
		}

		// Create log entry
		entry := LogEntry{
			RequestID:    requestID,
			Timestamp:    time.Now().Format(time.RFC3339),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			Query:        c.Request.URL.RawQuery,
			StatusCode:   c.Writer.Status(),
			Latency:      latency.String(),
			ClientIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			ErrorMessage: errorMessage,
			RequestBody:  requestBody,
			ResponseBody: responseBody,
			Metadata:     metadata,
		}

		// Output log based on format
		switch cfg.Format {
		case JSONFormat:
			logJSON(entry)
		case TextFormat:
			logText(entry)
		}
	}
}

// DefaultLogger creates a logger middleware with default configuration
func DefaultLogger() gin.HandlerFunc {
	return Logger(DefaultLoggerConfig())
}

// logJSON outputs log in JSON format
func logJSON(entry LogEntry) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Error marshaling log entry: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// logText outputs log in text format
func logText(entry LogEntry) {
	var statusColor string
	switch {
	case entry.StatusCode >= 500:
		statusColor = "\033[1;31m" // Red
	case entry.StatusCode >= 400:
		statusColor = "\033[1;33m" // Yellow
	case entry.StatusCode >= 300:
		statusColor = "\033[1;36m" // Cyan
	case entry.StatusCode >= 200:
		statusColor = "\033[1;32m" // Green
	default:
		statusColor = "\033[1;37m" // White
	}
	resetColor := "\033[0m"

	output := fmt.Sprintf("[%s] %s%d%s | %s | %s | %s | %s",
		entry.Timestamp,
		statusColor,
		entry.StatusCode,
		resetColor,
		entry.Latency,
		entry.Method,
		entry.Path,
		entry.ClientIP,
	)

	if entry.ErrorMessage != "" {
		output += fmt.Sprintf(" | Error: %s", entry.ErrorMessage)
	}

	if entry.Query != "" {
		output += fmt.Sprintf(" | Query: %s", entry.Query)
	}

	fmt.Println(output)
}

// RequestIDMiddleware adds a request ID to the context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
