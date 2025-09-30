// tests/unit/middleware/rate_limit_middleware_test.go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"barber-booking-system/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Test DefaultRateLimitConfig
func TestDefaultRateLimitConfig(t *testing.T) {
	config := middleware.DefaultRateLimitConfig()

	assert.Equal(t, 100, config.Limit)
	assert.Equal(t, 1*time.Minute, config.Window)
	assert.NotNil(t, config.KeyFunc)
	assert.Contains(t, config.SkipPaths, "/health")
	assert.Equal(t, http.StatusTooManyRequests, config.StatusCode)
}

// Test StrictRateLimitConfig
func TestStrictRateLimitConfig(t *testing.T) {
	config := middleware.StrictRateLimitConfig()

	assert.Equal(t, 50, config.Limit)
	assert.Equal(t, 1*time.Minute, config.Window)
}

// Test RateLimitMiddleware with default config
func TestRateLimitMiddleware_DefaultConfig(t *testing.T) {
	router := gin.New()
	router.Use(middleware.RateLimitMiddleware(middleware.DefaultRateLimitConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// First request should succeed
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

// Test rate limit exceeded
func TestRateLimitMiddleware_Exceeded(t *testing.T) {
	config := middleware.RateLimitConfig{
		Limit:      2,
		Window:     1 * time.Minute,
		KeyFunc:    middleware.IPKeyFunc,
		SkipPaths:  []string{},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// First request - should succeed
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request - should succeed
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Third request - should fail (rate limit exceeded)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}

// Test rate limit reset after window expires
func TestRateLimitMiddleware_Reset(t *testing.T) {
	config := middleware.RateLimitConfig{
		Limit:      2,
		Window:     100 * time.Millisecond,
		KeyFunc:    middleware.IPKeyFunc,
		SkipPaths:  []string{},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Use up the limit
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Next request should fail
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)

	// Now should succeed again
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test skip paths
func TestRateLimitMiddleware_SkipPaths(t *testing.T) {
	config := middleware.RateLimitConfig{
		Limit:      1,
		Window:     1 * time.Minute,
		KeyFunc:    middleware.IPKeyFunc,
		SkipPaths:  []string{"/health"},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()
	router.Use(middleware.RateLimitMiddleware(config))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Multiple requests to /health should all succeed
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

// Test different IPs have separate limits
func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	config := middleware.RateLimitConfig{
		Limit:      2,
		Window:     1 * time.Minute,
		KeyFunc:    middleware.IPKeyFunc,
		SkipPaths:  []string{},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Requests from different IPs should have separate limits
	ips := []string{"192.168.1.1", "192.168.1.2"}

	for _, ip := range ips {
		// Each IP should be able to make 2 requests
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", ip)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	}
}

// Test IPKeyFunc
func TestIPKeyFunc(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		key := middleware.IPKeyFunc(c)
		assert.NotEmpty(t, key)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test UserKeyFunc with authenticated user
func TestUserKeyFunc_Authenticated(t *testing.T) {
	router := gin.New()

	// Simulate authenticated user
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 123)
		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		key := middleware.UserKeyFunc(c)
		assert.Equal(t, "user:123", key)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test UserKeyFunc falls back to IP for unauthenticated users
func TestUserKeyFunc_Unauthenticated(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		key := middleware.UserKeyFunc(c)
		// Should fall back to IP
		assert.NotEmpty(t, key)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test rate limit headers are set correctly
func TestRateLimitMiddleware_Headers(t *testing.T) {
	config := middleware.RateLimitConfig{
		Limit:      5,
		Window:     1 * time.Minute,
		KeyFunc:    middleware.IPKeyFunc,
		SkipPaths:  []string{},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()
	router.Use(middleware.RateLimitMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make first request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "4", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

// Test rate limit with user-based key function
func TestRateLimitMiddleware_UserBased(t *testing.T) {
	config := middleware.RateLimitConfig{
		Limit:      3,
		Window:     1 * time.Minute,
		KeyFunc:    middleware.UserKeyFunc,
		SkipPaths:  []string{},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()

	// Simulate authentication
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 456)
		c.Next()
	})

	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// User should be able to make 3 requests
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 4th request should fail
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

// Benchmark tests
func BenchmarkRateLimitMiddleware(b *testing.B) {
	router := gin.New()
	router.Use(middleware.RateLimitMiddleware(middleware.DefaultRateLimitConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRateLimitMiddleware_Concurrent(b *testing.B) {
	router := gin.New()
	router.Use(middleware.RateLimitMiddleware(middleware.DefaultRateLimitConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
		}
	})
}

// Test concurrent requests don't cause race conditions
func TestRateLimitMiddleware_Concurrent(t *testing.T) {
	config := middleware.RateLimitConfig{
		Limit:      100,
		Window:     1 * time.Minute,
		KeyFunc:    middleware.IPKeyFunc,
		SkipPaths:  []string{},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()
	router.Use(middleware.RateLimitMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Run 50 concurrent requests
	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 50; i++ {
		<-done
	}
}

// Test custom key function
func TestRateLimitMiddleware_CustomKeyFunc(t *testing.T) {
	customKeyFunc := func(c *gin.Context) string {
		return "custom-key"
	}

	config := middleware.RateLimitConfig{
		Limit:      2,
		Window:     1 * time.Minute,
		KeyFunc:    customKeyFunc,
		SkipPaths:  []string{},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// All requests will use same key, so limit applies globally
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 3rd request should fail
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
