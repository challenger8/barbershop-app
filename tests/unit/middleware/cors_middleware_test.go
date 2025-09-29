package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"barber-booking-system/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDefaultCORS(t *testing.T) {
	router := gin.New()
	router.Use(middleware.DefaultCORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_SpecificOrigins(t *testing.T) {
	config := middleware.CORSConfig{
		AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	router := gin.New()
	router.Use(middleware.CORS(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		expectVary     bool
	}{
		{
			name:           "Allowed origin",
			origin:         "https://example.com",
			expectedOrigin: "https://example.com",
			expectVary:     true,
		},
		{
			name:           "Another allowed origin",
			origin:         "https://app.example.com",
			expectedOrigin: "https://app.example.com",
			expectVary:     true,
		},
		{
			name:           "Disallowed origin",
			origin:         "https://evil.com",
			expectedOrigin: "",
			expectVary:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.origin)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))

			if tt.expectVary {
				assert.Equal(t, "Origin", w.Header().Get("Vary"))
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
			}
		})
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	config := middleware.CORSConfig{
		AllowOrigins:     []string{"https://example.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	router := gin.New()
	router.Use(middleware.CORS(config))
	router.OPTIONS("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "options"})
	})
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "post"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_ExposeHeaders(t *testing.T) {
	config := middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"X-Request-ID", "X-Total-Count"},
		AllowCredentials: false,
		MaxAge:           3600,
	}

	router := gin.New()
	router.Use(middleware.CORS(config))
	router.GET("/test", func(c *gin.Context) {
		c.Header("X-Request-ID", "123")
		c.Header("X-Total-Count", "100")
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	exposeHeaders := w.Header().Get("Access-Control-Expose-Headers")
	assert.Contains(t, exposeHeaders, "X-Request-ID")
	assert.Contains(t, exposeHeaders, "X-Total-Count")
}

func TestCORS_WildcardSubdomain(t *testing.T) {
	config := middleware.CORSConfig{
		AllowOrigins:     []string{"https://*.example.com"},
		AllowMethods:     []string{http.MethodGet},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	router := gin.New()
	router.Use(middleware.CORS(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name        string
		origin      string
		shouldAllow bool
	}{
		{
			name:        "Matching subdomain",
			origin:      "https://api.example.com",
			shouldAllow: true,
		},
		{
			name:        "Another matching subdomain",
			origin:      "https://app.example.com",
			shouldAllow: true,
		},
		{
			name:        "Different domain",
			origin:      "https://api.other.com",
			shouldAllow: false,
		},
		{
			name:        "Base domain",
			origin:      "https://example.com",
			shouldAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.origin)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.shouldAllow {
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src")
}

func TestAllowedOriginsConfig(t *testing.T) {
	origins := []string{"https://example.com", "https://app.example.com"}
	config := middleware.AllowedOriginsConfig(origins...)

	assert.Equal(t, origins, config.AllowOrigins)
	assert.True(t, config.AllowCredentials)
	assert.NotEmpty(t, config.AllowMethods)
	assert.NotEmpty(t, config.AllowHeaders)
}

func TestDevelopmentCORSConfig(t *testing.T) {
	config := middleware.DevelopmentCORSConfig()

	assert.Equal(t, []string{"*"}, config.AllowOrigins)
	assert.False(t, config.AllowCredentials)
	assert.Contains(t, config.AllowMethods, http.MethodGet)
	assert.Contains(t, config.AllowMethods, http.MethodPost)
	assert.Equal(t, 86400, config.MaxAge)
}

func TestProductionCORSConfig(t *testing.T) {
	allowedOrigins := []string{"https://example.com", "https://app.example.com"}
	config := middleware.ProductionCORSConfig(allowedOrigins)

	assert.Equal(t, allowedOrigins, config.AllowOrigins)
	assert.True(t, config.AllowCredentials)
	assert.NotContains(t, config.AllowMethods, http.MethodHead)
	assert.Contains(t, config.AllowHeaders, "X-CSRF-Token")
	assert.Equal(t, 3600, config.MaxAge)
}

func TestDefaultCORSConfig(t *testing.T) {
	config := middleware.DefaultCORSConfig()

	assert.Equal(t, []string{"*"}, config.AllowOrigins)
	assert.False(t, config.AllowCredentials)
	assert.Contains(t, config.AllowMethods, http.MethodGet)
	assert.Contains(t, config.AllowMethods, http.MethodPost)
	assert.Contains(t, config.AllowHeaders, "Authorization")
	assert.Equal(t, 3600, config.MaxAge)
}

// Benchmark tests
func BenchmarkCORS(b *testing.B) {
	router := gin.New()
	router.Use(middleware.DefaultCORS())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkSecurityHeaders(b *testing.B) {
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
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
