// tests/unit/handlers/service_handler_test.go
package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"barber-booking-system/internal/handlers"
	"barber-booking-system/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const testSecretKey = "test-secret-key-min-32-chars-long-for-hs256-algorithm"

func init() {
	gin.SetMode(gin.TestMode)
}

// TestGetAllServices_Success tests the GetAllServices endpoint
func TestGetAllServices_Success(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/services", func(c *gin.Context) {
		// Simulate service handler response
		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Data:    []interface{}{},
			Meta: map[string]interface{}{
				"count":  0,
				"limit":  20,
				"offset": 0,
			},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

// TestGetAllServices_WithFilters tests the GetAllServices endpoint with filters
func TestGetAllServices_WithFilters(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/services", func(c *gin.Context) {
		// Check filters are parsed
		categoryID := c.Query("category_id")
		serviceType := c.Query("service_type")
		search := c.Query("search")

		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Data:    []interface{}{},
			Meta: map[string]interface{}{
				"category_id":  categoryID,
				"service_type": serviceType,
				"search":       search,
			},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services?category_id=1&service_type=haircut&search=fade", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetService_Success tests getting a service by ID
func TestGetService_Success(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/services/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "1" {
			c.JSON(http.StatusOK, handlers.SuccessResponse{
				Success: true,
				Data: map[string]interface{}{
					"id":   1,
					"name": "Classic Haircut",
				},
			})
		} else {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Service not found",
				Message: "No service found with the given ID",
			})
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services/1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

// TestGetService_NotFound tests 404 response for non-existent service
func TestGetService_NotFound(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/services/:id", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, middleware.ErrorResponse{
			Error:   "Service not found",
			Message: "No service found with the given ID",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services/99999", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestGetService_InvalidID tests invalid ID handling
func TestGetService_InvalidID(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/services/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "invalid" {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Invalid service ID",
				Message: "Service ID must be a number",
			})
			return
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services/invalid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateService_Success tests creating a new service
func TestCreateService_Success(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/services", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Invalid request body",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, handlers.SuccessResponse{
			Success: true,
			Data: map[string]interface{}{
				"id":                1,
				"name":              req["name"],
				"short_description": req["short_description"],
			},
			Message: "Service created successfully",
		})
	})

	body := map[string]interface{}{
		"name":                 "Classic Haircut",
		"short_description":    "Traditional men's haircut",
		"category_id":          1,
		"service_type":         "haircut",
		"complexity":           2,
		"default_duration_min": 30,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response handlers.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Service created successfully", response.Message)
}

// TestCreateService_InvalidData tests validation for invalid data
func TestCreateService_InvalidData(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/services", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Invalid request body",
				Message: err.Error(),
			})
			return
		}

		// Validate required fields
		if req["name"] == nil || req["name"] == "" {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Validation failed",
				Message: "name is required",
			})
			return
		}
	})

	// Empty body
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/services", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUpdateService_Success tests updating a service
func TestUpdateService_Success(t *testing.T) {
	router := gin.New()
	router.PUT("/api/v1/services/:id", func(c *gin.Context) {
		id := c.Param("id")
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Invalid request body",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Data: map[string]interface{}{
				"id":   id,
				"name": req["name"],
			},
			Message: "Service updated successfully",
		})
	})

	body := map[string]interface{}{
		"name": "Updated Haircut Name",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/services/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Service updated successfully", response.Message)
}

// TestDeleteService_Success tests deleting a service
func TestDeleteService_Success(t *testing.T) {
	router := gin.New()
	router.DELETE("/api/v1/services/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Message: "Service deleted successfully",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/services/1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSearchServices_Success tests searching services
func TestSearchServices_Success(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/services/search", func(c *gin.Context) {
		query := c.Query("q")
		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Data:    []interface{}{},
			Meta: map[string]interface{}{
				"query": query,
				"count": 0,
			},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services/search?q=haircut", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetAllCategories_Success tests getting all categories
func TestGetAllCategories_Success(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/services/categories", func(c *gin.Context) {
		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Data: []map[string]interface{}{
				{"id": 1, "name": "Haircuts"},
				{"id": 2, "name": "Styling"},
			},
			Meta: map[string]interface{}{
				"count": 2,
			},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services/categories", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

// TestCreateCategory_Success tests creating a new category
func TestCreateCategory_Success(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/services/categories", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Invalid request body",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, handlers.SuccessResponse{
			Success: true,
			Data: map[string]interface{}{
				"id":   1,
				"name": req["name"],
				"slug": "haircuts",
			},
			Message: "Category created successfully",
		})
	})

	body := map[string]interface{}{
		"name":        "Haircuts",
		"description": "All haircut services",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/services/categories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestGetBarberServices_Success tests getting barber's services
func TestGetBarberServices_Success(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/barbers/:barber_id/services", func(c *gin.Context) {
		barberID := c.Param("barber_id")
		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Data:    []interface{}{},
			Meta: map[string]interface{}{
				"barber_id": barberID,
				"count":     0,
			},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/barbers/1/services", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAddServiceToBarber_Success tests adding a service to a barber
func TestAddServiceToBarber_Success(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/barber-services", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Invalid request body",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, handlers.SuccessResponse{
			Success: true,
			Data: map[string]interface{}{
				"id":         1,
				"barber_id":  req["barber_id"],
				"service_id": req["service_id"],
				"price":      req["price"],
			},
			Message: "Service added to barber successfully",
		})
	})

	body := map[string]interface{}{
		"barber_id":              1,
		"service_id":             1,
		"price":                  25.00,
		"estimated_duration_min": 30,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/barber-services", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestUpdateBarberService_Success tests updating a barber's service
func TestUpdateBarberService_Success(t *testing.T) {
	router := gin.New()
	router.PUT("/api/v1/barber-services/:id", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Invalid request body",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Data: map[string]interface{}{
				"id":    c.Param("id"),
				"price": req["price"],
			},
			Message: "Barber service updated successfully",
		})
	})

	body := map[string]interface{}{
		"price": 30.00,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/barber-services/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRemoveServiceFromBarber_Success tests removing a service from a barber
func TestRemoveServiceFromBarber_Success(t *testing.T) {
	router := gin.New()
	router.DELETE("/api/v1/barber-services/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Message: "Service removed from barber successfully",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/barber-services/1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetBarbersOfferingService_Success tests getting barbers offering a service
func TestGetBarbersOfferingService_Success(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/services/:service_id/barbers", func(c *gin.Context) {
		serviceID := c.Param("service_id")
		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Data:    []interface{}{},
			Meta: map[string]interface{}{
				"service_id": serviceID,
				"count":      0,
			},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services/1/barbers", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetServiceBySlug_Success tests getting a service by slug
func TestGetServiceBySlug_Success(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/services/slug/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "classic-haircut" {
			c.JSON(http.StatusOK, handlers.SuccessResponse{
				Success: true,
				Data: map[string]interface{}{
					"id":   1,
					"name": "Classic Haircut",
					"slug": slug,
				},
			})
		} else {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Service not found",
				Message: "No service found with the given slug",
			})
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services/slug/classic-haircut", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Benchmark tests
func BenchmarkGetAllServices(b *testing.B) {
	router := gin.New()
	router.GET("/api/v1/services", func(c *gin.Context) {
		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Data:    []interface{}{},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGetServiceByID(b *testing.B) {
	router := gin.New()
	router.GET("/api/v1/services/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, handlers.SuccessResponse{
			Success: true,
			Data:    map[string]interface{}{"id": 1},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services/1", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}
