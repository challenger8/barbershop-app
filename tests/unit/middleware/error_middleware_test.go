// tests/unit/middleware/error_middleware_test.go
package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"barber-booking-system/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestErrorHandler_AppError(t *testing.T) {
	tests := []struct {
		name           string
		appError       *middleware.AppError
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Bad Request Error",
			appError:       middleware.NewBadRequestError("Invalid input", nil),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Unauthorized Error",
			appError:       middleware.NewUnauthorizedError("Please login"),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "Forbidden Error",
			appError:       middleware.NewForbiddenError("Access denied"),
			expectedStatus: http.StatusForbidden,
			expectedCode:   "FORBIDDEN",
		},
		{
			name:           "Not Found Error",
			appError:       middleware.NewNotFoundError("Resource not found"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "Conflict Error",
			appError:       middleware.NewConflictError("Resource already exists", nil),
			expectedStatus: http.StatusConflict,
			expectedCode:   "CONFLICT",
		},
		{
			name:           "Internal Server Error",
			appError:       middleware.NewInternalServerError("Database error", errors.New("connection failed")),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_SERVER_ERROR",
		},
		{
			name:           "Validation Error",
			appError:       middleware.NewValidationError("Invalid fields", map[string]interface{}{"email": "invalid format"}),
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			router := gin.New()
			router.Use(middleware.ErrorHandler())
			router.GET("/test", func(c *gin.Context) {
				_ = c.Error(tt.appError)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response middleware.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCode, response.Code)
			assert.Equal(t, tt.appError.Message, response.Message)
		})
	}
}

func TestErrorHandler_WithDetails(t *testing.T) {
	w := httptest.NewRecorder()

	details := map[string]interface{}{
		"field":  "email",
		"reason": "invalid format",
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		err := middleware.NewValidationError("Validation failed", details)
		_ = c.Error(err)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response middleware.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.NotNil(t, response.Details)
	assert.Equal(t, "email", response.Details["field"])
}

func TestErrorHandler_GenericError(t *testing.T) {
	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		_ = c.Error(errors.New("generic error"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response middleware.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(t, "An unexpected error occurred", response.Message)
}

func TestErrorHandler_NoError(t *testing.T) {
	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecoveryHandler(t *testing.T) {
	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(middleware.RecoveryHandler())
	router.GET("/test", func(c *gin.Context) {
		panic("something went wrong")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response middleware.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(t, "An unexpected error occurred", response.Message)
}

func TestAbortWithError(t *testing.T) {
	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		middleware.AbortWithError(c, middleware.NewUnauthorizedError("Access denied"))
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRespondWithError(t *testing.T) {
	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		middleware.RespondWithError(c, middleware.NewNotFoundError("User not found"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response middleware.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "NOT_FOUND", response.Code)
	assert.Equal(t, "User not found", response.Message)
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *middleware.AppError
		expected string
	}{
		{
			name:     "Error with wrapped error",
			appError: middleware.NewInternalServerError("Database error", errors.New("connection failed")),
			expected: "connection failed",
		},
		{
			name:     "Error without wrapped error",
			appError: middleware.NewBadRequestError("Invalid input", nil),
			expected: "Invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.appError.Error())
		})
	}
}

// Benchmark tests
func BenchmarkErrorHandler(b *testing.B) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		_ = c.Error(middleware.NewBadRequestError("Invalid input", nil))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRecoveryHandler(b *testing.B) {
	router := gin.New()
	router.Use(middleware.RecoveryHandler())
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
