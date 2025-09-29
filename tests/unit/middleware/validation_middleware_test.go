// tests/unit/middleware/validation_middleware_test.go
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

// Test models
type TestUser struct {
	Name  string `json:"name" binding:"required,min=3,max=50"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"required,gte=18,lte=100"`
}

type TestQuery struct {
	Page  int    `form:"page" binding:"required,gte=1"`
	Limit int    `form:"limit" binding:"required,gte=1,lte=100"`
	Sort  string `form:"sort" binding:"oneof=asc desc"`
}

type TestURI struct {
	ID int `uri:"id" binding:"required,gt=0"`
}

func TestValidateJSON_Success(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.POST("/test", middleware.ValidateJSON(TestUser{}), func(c *gin.Context) {
		validated, exists := middleware.GetValidatedBody(c)
		assert.True(t, exists)
		assert.NotNil(t, validated)

		user := validated.(*TestUser)
		c.JSON(http.StatusOK, gin.H{
			"name":  user.Name,
			"email": user.Email,
			"age":   user.Age,
		})
	})

	body := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", response["name"])
	assert.Equal(t, "john@example.com", response["email"])
}

func TestValidateJSON_RequiredFieldMissing(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.POST("/test", middleware.ValidateJSON(TestUser{}), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	body := map[string]interface{}{
		"email": "john@example.com",
		"age":   25,
		// name is missing
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response middleware.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.NotNil(t, response.Details)
}

func TestValidateJSON_InvalidEmail(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.POST("/test", middleware.ValidateJSON(TestUser{}), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	body := TestUser{
		Name:  "John Doe",
		Email: "invalid-email",
		Age:   25,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestValidateJSON_MinLengthViolation(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.POST("/test", middleware.ValidateJSON(TestUser{}), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	body := TestUser{
		Name:  "Jo", // Too short (min=3)
		Email: "john@example.com",
		Age:   25,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestValidateJSON_AgeRangeViolation(t *testing.T) {
	tests := []struct {
		name string
		age  int
	}{
		{"Age too low", 17},
		{"Age too high", 101},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.ErrorHandler())
			router.POST("/test", middleware.ValidateJSON(TestUser{}), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"success": true})
			})

			body := TestUser{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   tt.age,
			}
			jsonBody, _ := json.Marshal(body)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		})
	}
}

func TestValidateJSON_InvalidJSON(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.POST("/test", middleware.ValidateJSON(TestUser{}), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateQuery_Success(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/test", middleware.ValidateQuery(TestQuery{}), func(c *gin.Context) {
		validated, exists := middleware.GetValidatedQuery(c)
		assert.True(t, exists)

		query := validated.(*TestQuery)
		c.JSON(http.StatusOK, gin.H{
			"page":  query.Page,
			"limit": query.Limit,
			"sort":  query.Sort,
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=1&limit=10&sort=asc", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateQuery_InvalidParameters(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/test", middleware.ValidateQuery(TestQuery{}), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=0&limit=200&sort=invalid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestValidateURI_Success(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/test/:id", middleware.ValidateURI(TestURI{}), func(c *gin.Context) {
		validated, exists := middleware.GetValidatedURI(c)
		assert.True(t, exists)

		uri := validated.(*TestURI)
		c.JSON(http.StatusOK, gin.H{"id": uri.ID})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test/123", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateURI_InvalidID(t *testing.T) {
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/test/:id", middleware.ValidateURI(TestURI{}), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	tests := []struct {
		name string
		id   string
	}{
		{"Negative ID", "-1"},
		{"Zero ID", "0"},
		{"Non-numeric ID", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test/"+tt.id, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		})
	}
}

func TestMustGetValidatedBody(t *testing.T) {
	router := gin.New()
	router.POST("/test", middleware.ValidateJSON(TestUser{}), func(c *gin.Context) {
		var user TestUser
		middleware.MustGetValidatedBody(c, &user)

		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john@example.com", user.Email)
		assert.Equal(t, 25, user.Age)

		c.JSON(http.StatusOK, user)
	})

	body := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDefaultValidationConfig(t *testing.T) {
	config := middleware.DefaultValidationConfig()

	assert.NotNil(t, config.SkipPaths)
	assert.Contains(t, config.SkipPaths, "/health")
	assert.Contains(t, config.SkipPaths, "/metrics")
	assert.NotNil(t, config.CustomMessages)
	assert.NotEmpty(t, config.CustomMessages["required"])
}

// Benchmark tests
func BenchmarkValidateJSON(b *testing.B) {
	router := gin.New()
	router.POST("/test", middleware.ValidateJSON(TestUser{}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	body := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkValidateQuery(b *testing.B) {
	router := gin.New()
	router.GET("/test", middleware.ValidateQuery(TestQuery{}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=1&limit=10&sort=asc", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}
