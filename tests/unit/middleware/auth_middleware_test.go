// tests/unit/middleware/auth_middleware_test.go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"barber-booking-system/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecretKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsInVzZXJfdHlwZSI6ImN1c3RvbWVyIiwiZXhwIjoxNzU5MjQwNzAwLCJuYmYiOjE3NTkxNTQzMDAsImlhdCI6MTc1OTE1NDMwMH0.kHd-nNJQD8JiZEa7zuSOCagz5_uCK08CcNcfTeOeg2w"

func TestGenerateToken(t *testing.T) {
	token, err := middleware.GenerateToken(123, "test@example.com", "customer", testSecretKey, 24*time.Hour)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Generate valid token
	token, err := middleware.GenerateToken(123, "test@example.com", "customer", testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	router := gin.New()
	router.Use(middleware.RequireAuth(testSecretKey))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := middleware.GetUserID(c)
		email, _ := middleware.GetUserEmail(c)
		userType, _ := middleware.GetUserType(c)

		c.JSON(http.StatusOK, gin.H{
			"user_id":   userID,
			"email":     email,
			"user_type": userType,
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequireAuth(testSecretKey))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	// No Authorization header

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequireAuth(testSecretKey))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// Generate expired token
	token, err := middleware.GenerateToken(123, "test@example.com", "customer", testSecretKey, -1*time.Hour)
	require.NoError(t, err)

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequireAuth(testSecretKey))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_SkipPaths(t *testing.T) {
	config := middleware.DefaultAuthConfig(testSecretKey)

	router := gin.New()
	router.Use(middleware.AuthMiddleware(config))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "login"})
	})

	tests := []struct {
		name string
		path string
	}{
		{"Health check", "/health"},
		{"Login", "/api/v1/auth/login"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)
			// No token

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestRequireRole_AdminOnly(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequireAdmin(testSecretKey))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	tests := []struct {
		name           string
		userType       string
		expectedStatus int
	}{
		{"Admin user", "admin", http.StatusOK},
		{"Barber user", "barber", http.StatusForbidden},
		{"Customer user", "customer", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := middleware.GenerateToken(123, "test@example.com", tt.userType, testSecretKey, 24*time.Hour)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRequireBarber(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequireBarber(testSecretKey))
	router.GET("/barber", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	token, err := middleware.GenerateToken(123, "barber@example.com", "barber", testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/barber", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireBarberOrAdmin(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequireBarberOrAdmin(testSecretKey))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	tests := []struct {
		name           string
		userType       string
		expectedStatus int
	}{
		{"Admin allowed", "admin", http.StatusOK},
		{"Barber allowed", "barber", http.StatusOK},
		{"Customer denied", "customer", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := middleware.GenerateToken(123, "test@example.com", tt.userType, testSecretKey, 24*time.Hour)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestOptionalAuth_WithToken(t *testing.T) {
	router := gin.New()
	router.Use(middleware.OptionalAuth(testSecretKey))
	router.GET("/test", func(c *gin.Context) {
		authenticated := middleware.IsAuthenticated(c)
		userID, _ := middleware.GetUserID(c)

		c.JSON(http.StatusOK, gin.H{
			"authenticated": authenticated,
			"user_id":       userID,
		})
	})

	token, err := middleware.GenerateToken(123, "test@example.com", "customer", testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuth_WithoutToken(t *testing.T) {
	router := gin.New()
	router.Use(middleware.OptionalAuth(testSecretKey))
	router.GET("/test", func(c *gin.Context) {
		authenticated := middleware.IsAuthenticated(c)

		c.JSON(http.StatusOK, gin.H{
			"authenticated": authenticated,
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	// No token

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRefreshToken(t *testing.T) {
	// Generate original token
	originalToken, err := middleware.GenerateToken(123, "test@example.com", "customer", testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Refresh token
	newToken, err := middleware.RefreshToken(originalToken, testSecretKey, 24*time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, originalToken, newToken)
}

func TestGetUserHelpers(t *testing.T) {
	router := gin.New()
	router.Use(middleware.RequireAuth(testSecretKey))
	router.GET("/test", func(c *gin.Context) {
		userID := middleware.MustGetUserID(c)
		email, _ := middleware.GetUserEmail(c)
		userType, _ := middleware.GetUserType(c)
		claims, _ := middleware.GetClaims(c)

		assert.Equal(t, 123, userID)
		assert.Equal(t, "test@example.com", email)
		assert.Equal(t, "customer", userType)
		assert.NotNil(t, claims)

		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	token, err := middleware.GenerateToken(123, "test@example.com", "customer", testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIsAdmin(t *testing.T) {
	router := gin.New()
	router.Use(middleware.OptionalAuth(testSecretKey))
	router.GET("/test", func(c *gin.Context) {
		isAdmin := middleware.IsAdmin(c)
		isBarber := middleware.IsBarber(c)
		isCustomer := middleware.IsCustomer(c)

		c.JSON(http.StatusOK, gin.H{
			"is_admin":    isAdmin,
			"is_barber":   isBarber,
			"is_customer": isCustomer,
		})
	})

	token, err := middleware.GenerateToken(123, "admin@example.com", "admin", testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTokenFromCookie(t *testing.T) {
	config := middleware.AuthConfig{
		SecretKey:     testSecretKey,
		TokenLookup:   "cookie:auth_token",
		TokenHeadName: "",
		SkipPaths:     []string{},
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.AuthMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	token, err := middleware.GenerateToken(123, "test@example.com", "customer", testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDefaultAuthConfig(t *testing.T) {
	config := middleware.DefaultAuthConfig(testSecretKey)

	assert.Equal(t, testSecretKey, config.SecretKey)
	assert.Equal(t, "header:Authorization", config.TokenLookup)
	assert.Equal(t, "Bearer", config.TokenHeadName)
	assert.Contains(t, config.SkipPaths, "/health")
	assert.Contains(t, config.SkipPaths, "/api/v1/auth/login")
}

// Benchmark tests
func BenchmarkAuthMiddleware(b *testing.B) {
	token, _ := middleware.GenerateToken(123, "test@example.com", "customer", testSecretKey, 24*time.Hour)

	router := gin.New()
	router.Use(middleware.RequireAuth(testSecretKey))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = middleware.GenerateToken(123, "test@example.com", "customer", testSecretKey, 24*time.Hour)
	}
}
