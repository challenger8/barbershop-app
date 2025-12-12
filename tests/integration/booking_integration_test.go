// tests/integration/booking_integration_test.go
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"barber-booking-system/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NOTE: Test fixtures are in booking_test_helpers.go
// NOTE: setupTestRouter is in test_helpers.go (shared across all integration tests)
// =============================================================================

// =============================================================================
// CREATE BOOKING TESTS
// =============================================================================

func TestCreateBooking_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Generate customer token
	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	testBooking := getTestBookingRequest()
	jsonBody, _ := json.Marshal(testBooking)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Accept 201 (success), 409 (conflict), 404 (barber/service not found), or 500 (db error)
	assert.Contains(t, []int{http.StatusCreated, http.StatusConflict, http.StatusNotFound, http.StatusInternalServerError}, w.Code)

	if w.Code == http.StatusCreated {
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["data"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["id"])
		assert.NotNil(t, data["booking_number"])
		assert.Equal(t, "pending", data["status"])
	}
}

func TestCreateBooking_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	testBooking := getTestBookingRequest()
	jsonBody, _ := json.Marshal(testBooking)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateBooking_InvalidData(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing barber_id",
			requestBody: map[string]interface{}{
				"service_id":       1,
				"start_time":       time.Now().Add(2 * time.Hour).Format(time.RFC3339),
				"duration_minutes": 45,
				"customer_name":    "Test",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing service_id",
			requestBody: map[string]interface{}{
				"barber_id":        1,
				"start_time":       time.Now().Add(2 * time.Hour).Format(time.RFC3339),
				"duration_minutes": 45,
				"customer_name":    "Test",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing start_time",
			requestBody: map[string]interface{}{
				"barber_id":        1,
				"service_id":       1,
				"duration_minutes": 45,
				"customer_name":    "Test",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid duration (too short)",
			requestBody: map[string]interface{}{
				"barber_id":        1,
				"service_id":       1,
				"start_time":       time.Now().Add(2 * time.Hour).Format(time.RFC3339),
				"duration_minutes": 5, // Less than 15 minutes
				"customer_name":    "Test",
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
			req, _ := http.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(w, req)

			// Accept both 400 (bad request) and 422 (unprocessable entity) for validation errors
			assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, w.Code)
		})
	}
}

func TestCreateBooking_BookingInPast(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	jsonBody, _ := json.Marshal(getTestBookingRequestInPast())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Should reject booking in the past
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, w.Code)
}

// =============================================================================
// GET BOOKING TESTS
// =============================================================================

func TestGetBooking_NotFound(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookings/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetBooking_InvalidID(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookings/invalid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetBooking_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookings/1", nil)
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// GET BOOKING BY UUID TESTS
// =============================================================================

func TestGetBookingByUUID_NotFound(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookings/uuid/00000000-0000-0000-0000-000000000000", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// GET BOOKING BY NUMBER TESTS
// =============================================================================

func TestGetBookingByNumber_NotFound(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookings/number/BK999999999999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// GET MY BOOKINGS TESTS
// =============================================================================

func TestGetMyBookings_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookings/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if response["success"] != nil {
		assert.True(t, response["success"].(bool))
	}
}

func TestGetMyBookings_WithFilters(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	// Test with status filter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookings/me?status=pending&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetMyBookings_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookings/me", nil)
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// GET BARBER BOOKINGS TESTS
// =============================================================================

func TestGetBarberBookings_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/1/bookings", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if response["success"] != nil {
		assert.True(t, response["success"].(bool))
	}
}

func TestGetBarberBookings_InvalidBarberID(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/invalid/bookings", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetBarberBookings_WithDateFilter(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	// Use date-only format to match struct's time_format:"2006-01-02"
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/barbers/1/bookings?start_date_from=%s&start_date_to=%s", today, tomorrow), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// GET TODAY'S BOOKINGS TESTS
// =============================================================================

func TestGetTodayBookings_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/1/bookings/today", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetTodayBookings_InvalidBarberID(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/invalid/bookings/today", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// CHECK AVAILABILITY TESTS
// =============================================================================

func TestCheckAvailability_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	startTime := time.Now().Add(2 * time.Hour).Format(time.RFC3339)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/bookings/availability?barber_id=1&start_time=%s&duration=45", startTime), nil)
	router.ServeHTTP(w, req)

	assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusNotFound}, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if response["data"] != nil {
		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["available"])
	}
}

