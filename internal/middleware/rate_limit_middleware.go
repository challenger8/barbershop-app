// internal/middleware/rate_limit_middleware.go
package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig defines the configuration for rate limiting
type RateLimitConfig struct {
	Limit      int           // Maximum number of requests
	Window     time.Duration // Time window
	KeyFunc    KeyFunc       // Function to generate rate limit key
	SkipPaths  []string      // Paths to skip
	Message    string        // Custom message for rate limit exceeded
	StatusCode int           // Custom status code (default: 429)
}

// KeyFunc defines the function to generate rate limit key
type KeyFunc func(*gin.Context) string

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:      100,
		Window:     1 * time.Minute,
		KeyFunc:    IPKeyFunc,
		SkipPaths:  []string{"/health", "/metrics"},
		Message:    "Too many requests. Please try again later.",
		StatusCode: http.StatusTooManyRequests,
	}
}

// rateLimiter implements a simple in-memory rate limiter
type rateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*client
	config  RateLimitConfig
	stopCh  chan struct{} // Add this
}

// client represents a client's rate limit state
type client struct {
	count      int
	resetTime  time.Time
	lastAccess time.Time
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(config RateLimitConfig) *rateLimiter {
	rl := &rateLimiter{
		clients: make(map[string]*client),
		config:  config,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// allow checks if a request is allowed
func (rl *rateLimiter) allow(key string) (bool, int, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	c, exists := rl.clients[key]
	if !exists || now.After(c.resetTime) {
		// New client or window expired
		rl.clients[key] = &client{
			count:      1,
			resetTime:  now.Add(rl.config.Window),
			lastAccess: now,
		}
		return true, rl.config.Limit - 1, rl.config.Window
	}

	// Update last access
	c.lastAccess = now

	if c.count < rl.config.Limit {
		c.count++
		remaining := rl.config.Limit - c.count
		resetIn := c.resetTime.Sub(now)
		return true, remaining, resetIn
	}

	// Rate limit exceeded
	remaining := 0
	resetIn := c.resetTime.Sub(now)
	return false, remaining, resetIn
}

// cleanupLoop periodically removes stale clients
func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
			return
		}
	}
}
func (rl *rateLimiter) Stop() {
	close(rl.stopCh)
}

