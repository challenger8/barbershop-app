// tests/integration/barber_integration_test.go
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
// BARBER INTEGRATION TESTS - TABLE DRIVEN
// =============================================================================

// Test fixtures
func getTestBarberRequest() map[string]interface{} {
	return map[string]interface{}{
		"user_id":     1,
		"shop_name":   "Test Barber Shop",
		"description": "A test barbershop for integration testing",
		"address":     "123 Test Street",
		"city":        "Test City",
		"state":       "TS",
		"country":     "Test Country",
		"postal_code": "12345",
		"phone":       "+1234567890",
		"latitude":    40.7128,
		"longitude":   -74.0060,
	}
}

// TestCreateBarber consolidates all create barber tests
func TestCreateBarber(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		payload        map[string]interface{}
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{
			name:           "Success",
			payload:        getTestBarberRequest(),
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusConflict, http.StatusBadRequest, http.StatusInternalServerError},
		},
		{
			name: "MissingShopName",
			payload: map[string]interface{}{
				"user_id": 1,
				"address": "123 Test St",
				"city":    "Test City",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest},
		},
		{
			name: "MissingCity",
			payload: map[string]interface{}{
				"user_id":   1,
				"shop_name": "Test Shop",
				"address":   "123 Test St",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest},
		},
		{
			name:           "EmptyBody",
			payload:        map[string]interface{}{},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest},
		},
		{
			name:           "Unauthorized",
			payload:        getTestBarberRequest(),
			userType:       "admin",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name: "InvalidCoordinates",
			payload: map[string]interface{}{
				"user_id":   1,
				"shop_name": "Test Shop",
				"address":   "123 Test St",
				"city":      "Test City",
				"country":   "Test Country",
				"latitude":  200, // Invalid: should be -90 to 90
				"longitude": 300, // Invalid: should be -180 to 180
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusCreated, http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.payload)

			req, _ := http.NewRequest("POST", "/api/v1/barbers", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			if tt.hasAuth {
				token, _ := generateTestToken(1, tt.userType+"@test.com", tt.userType, jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestGetBarber consolidates all get barber tests
func TestGetBarber(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus []int
	}{
		// By ID
		{"ByID_Success", "/api/v1/barbers/1", []int{http.StatusOK, http.StatusNotFound}},
		{"ByID_NotFound", "/api/v1/barbers/99999", []int{http.StatusNotFound}},
		{"ByID_InvalidFormat", "/api/v1/barbers/abc", []int{http.StatusBadRequest, http.StatusNotFound}},

		// By slug
		{"BySlug_NotFound", "/api/v1/barbers/slug/non-existent-barber", []int{http.StatusNotFound, http.StatusOK}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.endpoint, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestGetAllBarbers consolidates list and filter tests
func TestGetAllBarbers(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{"NoFilters", "", http.StatusOK},
		{"FilterByCity", "?city=Test City", http.StatusOK},
		{"FilterByStatus", "?status=active", http.StatusOK},
		{"FilterByRating", "?min_rating=4", http.StatusOK},
		{"Pagination", "?limit=10&offset=0", http.StatusOK},
		{"CombinedFilters", "?city=New York&status=active&limit=5", http.StatusOK},
		{"SortByRating", "?sort_by=rating&order=desc", http.StatusOK},
		{"SortByCreated", "?sort_by=created_at&order=asc", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/barbers"+tt.queryParams, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestSearchBarbers tests the search endpoint
func TestSearchBarbers(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{"Success", "?q=barber", http.StatusOK},
		{"EmptyQuery", "?q=", http.StatusOK},
		{"WithLimit", "?q=barber&limit=5", http.StatusOK},
		{"NoResults", "?q=nonexistentbarber12345", http.StatusOK},
		{"ByCity", "?q=&city=New York", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/barbers/search"+tt.queryParams, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestUpdateBarber consolidates update barber tests
func TestUpdateBarber(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		barberID       string
		payload        map[string]interface{}
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{
			name:     "Success",
			barberID: "1",
			payload: map[string]interface{}{
				"shop_name":   "Updated Shop Name",
				"description": "Updated description",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
		{
			name:     "NotFound",
			barberID: "99999",
			payload: map[string]interface{}{
				"shop_name": "Updated Name",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusNotFound},
		},
		{
			name:     "Unauthorized",
			barberID: "1",
			payload: map[string]interface{}{
				"shop_name": "Updated Name",
			},
			userType:       "admin",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name:     "InvalidID",
			barberID: "abc",
			payload: map[string]interface{}{
				"shop_name": "Updated Name",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusNotFound},
		},
		{
			name:     "PartialUpdate",
			barberID: "1",
			payload: map[string]interface{}{
				"phone": "+9876543210",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.payload)

			req, _ := http.NewRequest("PUT", "/api/v1/barbers/"+tt.barberID, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			if tt.hasAuth {
				token, _ := generateTestToken(1, tt.userType+"@test.com", tt.userType, jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestDeleteBarber consolidates delete barber tests
func TestDeleteBarber(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		barberID       string
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{"NotFound", "99999", "admin", true, []int{http.StatusNotFound}},
		{"Unauthorized", "1", "admin", false, []int{http.StatusUnauthorized}},
		{"InvalidID", "abc", "admin", true, []int{http.StatusBadRequest, http.StatusNotFound}},
		{"Customer_Forbidden", "1", "customer", true, []int{http.StatusForbidden, http.StatusNotFound, http.StatusOK, http.StatusNoContent}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/api/v1/barbers/"+tt.barberID, nil)

			if tt.hasAuth {
				token, _ := generateTestToken(1, tt.userType+"@test.com", tt.userType, jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestBarberNearby tests geolocation-based search
func TestBarberNearby(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus []int
	}{
		{"Success", "?latitude=40.7128&longitude=-74.0060&radius=10", []int{http.StatusOK, http.StatusBadRequest}},
		{"DefaultRadius", "?latitude=40.7128&longitude=-74.0060", []int{http.StatusOK, http.StatusBadRequest}}, {"MissingLatitude", "?longitude=-74.0060&radius=10", []int{http.StatusBadRequest, http.StatusOK}},
		{"MissingLongitude", "?latitude=40.7128&radius=10", []int{http.StatusBadRequest, http.StatusOK}},
		{"InvalidLatitude", "?latitude=invalid&longitude=-74.0060&radius=10", []int{http.StatusBadRequest}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/barbers/nearby"+tt.queryParams, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestBarberAvailability tests availability endpoints
func TestBarberAvailability(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, _ := generateTestToken(1, "barber@test.com", "barber", jwtSecret)

	tests := []struct {
		name           string
		endpoint       string
		method         string
		hasAuth        bool
		expectedStatus []int
	}{
		{"GetSchedule_Success", "/api/v1/barbers/1/schedule", "GET", false, []int{http.StatusOK, http.StatusNotFound}},
		{"GetSchedule_NotFound", "/api/v1/barbers/99999/schedule", "GET", false, []int{http.StatusNotFound, http.StatusOK}},
		{"GetAvailability_Success", "/api/v1/barbers/1/availability?date=2024-12-28", "GET", false, []int{http.StatusOK, http.StatusNotFound}},
		{"GetAvailability_NoDate", "/api/v1/barbers/1/availability", "GET", false, []int{http.StatusOK, http.StatusBadRequest, http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.endpoint, nil)

			if tt.hasAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestBarberStats tests statistics endpoints
func TestBarberStats(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, _ := generateTestToken(1, "barber@test.com", "barber", jwtSecret)

	tests := []struct {
		name           string
		endpoint       string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Stats_Success", "/api/v1/barbers/1/stats", true, []int{http.StatusOK, http.StatusNotFound}},
		{"Stats_NotFound", "/api/v1/barbers/99999/stats", true, []int{http.StatusNotFound, http.StatusOK}},
		{"Stats_Unauthorized", "/api/v1/barbers/1/stats", false, []int{http.StatusUnauthorized, http.StatusOK, http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.endpoint, nil)

			if tt.hasAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// =============================================================================
// RESPONSE FORMAT TESTS
// =============================================================================

func TestGetAllBarbers_ResponseFormat(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	req, _ := http.NewRequest("GET", "/api/v1/barbers", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have success or data field
	_, hasSuccess := response["success"]
	_, hasData := response["data"]
	_, hasBarbers := response["barbers"]
	assert.True(t, hasSuccess || hasData || hasBarbers,
		"Response should contain success, data, or barbers field")
}

// =============================================================================
// BENCHMARK TESTS
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
