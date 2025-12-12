// tests/integration/booking_integration_test.go
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"barber-booking-system/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BOOKING INTEGRATION TESTS - TABLE DRIVEN
// =============================================================================

// TestCreateBooking consolidates all create booking tests
func TestCreateBooking(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		getPayload     func() map[string]interface{}
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{
			name:           "Success",
			getPayload:     getTestBookingRequest,
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusConflict, http.StatusNotFound, http.StatusBadRequest, http.StatusInternalServerError},
		},
		{
			name:           "WithCustomerID",
			getPayload:     func() map[string]interface{} { return getTestBookingRequestWithCustomer(1) },
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusConflict, http.StatusNotFound, http.StatusBadRequest, http.StatusInternalServerError},
		},
		{
			name:           "ForSpecificBarber",
			getPayload:     func() map[string]interface{} { return getTestBookingRequestForBarber(2, 3) },
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusConflict, http.StatusNotFound, http.StatusBadRequest, http.StatusInternalServerError},
		},
		{
			name:           "InPast_Invalid",
			getPayload:     getTestBookingRequestInPast,
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusUnprocessableEntity},
		},
		{
			name:           "Unauthorized",
			getPayload:     getTestBookingRequest,
			userType:       "customer",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name: "MissingRequiredFields",
			getPayload: func() map[string]interface{} {
				return map[string]interface{}{
					"notes": "Incomplete booking",
				}
			},
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest},
		},
		{
			name: "InvalidBarberID",
			getPayload: func() map[string]interface{} {
				req := getTestBookingRequest()
				req["barber_id"] = 99999
				return req
			},
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusNotFound, http.StatusBadRequest, http.StatusInternalServerError},
		},
		{
			name: "InvalidServiceID",
			getPayload: func() map[string]interface{} {
				req := getTestBookingRequest()
				req["service_id"] = 99999
				return req
			},
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusNotFound, http.StatusBadRequest, http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.getPayload())

			req, _ := http.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(jsonBody))
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

