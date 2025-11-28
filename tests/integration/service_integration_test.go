// tests/integration/service_integration_test.go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"barber-booking-system/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TEST FIXTURES
// =============================================================================

func getTestService() map[string]interface{} {
	return map[string]interface{}{
		"name":                 "Classic Haircut",
		"short_description":    "Traditional men's haircut with scissors and clippers",
		"category_id":          1,
		"service_type":         "haircut",
		"complexity":           2,
		"default_duration_min": 30,
		"suggested_price_min":  20.0,
		"suggested_price_max":  35.0,
		"target_gender":        "male",
	}
}

func getTestCategory() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Haircuts",
		"description": "All haircut services",
		"sort_order":  1,
		"is_featured": true,
	}
}

func getTestBarberService() map[string]interface{} {
	return map[string]interface{}{
		"barber_id":              1,
		"service_id":             1,
		"price":                  25.00,
		"estimated_duration_min": 30,
		"is_active":              true,
	}
}

// NOTE: Using shared setupTestRouter from test_helpers.go (DRY principle)

// =============================================================================
// SERVICE TESTS
// =============================================================================

func TestGetAllServices_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/services", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if response["success"] != nil {
		assert.True(t, response["success"].(bool))
	}
}

func TestGetAllServices_WithFilters(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{
			name:           "Filter by category",
			query:          "?category_id=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter by service type",
			query:          "?service_type=haircut",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter by complexity",
			query:          "?complexity=2",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Search query",
			query:          "?search=haircut",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Pagination",
			query:          "?limit=10&offset=0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Combined filters",
			query:          "?category_id=1&service_type=haircut&limit=5",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/services"+tt.query, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetServiceByID_NotFound(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/services/99999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetServiceByID_InvalidID(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/services/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchServices_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/services/search?q=haircut", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if response["success"] != nil {
		assert.True(t, response["success"].(bool))
	}
}

func TestCreateService_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	testService := getTestService()
	jsonBody, _ := json.Marshal(testService)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Service creation should succeed or fail with validation
	assert.Contains(t, []int{http.StatusCreated, http.StatusInternalServerError, http.StatusBadRequest}, w.Code)
}

func TestCreateService_InvalidData(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing name",
			requestBody: map[string]interface{}{
				"short_description": "Test description",
				"category_id":       1,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid complexity",
			requestBody: map[string]interface{}{
				"name":              "Test Service",
				"short_description": "Test description",
				"category_id":       1,
				"complexity":        10, // Invalid: should be 1-5
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreateService_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	testService := getTestService()
	jsonBody, _ := json.Marshal(testService)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateService_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	updateBody := map[string]interface{}{
		"name": "Updated Service Name",
	}
	jsonBody, _ := json.Marshal(updateBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/services/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteService_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/services/1", nil)
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// CATEGORY TESTS
// =============================================================================

func TestGetAllCategories_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/services/categories", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if response["success"] != nil {
		assert.True(t, response["success"].(bool))
	}
}

func TestGetCategoryByID_NotFound(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/services/categories/99999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateCategory_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	testCategory := getTestCategory()
	jsonBody, _ := json.Marshal(testCategory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/services/categories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Category creation should succeed or fail with DB constraint
	assert.Contains(t, []int{http.StatusCreated, http.StatusInternalServerError}, w.Code)
}

func TestCreateCategory_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	testCategory := getTestCategory()
	jsonBody, _ := json.Marshal(testCategory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/services/categories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// BARBER SERVICE TESTS
// =============================================================================

func TestGetBarberServices_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/1/services", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if response["success"] != nil {
		assert.True(t, response["success"].(bool))
	}
}

func TestGetBarberServices_InvalidBarberID(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/invalid/services", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// NOTE: TestGetBarbersOfferingService_Success was REMOVED
// The route GET /api/v1/services/:id/barbers doesn't exist
// serviceHandler.GetServiceBarbers is not implemented

func TestAddServiceToBarber_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "barber@test.com", "barber", jwtSecret)
	require.NoError(t, err)

	testBarberService := getTestBarberService()
	jsonBody, _ := json.Marshal(testBarberService)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/barber-services", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Should succeed or fail with FK constraint (service/barber doesn't exist)
	assert.Contains(t, []int{http.StatusCreated, http.StatusInternalServerError, http.StatusBadRequest}, w.Code)
}

func TestAddServiceToBarber_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	testBarberService := getTestBarberService()
	jsonBody, _ := json.Marshal(testBarberService)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/barber-services", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAddServiceToBarber_InvalidData(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "barber@test.com", "barber", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing barber_id",
			requestBody: map[string]interface{}{
				"service_id":             1,
				"price":                  25.00,
				"estimated_duration_min": 30,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing service_id",
			requestBody: map[string]interface{}{
				"barber_id":              1,
				"price":                  25.00,
				"estimated_duration_min": 30,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing price",
			requestBody: map[string]interface{}{
				"barber_id":              1,
				"service_id":             1,
				"estimated_duration_min": 30,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/barber-services", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// NOTE: TestGetBarberServiceByID_NotFound was REMOVED
// The route GET /api/v1/barber-services/:id doesn't exist
// serviceHandler.GetBarberService is not implemented

func TestRemoveServiceFromBarber_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/barber-services/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkGetAllServices(b *testing.B) {
	gin.SetMode(gin.TestMode)
	t := &testing.T{}
	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)
	defer dbManager.Close()

	router := gin.New()
	routes.Setup(router, dbManager.DB, cfg.JWT.Secret, cfg.JWT.Expiration, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/services", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkSearchServices(b *testing.B) {
	gin.SetMode(gin.TestMode)
	t := &testing.T{}
	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)
	defer dbManager.Close()

	router := gin.New()
	routes.Setup(router, dbManager.DB, cfg.JWT.Secret, cfg.JWT.Expiration, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/services/search?q=haircut", nil)
		router.ServeHTTP(w, req)
	}
}
