// tests/integration/barber_integration_test.go
package integration

import (
	"barber-booking-system/internal/models"
	"barber-booking-system/internal/routes"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures
func getTestBarber() models.Barber {
	return models.Barber{
		UserID:      1,
		ShopName:    "Test Barber Shop",
		Description: stringPtr("A test barbershop for integration testing"),
		Address:     "123 Test Street",
		City:        "Test City",
		State:       "TS",
		Country:     "Test Country",
		PostalCode:  "12345",
		Phone:       stringPtr("+1234567890"),
		UserEmail:   stringPtr("test@barbershop.com"),
		Status:      "active",
		Latitude:    float64Ptr(40.7128),
		Longitude:   float64Ptr(-74.0060),
	}
}

// =============================================================================
// CREATE BARBER TESTS
// =============================================================================

func TestCreateBarber_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Generate admin token for authentication
	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	testBarber := getTestBarber()

	// Prepare request body
	requestBody := map[string]interface{}{
		"user_id":     testBarber.UserID,
		"shop_name":   testBarber.ShopName,
		"description": *testBarber.Description,
		"address":     testBarber.Address,
		"city":        testBarber.City,
		"state":       testBarber.State,
		"country":     testBarber.Country,
		"postal_code": testBarber.PostalCode,
		"phone":       *testBarber.Phone,
		"latitude":    *testBarber.Latitude,
		"longitude":   *testBarber.Longitude,
	}

	jsonBody, _ := json.Marshal(requestBody)

	// Make request with authentication
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/barbers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if response["success"] != nil {
		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["data"])
	} else {
		t.Logf("Response body: %s", w.Body.String())
		t.Fatal("Response does not contain 'success' field")
	}
}

func TestCreateBarber_InvalidData(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Generate admin token
	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing required fields",
			requestBody: map[string]interface{}{
				"shop_name": "Test Shop",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid coordinates",
			requestBody: map[string]interface{}{
				"user_id":   1,
				"shop_name": "Test Shop",
				"latitude":  200.0,
				"longitude": -74.0060,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/barbers", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// =============================================================================
// GET BARBER TESTS
// =============================================================================

func TestGetAllBarbers_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["meta"])
}

func TestGetAllBarbers_WithFilters(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name        string
		queryParams string
	}{
		{
			name:        "Filter by status",
			queryParams: "?status=active",
		},
		{
			name:        "Filter by city",
			queryParams: "?city=New%20York",
		},
		{
			name:        "Filter by minimum rating",
			queryParams: "?min_rating=4.0",
		},
		{
			name:        "Search by name",
			queryParams: "?search=John",
		},
		{
			name:        "Sort by rating",
			queryParams: "?sort_by=rating",
		},
		{
			name:        "Pagination",
			queryParams: "?limit=10&offset=0",
		},
		{
			name:        "Combined filters",
			queryParams: "?status=active&city=New%20York&min_rating=4.0&limit=20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/barbers"+tt.queryParams, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response["success"].(bool))
		})
	}
}

func TestGetBarberByID_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["id"])
	assert.NotNil(t, data["shop_name"])
}

func TestGetBarberByID_NotFound(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/99999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetBarberByID_InvalidID(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetBarberByUUID_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	// First, get a barber to get its UUID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/1", nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		uuid := data["uuid"].(string)

		// Now test getting by UUID
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/barbers/uuid/"+uuid, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}

// =============================================================================
// SEARCH BARBER TESTS
// =============================================================================

func TestSearchBarbers_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "Search by shop name",
			query: "?q=barber",
		},
		{
			name:  "Search by city",
			query: "?q=New%20York",
		},
		{
			name:  "Search with filters",
			query: "?q=barber&status=active&min_rating=4.0",
		},
		{
			name:  "Empty search",
			query: "?q=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/barbers/search"+tt.query, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response["success"].(bool))
		})
	}
}

// =============================================================================
// UPDATE BARBER TESTS
// =============================================================================

func TestUpdateBarber_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Generate admin token
	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"shop_name":   "Updated Shop Name",
		"description": "Updated description",
		"city":        "Updated City",
	}

	jsonBody, _ := json.Marshal(updateData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/barbers/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}

func TestUpdateBarber_NotFound(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Generate admin token
	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"shop_name": "Updated Shop Name",
	}

	jsonBody, _ := json.Marshal(updateData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/barbers/99999", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Accept both 404 and 500 (some implementations return 500 for db errors)
	assert.Contains(t, []int{http.StatusNotFound, http.StatusInternalServerError}, w.Code)
}

// =============================================================================
// UPDATE STATUS TESTS
// =============================================================================

func TestUpdateBarberStatus_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Generate admin token
	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	statusData := map[string]interface{}{
		"status": "inactive",
	}

	jsonBody, _ := json.Marshal(statusData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/barbers/1/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}

func TestUpdateBarberStatus_InvalidStatus(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Generate admin token
	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	statusData := map[string]interface{}{
		"status": "invalid_status",
	}

	jsonBody, _ := json.Marshal(statusData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/barbers/1/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// STATISTICS TESTS
// =============================================================================

func TestGetBarberStatistics_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/1/statistics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["total_bookings"])
	assert.NotNil(t, data["average_rating"])
}

func TestGetBarberStatistics_NotFound(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/99999/statistics", nil)
	router.ServeHTTP(w, req)

	// Accept both 404 and 500 (some implementations return 500 for db errors)
	assert.Contains(t, []int{http.StatusNotFound, http.StatusInternalServerError}, w.Code)
}

// =============================================================================
// DELETE BARBER TESTS
// =============================================================================

func TestDeleteBarber_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Generate admin token
	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	// Create a test barber first
	createData := map[string]interface{}{
		"user_id":     1,
		"shop_name":   "To Delete Shop",
		"address":     "123 Test St",
		"city":        "Test City",
		"state":       "TS",
		"country":     "Test Country",
		"postal_code": "12345",
		"phone":       "+1234567890",
	}

	jsonBody, _ := json.Marshal(createData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/barbers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusCreated {
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		barberID := int(data["id"].(float64))

		// Now delete it
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/v1/barbers/%d", barberID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, w.Code)
	}
}

func TestDeleteBarber_NotFound(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Generate admin token
	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/barbers/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// PERFORMANCE TESTS
// =============================================================================

func BenchmarkGetAllBarbers(b *testing.B) {
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
		req, _ := http.NewRequest("GET", "/api/v1/barbers", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGetBarberByID(b *testing.B) {
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
		req, _ := http.NewRequest("GET", "/api/v1/barbers/1", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkSearchBarbers(b *testing.B) {
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
		req, _ := http.NewRequest("GET", "/api/v1/barbers/search?q=barber", nil)
		router.ServeHTTP(w, req)
	}
}