// cleanup removes stale clients
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	staleThreshold := 10 * time.Minute

	for key, c := range rl.clients {
		if now.Sub(c.lastAccess) > staleThreshold {
			delete(rl.clients, key)
		}
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	limiter := newRateLimiter(config)

	// Create skip paths map
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// Skip rate limiting for certain paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Generate key
		key := config.KeyFunc(c)

		// Check rate limit
		allowed, remaining, resetIn := limiter.allow(key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(resetIn).Unix()))

		if !allowed {
			// Rate limit exceeded
			c.Header("Retry-After", fmt.Sprintf("%d", int(resetIn.Seconds())))

			RespondWithError(c, &AppError{
				StatusCode: config.StatusCode,
				Code:       "RATE_LIMIT_EXCEEDED",
				Message:    config.Message,
				Details: map[string]interface{}{
					"retry_after": int(resetIn.Seconds()),
					"reset_at":    time.Now().Add(resetIn).Unix(),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// DefaultRateLimit creates a rate limiting middleware with default configuration
func DefaultRateLimit() gin.HandlerFunc {
	return RateLimit(DefaultRateLimitConfig())
}

// Key generation functions

// IPKeyFunc generates rate limit key based on IP address
func IPKeyFunc(c *gin.Context) string {
	return "ip:" + c.ClientIP()
}

// UserKeyFunc generates rate limit key based on user ID
func UserKeyFunc(c *gin.Context) string {
	userID, exists := GetUserID(c)
	if !exists {
		// Fall back to IP if user not authenticated
		return IPKeyFunc(c)
	}
	return fmt.Sprintf("user:%d", userID)
}

// PathKeyFunc generates rate limit key based on path and IP
func PathKeyFunc(c *gin.Context) string {
	return fmt.Sprintf("path:%s:ip:%s", c.Request.URL.Path, c.ClientIP())
}

// UserAndPathKeyFunc generates rate limit key based on user ID and path
func UserAndPathKeyFunc(c *gin.Context) string {
	userID, exists := GetUserID(c)
	if !exists {
		return PathKeyFunc(c)
	}
	return fmt.Sprintf("user:%d:path:%s", userID, c.Request.URL.Path)
}

// CustomKeyFunc allows creating custom key functions
func CustomKeyFunc(fn func(*gin.Context) string) KeyFunc {
	return fn
}

// Preset configurations

// StrictRateLimit returns a strict rate limit configuration
func StrictRateLimit() RateLimitConfig {
	config := DefaultRateLimitConfig()
	config.Limit = 20
	config.Window = 1 * time.Minute
	return config
}

// RelaxedRateLimit returns a relaxed rate limit configuration
func RelaxedRateLimit() RateLimitConfig {
	config := DefaultRateLimitConfig()
	config.Limit = 1000
	config.Window = 1 * time.Minute
	return config
}

// PerSecondRateLimit returns a per-second rate limit configuration
func PerSecondRateLimit(limit int) RateLimitConfig {
	config := DefaultRateLimitConfig()
	config.Limit = limit
	config.Window = 1 * time.Second
	return config
}

// PerHourRateLimit returns a per-hour rate limit configuration
func PerHourRateLimit(limit int) RateLimitConfig {
	config := DefaultRateLimitConfig()
	config.Limit = limit
	config.Window = 1 * time.Hour
	return config
}

// PerDayRateLimit returns a per-day rate limit configuration
func PerDayRateLimit(limit int) RateLimitConfig {
	config := DefaultRateLimitConfig()
	config.Limit = limit
	config.Window = 24 * time.Hour
	return config
}

// AuthenticatedRateLimit returns different limits for authenticated vs unauthenticated users
func AuthenticatedRateLimit(authenticatedLimit, unauthenticatedLimit int) gin.HandlerFunc {
	authenticatedConfig := RateLimitConfig{
		Limit:      authenticatedLimit,
		Window:     1 * time.Minute,
		KeyFunc:    UserKeyFunc,
		SkipPaths:  []string{"/health", "/metrics"},
		Message:    "Too many requests. Please try again later.",
		StatusCode: http.StatusTooManyRequests,
	}

	unauthenticatedConfig := RateLimitConfig{
		Limit:      unauthenticatedLimit,
		Window:     1 * time.Minute,
		KeyFunc:    IPKeyFunc,
		SkipPaths:  []string{"/health", "/metrics"},
		Message:    "Too many requests. Please try again later.",
		StatusCode: http.StatusTooManyRequests,
	}

	authenticatedLimiter := newRateLimiter(authenticatedConfig)
	unauthenticatedLimiter := newRateLimiter(unauthenticatedConfig)

	return func(c *gin.Context) {
		// Skip rate limiting for certain paths
		skipPaths := map[string]bool{
			"/health":  true,
			"/metrics": true,
		}
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Check if user is authenticated
		isAuth := IsAuthenticated(c)

		var limiter *rateLimiter
		var config RateLimitConfig
		var key string

		if isAuth {
			limiter = authenticatedLimiter
			config = authenticatedConfig
			key = UserKeyFunc(c)
		} else {
			limiter = unauthenticatedLimiter
			config = unauthenticatedConfig
			key = IPKeyFunc(c)
		}

		// Check rate limit
		allowed, remaining, resetIn := limiter.allow(key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(resetIn).Unix()))

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(resetIn.Seconds())))

			RespondWithError(c, &AppError{
				StatusCode: config.StatusCode,
				Code:       "RATE_LIMIT_EXCEEDED",
				Message:    config.Message,
				Details: map[string]interface{}{
					"retry_after": int(resetIn.Seconds()),
					"reset_at":    time.Now().Add(resetIn).Unix(),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RouteSpecificRateLimit applies different limits to different routes
type RouteLimit struct {
	Path   string
	Limit  int
	Window time.Duration
}

// RouteSpecificRateLimiter creates rate limiters for specific routes
func RouteSpecificRateLimiter(routeLimits []RouteLimit, defaultLimit int) gin.HandlerFunc {
	limiters := make(map[string]*rateLimiter)
	configs := make(map[string]RateLimitConfig)

	// Create limiters for each route
	for _, rl := range routeLimits {
		config := RateLimitConfig{
			Limit:      rl.Limit,
			Window:     rl.Window,
			KeyFunc:    IPKeyFunc,
			Message:    "Too many requests. Please try again later.",
			StatusCode: http.StatusTooManyRequests,
		}
		limiters[rl.Path] = newRateLimiter(config)
		configs[rl.Path] = config
	}

	// Create default limiter
	defaultConfig := RateLimitConfig{
		Limit:      defaultLimit,
		Window:     1 * time.Minute,
		KeyFunc:    IPKeyFunc,
		Message:    "Too many requests. Please try again later.",
		StatusCode: http.StatusTooManyRequests,
	}
	defaultLimiter := newRateLimiter(defaultConfig)

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Get limiter for this path
		limiter, exists := limiters[path]
		config := defaultConfig
		if exists {
			config = configs[path]
		} else {
			limiter = defaultLimiter
		}

		// Generate key
		key := config.KeyFunc(c)

		// Check rate limit
		allowed, remaining, resetIn := limiter.allow(key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(resetIn).Unix()))

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(resetIn.Seconds())))

			RespondWithError(c, &AppError{
				StatusCode: config.StatusCode,
				Code:       "RATE_LIMIT_EXCEEDED",
				Message:    config.Message,
				Details: map[string]interface{}{
					"retry_after": int(resetIn.Seconds()),
					"reset_at":    time.Now().Add(resetIn).Unix(),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