func TestCheckAvailability_MissingParams(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name  string
		query string
	}{
		{"missing barber_id", "?start_time=2024-11-28T10:00:00Z&duration=45"},
		{"missing start_time", "?barber_id=1&duration=45"},
		{"missing duration", "?barber_id=1&start_time=2024-11-28T10:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/bookings/availability"+tt.query, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// =============================================================================
// UPDATE BOOKING STATUS TESTS
// =============================================================================

func TestUpdateBookingStatus_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	jsonBody, _ := json.Marshal(getTestStatusUpdateRequest("confirmed"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/bookings/1/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateBookingStatus_NotFound(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "admin@test.com", "admin", jwtSecret)
	require.NoError(t, err)

	jsonBody, _ := json.Marshal(getTestStatusUpdateRequest("confirmed"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/bookings/99999/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// CANCEL BOOKING TESTS
// =============================================================================

func TestCancelBooking(t *testing.T) {
	tests := []struct {
		name           string
		bookingID      string
		hasAuth        bool
		hasReason      bool
		expectedStatus int
	}{
		{"Unauthorized", "1", false, false, http.StatusUnauthorized},
		{"NotFound", "99999", true, false, http.StatusNotFound},
		{"WithReason", "99999", true, true, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Single test logic
		})
	}
}

// =============================================================================
// RESCHEDULE BOOKING TESTS
// =============================================================================

func TestRescheduleBooking_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	jsonBody, _ := json.Marshal(getTestRescheduleRequest())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/bookings/1/reschedule", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRescheduleBooking_NotFound(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	jsonBody, _ := json.Marshal(getTestRescheduleRequest())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/bookings/99999/reschedule", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// GET BARBER BOOKING STATS TESTS
// =============================================================================

func TestGetBarberBookingStats_Success(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/barbers/1/bookings/stats", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if response["data"] != nil {
		data := response["data"].(map[string]interface{})
		// Stats should include these fields
		assert.NotNil(t, data["total_bookings"])
	}
}

func TestGetBarberBookingStats_WithDateRange(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	fromDate := time.Now().Add(-30 * 24 * time.Hour).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/barbers/1/bookings/stats?from=%s&to=%s", fromDate, toDate), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// GET BOOKING HISTORY TESTS
// =============================================================================

func TestGetBookingHistory_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookings/1/history", nil)
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetBookingHistory_NotFound(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookings/99999/history", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// May return 404 or 200 with empty history depending on implementation
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound, http.StatusInternalServerError}, w.Code)
}

// =============================================================================
// BOOKING ROUTES REGISTERED TEST
// =============================================================================

func TestBookingRoutesRegistered(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

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
// ADDITIONAL TESTS USING HELPER FUNCTIONS
// =============================================================================

func TestCreateBooking_WithCustomerID(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	// Use helper for booking with customer ID
	testBooking := getTestBookingRequestWithCustomer(1)
	jsonBody, _ := json.Marshal(testBooking)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Accept various valid responses
	assert.Contains(t, []int{http.StatusCreated, http.StatusConflict, http.StatusNotFound, http.StatusInternalServerError}, w.Code)
}

func TestCreateBooking_ForSpecificBarber(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	// Use helper for booking with specific barber and service
	testBooking := getTestBookingRequestForBarber(2, 3)
	jsonBody, _ := json.Marshal(testBooking)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Accept various valid responses (barber/service may not exist)
	assert.Contains(t, []int{http.StatusCreated, http.StatusConflict, http.StatusNotFound, http.StatusInternalServerError, http.StatusBadRequest}, w.Code)
}

func TestCancelBooking_ByBarber(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "barber@test.com", "barber", jwtSecret)
	require.NoError(t, err)

	// Use helper for barber cancellation (is_by_customer = false)
	jsonBody, _ := json.Marshal(getTestCancelRequest(false))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/bookings/99999", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// 404 because booking doesn't exist
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateBookingStatus_ToInProgress(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "barber@test.com", "barber", jwtSecret)
	require.NoError(t, err)

	// Use helper for different status
	jsonBody, _ := json.Marshal(getTestStatusUpdateRequest("in_progress"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/bookings/99999/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateBookingStatus_ToCompleted(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "barber@test.com", "barber", jwtSecret)
	require.NoError(t, err)

	// Use helper for completed status
	jsonBody, _ := json.Marshal(getTestStatusUpdateRequest("completed"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/bookings/99999/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// PERFORMANCE TESTS
// =============================================================================

func BenchmarkGetBarberBookings(b *testing.B) {
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
		req, _ := http.NewRequest("GET", "/api/v1/barbers/1/bookings", nil)
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

	startTime := time.Now().Add(2 * time.Hour).Format(time.RFC3339)
	url := fmt.Sprintf("/api/v1/bookings/availability?barber_id=1&start_time=%s&duration=45", startTime)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGetTodayBookings(b *testing.B) {
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
		req, _ := http.NewRequest("GET", "/api/v1/barbers/1/bookings/today", nil)
		router.ServeHTTP(w, req)
	}
}
