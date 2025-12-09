// tests/unit/middleware/logging_middleware_test.go
package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"barber-booking-system/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultLogger(t *testing.T) {
	router := gin.New()
	router.Use(middleware.DefaultLogger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestLogger_WithCustomConfig(t *testing.T) {
	config := middleware.LoggerConfig{
		Format:          middleware.JSONFormat,
		SkipPaths:       []string{"/health"},
		LogRequestBody:  true,
		LogResponseBody: true,
		MaxBodySize:     1024,
	}

	router := gin.New()
	router.Use(middleware.Logger(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestLogger_SkipPaths(t *testing.T) {
	config := middleware.LoggerConfig{
		Format:    middleware.JSONFormat,
		SkipPaths: []string{"/health", "/metrics"},
	}

	router := gin.New()
	router.Use(middleware.Logger(config))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name            string
		path            string
		expectRequestID bool
	}{
		{
			name:            "Skip health endpoint",
			path:            "/health",
			expectRequestID: false,
		},
		{
			name:            "Log regular endpoint",
			path:            "/test",
			expectRequestID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			requestID := w.Header().Get("X-Request-ID")
			if tt.expectRequestID {
				assert.NotEmpty(t, requestID)
			} else {
				assert.Empty(t, requestID)
			}
		})
	}
}

func TestLogger_WithRequestBody(t *testing.T) {
	config := middleware.LoggerConfig{
		Format:         middleware.JSONFormat,
		LogRequestBody: true,
		MaxBodySize:    1024,
	}

	router := gin.New()
	router.Use(middleware.Logger(config))
	router.POST("/test", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"received": body})
	})

	requestBody := `{"name":"test","email":"test@example.com"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response["received"])
}

func TestLogger_WithError(t *testing.T) {
	router := gin.New()
	router.Use(middleware.DefaultLogger())
	router.Use(middleware.ErrorHandler())

	router.GET("/error", func(c *gin.Context) {
		_ = c.Error(middleware.NewBadRequestError("Invalid input", nil))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestLogger_StatusCodeColors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		handler    gin.HandlerFunc
	}{
		{
			name:       "200 OK",
			statusCode: http.StatusOK,
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			},
		},
		{
			name:       "400 Bad Request",
			statusCode: http.StatusBadRequest,
			handler: func(c *gin.Context) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			},
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			handler: func(c *gin.Context) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.DefaultLogger())
			router.GET("/test", tt.handler)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	var capturedRequestID string
	router.GET("/test", func(c *gin.Context) {
		capturedRequestID = middleware.GetRequestID(c)
		c.JSON(http.StatusOK, gin.H{"request_id": capturedRequestID})
	})

	t.Run("Generate new request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, capturedRequestID)
		assert.Equal(t, capturedRequestID, w.Header().Get("X-Request-ID"))
	})

	t.Run("Use existing request ID", func(t *testing.T) {
		existingID := "test-request-id-123"
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", existingID)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, existingID, capturedRequestID)
		assert.Equal(t, existingID, w.Header().Get("X-Request-ID"))
	})
}

func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(*gin.Context)
		expectedID string
	}{
		{
			name: "Request ID exists",
			setupFunc: func(c *gin.Context) {
				c.Set("request_id", "test-id-123")
			},
			expectedID: "test-id-123",
		},
		{
			name: "Request ID does not exist",
			setupFunc: func(c *gin.Context) {
				// Don't set request_id
			},
			expectedID: "",
		},
		{
			name: "Request ID is wrong type",
			setupFunc: func(c *gin.Context) {
				c.Set("request_id", 12345)
			},
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupFunc(c)

			requestID := middleware.GetRequestID(c)
			assert.Equal(t, tt.expectedID, requestID)
		})
	}
}

func TestLogger_WithMetadata(t *testing.T) {
	router := gin.New()
	router.Use(middleware.DefaultLogger())
	router.GET("/test", func(c *gin.Context) {
		c.Set("user_id", 123)
		c.Set("user_type", "customer")
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogger_WithQueryParams(t *testing.T) {
	router := gin.New()
	router.Use(middleware.DefaultLogger())
	router.GET("/search", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "search results"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?q=test&limit=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestLogger_LargeRequestBody(t *testing.T) {
	config := middleware.LoggerConfig{
		Format:         middleware.JSONFormat,
		LogRequestBody: true,
		MaxBodySize:    10,
	}

	router := gin.New()
	router.Use(middleware.Logger(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "received"})
	})

	largeBody := `{"data":"this is a very large request body that exceeds the limit"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(largeBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDefaultLoggerConfig(t *testing.T) {
	config := middleware.DefaultLoggerConfig()

	assert.Equal(t, middleware.JSONFormat, config.Format)
	assert.Contains(t, config.SkipPaths, "/health")
	assert.Contains(t, config.SkipPaths, "/metrics")
	assert.False(t, config.LogRequestBody)
	assert.False(t, config.LogResponseBody)
	assert.Equal(t, 1024, config.MaxBodySize)
}

// Benchmark tests
func BenchmarkLogger(b *testing.B) {
	router := gin.New()
	router.Use(middleware.DefaultLogger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRequestIDMiddleware(b *testing.B) {
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
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
