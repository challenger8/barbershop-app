// internal/middleware/rate_limit_middleware.go
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/config"
	"barber-booking-system/internal/utils"

	"github.com/gin-gonic/gin"
)

// ============================================
// Redis-based Rate Limiter (Distributed)
// ============================================

// RateLimiter implements rate limiting using Redis
type RateLimiter struct {
	redis       *cache.RedisClient
	maxRequests int
	window      time.Duration
}

// NewRateLimiter creates a new Redis-based rate limiter
func NewRateLimiter(redis *cache.RedisClient, maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:       redis,
		maxRequests: maxRequests,
		window:      window,
	}
}

// Middleware creates rate limiting middleware using Redis
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP or user ID)
		identifier := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		key := fmt.Sprintf("ratelimit:%s", identifier)
		ctx := context.Background()

		// Get current count
		count, err := rl.redis.Increment(ctx, key)
		if err != nil {
			// If Redis fails, allow the request (fail open)
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			rl.redis.Expire(ctx, key, rl.window)
		}

		// Check if rate limit exceeded
		if count > int64(rl.maxRequests) {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rl.window).Unix()))

			RespondWithError(c, &AppError{
				Code:       "RATE_LIMIT_EXCEEDED",
				Message:    "Too many requests. Please try again later.",
				StatusCode: http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		remaining := rl.maxRequests - int(count)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// ============================================
// In-Memory Rate Limiter (Fallback)
// ============================================

// RateLimitConfig defines the configuration for in-memory rate limiting
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
		SkipPaths:  config.DefaultSkipPaths,
		Message:    "Too many requests. Please try again later.",
		StatusCode: http.StatusTooManyRequests,
	}
}

// StrictRateLimitConfig returns strict rate limit configuration for production
func StrictRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:      50,
		Window:     1 * time.Minute,
		KeyFunc:    IPKeyFunc,
		SkipPaths:  config.DefaultSkipPaths,
		Message:    "Too many requests. Please try again later.",
		StatusCode: http.StatusTooManyRequests,
	}
}

// AuthRateLimitConfig returns strict rate limit for authentication endpoints
// This prevents brute force attacks on login/register endpoints
func AuthRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:      10,              // Only 10 attempts per window
		Window:     5 * time.Minute, // 5 minute window
		KeyFunc:    IPKeyFunc,       // Rate limit by IP
		SkipPaths:  []string{},      // No skip paths for auth
		Message:    "Too many authentication attempts. Please try again in a few minutes.",
		StatusCode: http.StatusTooManyRequests,
	}
}

// inMemoryRateLimiter implements a simple in-memory rate limiter
type inMemoryRateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*client
	config  RateLimitConfig
	stopCh  chan struct{}
}

// client represents a client's rate limit state
type client struct {
	count      int
	resetTime  time.Time
	lastAccess time.Time
}

// newInMemoryRateLimiter creates a new in-memory rate limiter
func newInMemoryRateLimiter(config RateLimitConfig) *inMemoryRateLimiter {
	rl := &inMemoryRateLimiter{
		clients: make(map[string]*client),
		config:  config,
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// allow checks if a request is allowed
func (rl *inMemoryRateLimiter) allow(key string) (bool, int, time.Duration) {
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
func (rl *inMemoryRateLimiter) cleanupLoop() {
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

// Stop stops the cleanup goroutine
func (rl *inMemoryRateLimiter) Stop() {
	close(rl.stopCh)
}

// cleanup removes stale clients
func (rl *inMemoryRateLimiter) cleanup() {
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

// RateLimitMiddleware creates an in-memory rate limiting middleware
func RateLimitMiddleware(cfg RateLimitConfig) gin.HandlerFunc {
	limiter := newInMemoryRateLimiter(cfg)

	// Create skip paths map for O(1) lookup
	skipPaths := utils.BuildStringSet(cfg.SkipPaths)

	return func(c *gin.Context) {
		// Skip rate limiting for certain paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Get rate limit key
		key := cfg.KeyFunc(c)

		// Check if request is allowed
		allowed, remaining, resetIn := limiter.allow(key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(resetIn).Unix()))

		if !allowed {
			RespondWithError(c, &AppError{
				Code:       "RATE_LIMIT_EXCEEDED",
				Message:    cfg.Message,
				StatusCode: cfg.StatusCode,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPKeyFunc generates rate limit key based on client IP
func IPKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// UserKeyFunc generates rate limit key based on user ID (if authenticated)
func UserKeyFunc(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}
	return c.ClientIP()
}
