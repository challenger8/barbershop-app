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

func TestDefaultRateLimit(t *testing.T) {
	router := gin.New()
	router.Use(middleware.DefaultRateLimit())
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

func TestRateLimit_Exceeded(t *testing.T) {
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
	router.Use(middleware.RateLimit(config))
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
	assert.NotEmpty(t, w3.Header().Get("Retry-After"))
}

func TestRateLimit_Reset(t *testing.T) {
	config := middleware.RateLimitConfig{
		Limit:      2,
		Window:     100 * time.Millisecond,
		KeyFunc:    middleware.IPKeyFunc,
		SkipPaths:  []string{},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()
	router.Use(middleware.RateLimit(config))
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

func TestRateLimit_SkipPaths(t *testing.T) {
	config := middleware.RateLimitConfig{
		Limit:      1,
		Window:     1 * time.Minute,
		KeyFunc:    middleware.IPKeyFunc,
		SkipPaths:  []string{"/health"},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()
	router.Use(middleware.RateLimit(config))
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

func TestRateLimit_DifferentIPs(t *testing.T) {
	config := middleware.RateLimitConfig{
		Limit:      2,
		Window:     1 * time.Minute,
		KeyFunc:    middleware.IPKeyFunc,
		SkipPaths:  []string{},
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	router := gin.New()
	router.Use(middleware.RateLimit(config))
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

func TestUserKeyFunc(t *testing.T) {
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

func TestUserKeyFunc_Unauthenticated(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		key := middleware.UserKeyFunc(c)
		// Should fall back to IP
		assert.Contains(t, key, "ip:")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPathKeyFunc(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		key := middleware.PathKeyFunc(c)
		assert.Contains(t, key, "path:/test")
		assert.Contains(t, key, "ip:")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStrictRateLimit(t *testing.T) {
	config := middleware.StrictRateLimit()

	assert.Equal(t, 20, config.Limit)
	assert.Equal(t, 1*time.Minute, config.Window)
}

func TestRelaxedRateLimit(t *testing.T) {
	config := middleware.RelaxedRateLimit()

	assert.Equal(t, 1000, config.Limit)
	assert.Equal(t, 1*time.Minute, config.Window)
}

func TestPerSecondRateLimit(t *testing.T) {
	config := middleware.PerSecondRateLimit(10)

	assert.Equal(t, 10, config.Limit)
	assert.Equal(t, 1*time.Second, config.Window)
}

func TestPerHourRateLimit(t *testing.T) {
	config := middleware.PerHourRateLimit(1000)

	assert.Equal(t, 1000, config.Limit)
	assert.Equal(t, 1*time.Hour, config.Window)
}

func TestPerDayRateLimit(t *testing.T) {
	config := middleware.PerDayRateLimit(10000)

	assert.Equal(t, 10000, config.Limit)
	assert.Equal(t, 24*time.Hour, config.Window)
}

func TestAuthenticatedRateLimit(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.AuthenticatedRateLimit(10, 2))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Unauthenticated requests (limit: 2)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Third unauthenticated request should fail
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestAuthenticatedRateLimit_AuthenticatedUser(t *testing.T) {
	router := gin.New()

	// Simulate authentication
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 123)
		c.Next()
	})

	router.Use(middleware.ErrorHandler())
	router.Use(middleware.AuthenticatedRateLimit(10, 2))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Authenticated user should have higher limit (10)
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 11th request should fail
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRouteSpecificRateLimiter(t *testing.T) {
	routeLimits := []middleware.RouteLimit{
		{Path: "/api/login", Limit: 5, Window: 1 * time.Minute},
		{Path: "/api/register", Limit: 3, Window: 1 * time.Minute},
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RouteSpecificRateLimiter(routeLimits, 100))

	router.POST("/api/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	router.POST("/api/register", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Test login endpoint (limit: 5)
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/login", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 6th login request should fail
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/login", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := middleware.DefaultRateLimitConfig()

	assert.Equal(t, 100, config.Limit)
	assert.Equal(t, 1*time.Minute, config.Window)
	assert.NotNil(t, config.KeyFunc)
	assert.Contains(t, config.SkipPaths, "/health")
	assert.Equal(t, http.StatusTooManyRequests, config.StatusCode)
}

// Benchmark tests
func BenchmarkRateLimit(b *testing.B) {
	router := gin.New()
	router.Use(middleware.DefaultRateLimit())
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

func BenchmarkRateLimit_Concurrent(b *testing.B) {
	router := gin.New()
	router.Use(middleware.DefaultRateLimit())
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
