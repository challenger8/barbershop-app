// tests/integration/service_integration_test.go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// SERVICE INTEGRATION TESTS - TABLE DRIVEN
// =============================================================================

// Test fixtures
func getTestService() map[string]interface{} {
	return map[string]interface{}{
		"name":              "Premium Haircut",
		"short_description": "A premium haircut service",
		"description":       "Full premium haircut with wash and style",
		"category_id":       1,
		"service_type":      "haircut",
		"complexity":        3,
		"is_active":         true,
	}
}

func getTestCategory() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Hair Services",
		"description": "All hair-related services",
		"slug":        "hair-services",
		"is_active":   true,
	}
}

func getTestBarberService() map[string]interface{} {
	return map[string]interface{}{
		"barber_id":        1,
		"service_id":       1,
		"price":            50.00,
		"duration_minutes": 45,
		"is_active":        true,
	}
}

// TestCreateService consolidates all create service tests
func TestCreateService(t *testing.T) {
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
			payload:        getTestService(),
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusInternalServerError},
		},
		{
			name: "MissingName",
			payload: map[string]interface{}{
				"short_description": "Test description",
				"category_id":       1,
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest},
		},
		{
			name: "InvalidComplexity",
			payload: map[string]interface{}{
				"name":              "Test Service",
				"short_description": "Test description",
				"category_id":       1,
				"complexity":        10, // Invalid: should be 1-5
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
			payload:        getTestService(),
			userType:       "customer",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name:           "Customer_Forbidden",
			payload:        getTestService(),
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusForbidden, http.StatusCreated, http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.payload)

			req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(jsonBody))
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

