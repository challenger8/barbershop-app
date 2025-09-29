// internal/middleware/cors_middleware.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig defines the configuration for CORS middleware
type CORSConfig struct {
	AllowOrigins     []string // List of allowed origins
	AllowMethods     []string // List of allowed HTTP methods
	AllowHeaders     []string // List of allowed headers
	ExposeHeaders    []string // List of headers exposed to the client
	AllowCredentials bool     // Allow credentials (cookies, authorization headers)
	MaxAge           int      // Preflight cache duration in seconds
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			"X-Total-Count",
		},
		AllowCredentials: false,
		MaxAge:           3600, // 1 hour
	}
}

// CORS creates a CORS middleware with custom configuration
func CORS(config CORSConfig) gin.HandlerFunc {
	// Precompute joined strings for better performance
	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")
	exposeHeaders := ""
	if len(config.ExposeHeaders) > 0 {
		exposeHeaders = strings.Join(config.ExposeHeaders, ", ")
	}

	// Create origin map for fast lookup
	allowAllOrigins := false
	allowedOriginsMap := make(map[string]bool)

	for _, origin := range config.AllowOrigins {
		if origin == "*" {
			allowAllOrigins = true
			break
		}
		allowedOriginsMap[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Set Access-Control-Allow-Origin
		if allowAllOrigins {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && allowedOriginsMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else if origin != "" {
			// Check for wildcard subdomain matching
			for allowedOrigin := range allowedOriginsMap {
				if matchesWildcard(origin, allowedOrigin) {
					c.Header("Access-Control-Allow-Origin", origin)
					c.Header("Vary", "Origin")
					break
				}
			}
		}

		// Set Access-Control-Allow-Credentials
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			// Set preflight headers
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)

			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}

			// Respond to preflight request
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Set expose headers for actual requests
		if exposeHeaders != "" {
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
		}

		c.Next()
	}
}

// DefaultCORS creates a CORS middleware with default configuration
func DefaultCORS() gin.HandlerFunc {
	return CORS(DefaultCORSConfig())
}

// matchesWildcard checks if origin matches a wildcard pattern
// Example: "https://api.example.com" matches "https://*.example.com"
func matchesWildcard(origin, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return origin == pattern
	}

	// Split by wildcard
	parts := strings.Split(pattern, "*")
	if len(parts) != 2 {
		return false
	}

	prefix, suffix := parts[0], parts[1]

	// Check if origin starts with prefix and ends with suffix
	return strings.HasPrefix(origin, prefix) && strings.HasSuffix(origin, suffix) &&
		len(origin) >= len(prefix)+len(suffix)
}

// SecurityHeaders adds common security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Strict Transport Security (HTTPS only)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// AllowedOriginsConfig creates a CORS config with specific origins
func AllowedOriginsConfig(origins ...string) CORSConfig {
	config := DefaultCORSConfig()
	config.AllowOrigins = origins
	config.AllowCredentials = true
	return config
}

// DevelopmentCORSConfig creates a permissive CORS config for development
func DevelopmentCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// ProductionCORSConfig creates a strict CORS config for production
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
			"X-CSRF-Token",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			"X-Total-Count",
			"X-Page",
			"X-Per-Page",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}
}