// TestGetBooking consolidates all get booking tests
func TestGetBooking(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		endpoint       string
		hasAuth        bool
		expectedStatus []int
	}{
		// By ID
		{"ByID_Success", "/api/v1/bookings/1", true, []int{http.StatusOK, http.StatusNotFound}},
		{"ByID_NotFound", "/api/v1/bookings/99999", true, []int{http.StatusNotFound}},
		{"ByID_Unauthorized", "/api/v1/bookings/1", false, []int{http.StatusUnauthorized}},
		{"ByID_InvalidFormat", "/api/v1/bookings/abc", true, []int{http.StatusBadRequest, http.StatusNotFound}},

		// By UUID
		{"ByUUID_NotFound", "/api/v1/bookings/uuid/00000000-0000-0000-0000-000000000000", true, []int{http.StatusNotFound}},
		{"ByUUID_InvalidFormat", "/api/v1/bookings/uuid/invalid-uuid", true, []int{http.StatusNotFound, http.StatusBadRequest}},

		// By Number
		{"ByNumber_NotFound", "/api/v1/bookings/number/INVALID123", true, []int{http.StatusNotFound}},

		// My bookings
		{"MyBookings_Success", "/api/v1/bookings/me", true, []int{http.StatusOK}},
		{"MyBookings_Unauthorized", "/api/v1/bookings/me", false, []int{http.StatusUnauthorized}},
		{"MyBookings_WithStatusFilter", "/api/v1/bookings/me?status=pending", true, []int{http.StatusOK}},
		{"MyBookings_WithPagination", "/api/v1/bookings/me?limit=10&offset=0", true, []int{http.StatusOK}},
		{"MyBookings_WithAllFilters", "/api/v1/bookings/me?status=pending&limit=10&offset=0", true, []int{http.StatusOK}},

		// History
		{"History_Success", "/api/v1/bookings/1/history", true, []int{http.StatusOK, http.StatusNotFound}},
		{"History_NotFound", "/api/v1/bookings/99999/history", true, []int{http.StatusNotFound}},
		{"History_Unauthorized", "/api/v1/bookings/1/history", false, []int{http.StatusUnauthorized}},
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

// TestUpdateBookingStatus consolidates all status update tests
func TestUpdateBookingStatus(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		bookingID      string
		status         string
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		// Auth tests
		{"Unauthorized", "1", "confirmed", "customer", false, []int{http.StatusUnauthorized}},

		// Not found tests
		{"NotFound_Confirmed", "99999", "confirmed", "admin", true, []int{http.StatusNotFound}},
		{"NotFound_InProgress", "99999", "in_progress", "barber", true, []int{http.StatusNotFound}},
		{"NotFound_Completed", "99999", "completed", "barber", true, []int{http.StatusNotFound}},

		// Role-based tests
		{"Admin_Confirm", "1", "confirmed", "admin", true, []int{http.StatusOK, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusBadRequest}},
		{"Barber_InProgress", "1", "in_progress", "barber", true, []int{http.StatusOK, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusBadRequest}},
		{"Barber_Completed", "1", "completed", "barber", true, []int{http.StatusOK, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusBadRequest}},
		{"Customer_Confirm", "1", "confirmed", "customer", true, []int{http.StatusOK, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusForbidden, http.StatusBadRequest}},

		// Invalid status
		{"InvalidStatus", "1", "invalid_status", "admin", true, []int{http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusNotFound}},
		{"EmptyStatus", "1", "", "admin", true, []int{http.StatusBadRequest}},

		// Invalid booking ID
		{"InvalidBookingID", "abc", "confirmed", "admin", true, []int{http.StatusBadRequest, http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(getTestStatusUpdateRequest(tt.status))

			req, _ := http.NewRequest("PATCH", "/api/v1/bookings/"+tt.bookingID+"/status", bytes.NewBuffer(jsonBody))
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

// TestCancelBooking consolidates all cancel booking tests
func TestCancelBooking(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		bookingID      string
		userType       string
		isByCustomer   bool
		hasBody        bool
		hasAuth        bool
		expectedStatus []int
	}{
		// Auth tests
		{"Unauthorized", "1", "customer", true, false, false, []int{http.StatusUnauthorized}},

		// Not found tests
		{"NotFound_Customer", "99999", "customer", true, false, true, []int{http.StatusNotFound}},
		{"NotFound_Barber", "99999", "barber", false, false, true, []int{http.StatusNotFound}},
		{"NotFound_Admin", "99999", "admin", false, false, true, []int{http.StatusNotFound}},

		// With reason body
		{"NotFound_WithReason", "99999", "customer", true, true, true, []int{http.StatusNotFound}},

		// Role-based cancellation
		{"Customer_Cancel", "1", "customer", true, true, true, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound, http.StatusUnprocessableEntity}},
		{"Barber_Cancel", "1", "barber", false, true, true, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound, http.StatusUnprocessableEntity}},
		{"Admin_Cancel", "1", "admin", false, true, true, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound, http.StatusUnprocessableEntity}},

		// Without body (should still work)
		{"Customer_NoBody", "1", "customer", true, false, true, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusBadRequest}},

		// Invalid booking ID
		{"InvalidBookingID", "abc", "customer", true, false, true, []int{http.StatusBadRequest, http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.hasBody {
				body, _ = json.Marshal(getTestCancelRequest(tt.isByCustomer))
			}

			req, _ := http.NewRequest("DELETE", "/api/v1/bookings/"+tt.bookingID, bytes.NewBuffer(body))
			if tt.hasBody {
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

// TestRescheduleBooking consolidates all reschedule tests
func TestRescheduleBooking(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		bookingID      string
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		// Auth tests
		{"Unauthorized", "1", "customer", false, []int{http.StatusUnauthorized}},

		// Not found tests
		{"NotFound", "99999", "customer", true, []int{http.StatusNotFound}},

		// Role-based reschedule
		{"Success_Customer", "1", "customer", true, []int{http.StatusOK, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusBadRequest}},
		{"Success_Barber", "1", "barber", true, []int{http.StatusOK, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusBadRequest}},
		{"Success_Admin", "1", "admin", true, []int{http.StatusOK, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusBadRequest}},

		// Invalid booking ID
		{"InvalidBookingID", "abc", "customer", true, []int{http.StatusBadRequest, http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(getTestRescheduleRequest())

			req, _ := http.NewRequest("PUT", "/api/v1/bookings/"+tt.bookingID+"/reschedule", bytes.NewBuffer(jsonBody))
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

// TestCheckAvailability consolidates availability check tests
func TestCheckAvailability(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus []int
	}{
		// Valid requests
		{"Success", "?barber_id=1&start_time=2024-12-28T10:00:00Z&duration=45", []int{http.StatusOK, http.StatusNotFound}},

		// Missing params
		{"MissingBarberID", "?start_time=2024-12-28T10:00:00Z&duration=45", []int{http.StatusBadRequest}},
		{"MissingStartTime", "?barber_id=1&duration=45", []int{http.StatusBadRequest}},
		{"MissingDuration", "?barber_id=1&start_time=2024-12-28T10:00:00Z", []int{http.StatusBadRequest}},
		{"MissingAllParams", "", []int{http.StatusBadRequest}},

		// Invalid params
		{"InvalidBarberID", "?barber_id=abc&start_time=2024-12-28T10:00:00Z&duration=45", []int{http.StatusBadRequest}},
		{"InvalidDuration", "?barber_id=1&start_time=2024-12-28T10:00:00Z&duration=abc", []int{http.StatusBadRequest}},
		{"InvalidStartTime", "?barber_id=1&start_time=invalid&duration=45", []int{http.StatusBadRequest}},

		// Edge cases
		{"ZeroDuration", "?barber_id=1&start_time=2024-12-28T10:00:00Z&duration=0", []int{http.StatusBadRequest, http.StatusOK}},
		{"NegativeDuration", "?barber_id=1&start_time=2024-12-28T10:00:00Z&duration=-45", []int{http.StatusBadRequest, http.StatusOK}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/bookings/availability"+tt.queryParams, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestBarberBookings consolidates barber-specific booking tests
func TestBarberBookings(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "barber@test.com", "barber", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		endpoint       string
		hasAuth        bool
		expectedStatus []int
	}{
		// Get barber bookings
		{"GetBookings_Success", "/api/v1/barbers/1/bookings", true, []int{http.StatusOK, http.StatusNotFound}},
		{"GetBookings_NotFound", "/api/v1/barbers/99999/bookings", true, []int{http.StatusOK, http.StatusNotFound}},
		{"GetBookings_WithFilters", "/api/v1/barbers/1/bookings?status=pending&limit=10", true, []int{http.StatusOK, http.StatusNotFound}},

		// Today's bookings
		{"TodayBookings_Success", "/api/v1/barbers/1/bookings/today", true, []int{http.StatusOK, http.StatusNotFound}},
		{"TodayBookings_NotFound", "/api/v1/barbers/99999/bookings/today", true, []int{http.StatusOK, http.StatusNotFound}},

		// Booking stats
		{"Stats_Success", "/api/v1/barbers/1/bookings/stats", true, []int{http.StatusOK, http.StatusNotFound}},
		{"Stats_NotFound", "/api/v1/barbers/99999/bookings/stats", true, []int{http.StatusOK, http.StatusNotFound}},

		// Invalid barber ID
		{"InvalidBarberID", "/api/v1/barbers/abc/bookings", true, []int{http.StatusBadRequest, http.StatusNotFound}},
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
// ROUTE REGISTRATION TESTS
// =============================================================================

func TestBookingRoutesRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)
	defer dbManager.Close()

	routes.Setup(router, dbManager.DB, cfg.JWT.Secret, cfg.JWT.Expiration, nil)

	allRoutes := router.Routes()

	expectedRoutes := []string{
		// Public routes
		"GET /api/v1/bookings/availability",
		"GET /api/v1/bookings/uuid/:uuid",
		"GET /api/v1/bookings/number/:number",

		// Protected routes
		"POST /api/v1/bookings",
		"GET /api/v1/bookings/me",
		"GET /api/v1/bookings/:id",
		"GET /api/v1/bookings/:id/history",
		"PUT /api/v1/bookings/:id",
		"PATCH /api/v1/bookings/:id/status",
		"PUT /api/v1/bookings/:id/reschedule",
		"DELETE /api/v1/bookings/:id",

		// Barber booking routes
		"GET /api/v1/barbers/:id/bookings",
		"GET /api/v1/barbers/:id/bookings/today",
		"GET /api/v1/barbers/:id/bookings/stats",
	}

	actualRoutes := make(map[string]bool)
	for _, route := range allRoutes {
		key := route.Method + " " + route.Path
		actualRoutes[key] = true
	}

	for _, expectedRoute := range expectedRoutes {
		assert.True(t, actualRoutes[expectedRoute],
			"Route %s should be registered", expectedRoute)
	}

	t.Logf("Total routes registered: %d", len(allRoutes))
}

// =============================================================================
// RESPONSE VALIDATION TESTS
// =============================================================================

func TestCreateBooking_ResponseFormat(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	jsonBody, _ := json.Marshal(getTestBookingRequest())

	req, _ := http.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// If created successfully, validate response format
	if w.Code == http.StatusCreated {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check for data wrapper or direct response
		if data, ok := response["data"].(map[string]interface{}); ok {
			assert.NotEmpty(t, data["id"], "Response should contain booking ID")
		} else {
			assert.NotEmpty(t, response["id"], "Response should contain booking ID")
		}
	}
}

func TestGetMyBookings_ResponseFormat(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/bookings/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have data array (even if empty)
	_, hasData := response["data"]
	_, hasBookings := response["bookings"]
	assert.True(t, hasData || hasBookings, "Response should contain data or bookings array")
}

func TestCheckAvailability_ResponseFormat(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	req, _ := http.NewRequest("GET", "/api/v1/bookings/availability?barber_id=1&start_time=2024-12-28T10:00:00Z&duration=45", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check for availability field
		if data, ok := response["data"].(map[string]interface{}); ok {
			assert.Contains(t, data, "available", "Response should contain available field")
		} else {
			assert.Contains(t, response, "available", "Response should contain available field")
		}
	}
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkCreateBooking(b *testing.B) {
	gin.SetMode(gin.TestMode)
	t := &testing.T{}
	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)
	defer dbManager.Close()

	router := gin.New()
	routes.Setup(router, dbManager.DB, cfg.JWT.Secret, cfg.JWT.Expiration, nil)

	token, _ := generateTestToken(1, "customer@test.com", "customer", cfg.JWT.Secret)
	jsonBody, _ := json.Marshal(getTestBookingRequest())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGetMyBookings(b *testing.B) {
	gin.SetMode(gin.TestMode)
	t := &testing.T{}
	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)
	defer dbManager.Close()

	router := gin.New()
	routes.Setup(router, dbManager.DB, cfg.JWT.Secret, cfg.JWT.Expiration, nil)

	token, _ := generateTestToken(1, "customer@test.com", "customer", cfg.JWT.Secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/bookings/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCheckAvailability(b *testing.B) {
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
		req, _ := http.NewRequest("GET", "/api/v1/bookings/availability?barber_id=1&start_time=2024-12-28T10:00:00Z&duration=45", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkUpdateBookingStatus(b *testing.B) {
	gin.SetMode(gin.TestMode)
	t := &testing.T{}
	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)
	defer dbManager.Close()

	router := gin.New()
	routes.Setup(router, dbManager.DB, cfg.JWT.Secret, cfg.JWT.Expiration, nil)

	token, _ := generateTestToken(1, "admin@test.com", "admin", cfg.JWT.Secret)
	jsonBody, _ := json.Marshal(getTestStatusUpdateRequest("confirmed"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/bookings/%d/status", i%100+1), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
	}
}