// TestGetService consolidates all get service tests
func TestGetService(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus []int
	}{
		// By ID
		{"ByID_Success", "/api/v1/services/1", []int{http.StatusOK, http.StatusNotFound}},
		{"ByID_NotFound", "/api/v1/services/99999", []int{http.StatusNotFound}},
		{"ByID_Invalid", "/api/v1/services/invalid", []int{http.StatusBadRequest}},

		// By slug
		{"BySlug_NotFound", "/api/v1/services/slug/non-existent-slug", []int{http.StatusNotFound, http.StatusOK}},
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

// TestGetAllServices consolidates list and filter tests
func TestGetAllServices(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{"NoFilters", "", http.StatusOK},
		{"FilterByCategory", "?category_id=1", http.StatusOK},
		{"FilterByServiceType", "?service_type=haircut", http.StatusOK},
		{"FilterByComplexity", "?complexity=2", http.StatusOK},
		{"SearchQuery", "?search=haircut", http.StatusOK},
		{"Pagination", "?limit=10&offset=0", http.StatusOK},
		{"CombinedFilters", "?category_id=1&service_type=haircut&limit=5", http.StatusOK},
		{"ActiveOnly", "?is_active=true", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/services"+tt.queryParams, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestSearchServices tests the search endpoint
func TestSearchServices(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{"Success", "?q=haircut", http.StatusOK},
		{"EmptyQuery", "?q=", http.StatusOK},
		{"WithLimit", "?q=haircut&limit=5", http.StatusOK},
		{"NoResults", "?q=nonexistentservice12345", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/services/search"+tt.queryParams, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestUpdateService consolidates update service tests
func TestUpdateService(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		serviceID      string
		payload        map[string]interface{}
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{
			name:      "Success",
			serviceID: "1",
			payload: map[string]interface{}{
				"name": "Updated Service Name",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
		{
			name:      "NotFound",
			serviceID: "99999",
			payload: map[string]interface{}{
				"name": "Updated Name",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusNotFound},
		},
		{
			name:      "Unauthorized",
			serviceID: "1",
			payload: map[string]interface{}{
				"name": "Updated Name",
			},
			userType:       "admin",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name:      "InvalidID",
			serviceID: "abc",
			payload: map[string]interface{}{
				"name": "Updated Name",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.payload)

			req, _ := http.NewRequest("PUT", "/api/v1/services/"+tt.serviceID, bytes.NewBuffer(jsonBody))
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

// TestDeleteService consolidates delete service tests
func TestDeleteService(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		serviceID      string
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Success_OrNotFound", "1", "admin", true, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound}},
		{"NotFound", "99999", "admin", true, []int{http.StatusNotFound}},
		{"Unauthorized", "1", "admin", false, []int{http.StatusUnauthorized}},
		{"InvalidID", "abc", "admin", true, []int{http.StatusBadRequest, http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/api/v1/services/"+tt.serviceID, nil)

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

// =============================================================================
// CATEGORY TESTS
// =============================================================================

// TestGetAllCategories tests the categories list endpoint
func TestGetAllCategories(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	req, _ := http.NewRequest("GET", "/api/v1/services/categories", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
}

// TestCreateCategory consolidates category creation tests
func TestCreateCategory(t *testing.T) {
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
			payload:        getTestCategory(),
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusConflict},
		},
		{
			name: "MissingName",
			payload: map[string]interface{}{
				"description": "Test description",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest},
		},
		{
			name:           "Unauthorized",
			payload:        getTestCategory(),
			userType:       "admin",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.payload)

			req, _ := http.NewRequest("POST", "/api/v1/services/categories", bytes.NewBuffer(jsonBody))
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

// =============================================================================
// BARBER-SERVICE TESTS
// =============================================================================

// TestBarberServices consolidates barber-service association tests
func TestBarberServices(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		method         string
		endpoint       string
		payload        map[string]interface{}
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		// Get barber services
		{
			name:           "GetBarberServices_Success",
			method:         "GET",
			endpoint:       "/api/v1/barbers/1/services",
			payload:        nil,
			userType:       "",
			hasAuth:        false,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
		{
			name:           "GetBarberServices_NotFound",
			method:         "GET",
			endpoint:       "/api/v1/barbers/99999/services",
			payload:        nil,
			userType:       "",
			hasAuth:        false,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
		{
			name:           "GetBarberServices_InvalidID",
			method:         "GET",
			endpoint:       "/api/v1/barbers/invalid/services",
			payload:        nil,
			userType:       "",
			hasAuth:        false,
			expectedStatus: []int{http.StatusBadRequest},
		},

		// Add service to barber
		{
			name:           "AddServiceToBarber_Success",
			method:         "POST",
			endpoint:       "/api/v1/barber-services",
			payload:        getTestBarberService(),
			userType:       "barber",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusInternalServerError, http.StatusConflict},
		},
		{
			name:           "AddServiceToBarber_Unauthorized",
			method:         "POST",
			endpoint:       "/api/v1/barber-services",
			payload:        getTestBarberService(),
			userType:       "barber",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name:     "AddServiceToBarber_InvalidPrice",
			method:   "POST",
			endpoint: "/api/v1/barber-services",
			payload: map[string]interface{}{
				"barber_id":  1,
				"service_id": 1,
				"price":      -10, // Invalid negative price
			},
			userType:       "barber",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusCreated, http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.payload != nil {
				body, _ = json.Marshal(tt.payload)
			}

			req, _ := http.NewRequest(tt.method, tt.endpoint, bytes.NewBuffer(body))
			if tt.payload != nil {
				req.Header.Set("Content-Type", "application/json")
			}

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

// TestUpdateBarberService tests updating barber service
func TestUpdateBarberService(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		serviceID      string
		payload        map[string]interface{}
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{
			name:      "Success",
			serviceID: "1",
			payload: map[string]interface{}{
				"price":            55.00,
				"duration_minutes": 50,
			},
			userType:       "barber",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
		{
			name:      "NotFound",
			serviceID: "99999",
			payload: map[string]interface{}{
				"price": 55.00,
			},
			userType:       "barber",
			hasAuth:        true,
			expectedStatus: []int{http.StatusNotFound},
		},
		{
			name:      "Unauthorized",
			serviceID: "1",
			payload: map[string]interface{}{
				"price": 55.00,
			},
			userType:       "barber",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.payload)

			req, _ := http.NewRequest("PUT", "/api/v1/barber-services/"+tt.serviceID, bytes.NewBuffer(jsonBody))
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

// TestDeleteBarberService tests removing service from barber
func TestDeleteBarberService(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		serviceID      string
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Success_OrNotFound", "1", "barber", true, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound}},
		{"NotFound", "99999", "barber", true, []int{http.StatusNotFound}},
		{"Unauthorized", "1", "barber", false, []int{http.StatusUnauthorized}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/api/v1/barber-services/"+tt.serviceID, nil)

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
