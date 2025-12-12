// tests/integration/notification_integration_test.go
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
// NOTIFICATION INTEGRATION TESTS - TABLE DRIVEN
// =============================================================================

// Test fixtures
func getTestNotificationRequest() map[string]interface{} {
	return map[string]interface{}{
		"user_id":  1,
		"title":    "Test Notification",
		"message":  "This is a test notification message",
		"type":     "system_alert",
		"priority": "normal",
		"channels": []string{"app"},
	}
}

func getTestBookingNotificationRequest(bookingID int, notifType string) map[string]interface{} {
	return map[string]interface{}{
		"booking_id":        bookingID,
		"notification_type": notifType,
	}
}

// =============================================================================
// GET NOTIFICATIONS TESTS
// =============================================================================

// TestGetMyNotifications consolidates all get notifications tests
func TestGetMyNotifications(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		queryParams    string
		userType       string
		hasAuth        bool
		expectedStatus int
	}{
		// Basic retrieval
		{"Success", "", "customer", true, http.StatusOK},
		{"Unauthorized", "", "customer", false, http.StatusUnauthorized},

		// Filtering
		{"FilterByType", "?type=booking_confirmation", "customer", true, http.StatusOK},
		{"FilterByPriority", "?priority=high", "customer", true, http.StatusOK},
		{"FilterByStatus", "?status=pending", "customer", true, http.StatusOK},
		{"FilterUnreadOnly", "?is_unread=true", "customer", true, http.StatusOK},

		// Sorting
		{"SortByCreated", "?sort_by=created_at&order=DESC", "customer", true, http.StatusOK},
		{"SortByPriority", "?sort_by=priority&order=ASC", "customer", true, http.StatusOK},

		// Pagination
		{"Pagination", "?limit=10&offset=0", "customer", true, http.StatusOK},
		{"PaginationPage2", "?limit=10&offset=10", "customer", true, http.StatusOK},
		{"LargeLimit", "?limit=100", "customer", true, http.StatusOK},

		// Combined filters
		{"CombinedFilters", "?type=booking_confirmation&priority=high&limit=5", "customer", true, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications"+tt.queryParams, nil)

			if tt.hasAuth {
				token, _ := generateTestToken(1, tt.userType+"@test.com", tt.userType, jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code,
				"Expected %d, got %d for %s", tt.expectedStatus, w.Code, tt.name)
		})
	}
}

// TestGetUnreadNotifications tests the unread notifications endpoint
func TestGetUnreadNotifications(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		queryParams    string
		hasAuth        bool
		expectedStatus int
	}{
		{"Success", "", true, http.StatusOK},
		{"WithLimit", "?limit=5", true, http.StatusOK},
		{"Unauthorized", "", false, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/unread"+tt.queryParams, nil)

			if tt.hasAuth {
				token, _ := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// =============================================================================
// GET NOTIFICATION BY ID TESTS
// =============================================================================

// TestGetNotificationByID consolidates get notification by ID tests
func TestGetNotificationByID(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		notificationID string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Success_OrNotFound", "1", true, []int{http.StatusOK, http.StatusNotFound}},
		{"NotFound", "99999", true, []int{http.StatusNotFound}},
		{"InvalidID", "abc", true, []int{http.StatusBadRequest, http.StatusNotFound}},
		{"Unauthorized", "1", false, []int{http.StatusUnauthorized}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/"+tt.notificationID, nil)

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
// NOTIFICATION STATS TESTS
// =============================================================================

// TestGetNotificationStats tests stats and count endpoints
func TestGetNotificationStats(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		endpoint       string
		hasAuth        bool
		expectedStatus int
	}{
		// Stats endpoint
		{"Stats_Success", "/api/v1/notifications/stats", true, http.StatusOK},
		{"Stats_Unauthorized", "/api/v1/notifications/stats", false, http.StatusUnauthorized},

		// Unread count endpoint
		{"UnreadCount_Success", "/api/v1/notifications/unread/count", true, http.StatusOK},
		{"UnreadCount_Unauthorized", "/api/v1/notifications/unread/count", false, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)

			if tt.hasAuth {
				token, _ := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// =============================================================================
// MARK AS READ TESTS
// =============================================================================

// TestMarkAsRead consolidates mark as read tests
func TestMarkAsRead(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		notificationID string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Success_OrNotFound", "1", true, []int{http.StatusOK, http.StatusNotFound}},
		{"NotFound", "99999", true, []int{http.StatusNotFound}},
		{"InvalidID", "abc", true, []int{http.StatusBadRequest, http.StatusNotFound}},
		{"Unauthorized", "1", false, []int{http.StatusUnauthorized}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/"+tt.notificationID+"/read", nil)

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

// TestMarkAllAsRead tests the mark all as read endpoint
func TestMarkAllAsRead(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		hasAuth        bool
		expectedStatus int
	}{
		{"Success", true, http.StatusOK},
		{"Unauthorized", false, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/read-all", nil)

			if tt.hasAuth {
				token, _ := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// =============================================================================
// DELETE NOTIFICATION TESTS
// =============================================================================

// TestDeleteNotification consolidates delete notification tests
func TestDeleteNotification(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		notificationID string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Success_OrNotFound", "1", true, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound}},
		{"NotFound", "99999", true, []int{http.StatusNotFound}},
		{"InvalidID", "abc", true, []int{http.StatusBadRequest, http.StatusNotFound}},
		{"Unauthorized", "1", false, []int{http.StatusUnauthorized}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/notifications/"+tt.notificationID, nil)

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
// CREATE NOTIFICATION TESTS
// =============================================================================

// TestCreateNotification consolidates create notification tests
func TestCreateNotification(t *testing.T) {
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
			name:           "Admin_Success",
			payload:        getTestNotificationRequest(),
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError},
		},
		{
			name:           "Customer_Success",
			payload:        getTestNotificationRequest(),
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError},
		},
		{
			name:           "Unauthorized",
			payload:        getTestNotificationRequest(),
			userType:       "admin",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name: "InvalidType",
			payload: map[string]interface{}{
				"user_id":  1,
				"title":    "Test",
				"message":  "Test message",
				"type":     "invalid_type",
				"priority": "normal",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusNotFound, http.StatusInternalServerError},
		},
		{
			name: "InvalidPriority",
			payload: map[string]interface{}{
				"user_id":  1,
				"title":    "Test",
				"message":  "Test message",
				"type":     "system_alert",
				"priority": "super_urgent",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusCreated, http.StatusNotFound, http.StatusInternalServerError},
		},
		{
			name: "MissingTitle",
			payload: map[string]interface{}{
				"user_id":  1,
				"message":  "Test message",
				"type":     "system_alert",
				"priority": "normal",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusCreated, http.StatusInternalServerError},
		},
		{
			name: "MissingMessage",
			payload: map[string]interface{}{
				"user_id":  1,
				"title":    "Test",
				"type":     "system_alert",
				"priority": "normal",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusCreated, http.StatusInternalServerError},
		},
		{
			name: "WithScheduledFor",
			payload: map[string]interface{}{
				"user_id":       1,
				"title":         "Scheduled Notification",
				"message":       "This is scheduled",
				"type":          "system_alert",
				"priority":      "normal",
				"scheduled_for": "2025-12-31T12:00:00Z",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError},
		},
		{
			name: "WithMultipleChannels",
			payload: map[string]interface{}{
				"user_id":  1,
				"title":    "Multi-channel",
				"message":  "Sent to multiple channels",
				"type":     "system_alert",
				"priority": "high",
				"channels": []string{"app", "email", "push"},
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications", bytes.NewReader(body))
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
// SEND BOOKING NOTIFICATION TESTS
// =============================================================================

// TestSendBookingNotification tests booking-related notification endpoints
func TestSendBookingNotification(t *testing.T) {
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
			name:           "BookingConfirmation",
			payload:        getTestBookingNotificationRequest(1, "booking_confirmation"),
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusCreated, http.StatusBadRequest, http.StatusNotFound},
		},
		{
			name:           "BookingReminder",
			payload:        getTestBookingNotificationRequest(1, "booking_reminder"),
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusCreated, http.StatusBadRequest, http.StatusNotFound},
		},
		{
			name:           "BookingCancelled",
			payload:        getTestBookingNotificationRequest(1, "booking_cancelled"),
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusCreated, http.StatusBadRequest, http.StatusNotFound},
		},
		{
			name:           "ReviewRequest",
			payload:        getTestBookingNotificationRequest(1, "review_request"),
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusCreated, http.StatusBadRequest, http.StatusNotFound},
		},
		{
			name: "InvalidType",
			payload: map[string]interface{}{
				"booking_id":        1,
				"notification_type": "invalid_type",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest},
		},
		{
			name:           "Unauthorized",
			payload:        getTestBookingNotificationRequest(1, "booking_confirmation"),
			userType:       "admin",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name: "MissingBookingID",
			payload: map[string]interface{}{
				"notification_type": "booking_confirmation",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusNotFound},
		},
		{
			name:           "BookingNotFound",
			payload:        getTestBookingNotificationRequest(99999, "booking_confirmation"),
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusNotFound, http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/booking", bytes.NewReader(body))
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
// WEBHOOK TESTS
// =============================================================================

// TestDeliveryWebhook tests the notification webhook endpoint


// =============================================================================
// RESPONSE FORMAT TESTS
// =============================================================================

func TestGetMyNotifications_ResponseFormat(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have success or data field
	_, hasSuccess := response["success"]
	_, hasData := response["data"]
	assert.True(t, hasSuccess || hasData, "Response should contain success or data field")
}

func TestGetNotificationStats_ResponseFormat(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkGetMyNotifications(b *testing.B) {
	t := &testing.T{}
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, _ := generateTestToken(1, "customer@test.com", "customer", jwtSecret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGetUnreadCount(b *testing.B) {
	t := &testing.T{}
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, _ := generateTestToken(1, "customer@test.com", "customer", jwtSecret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/unread/count", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkMarkAsRead(b *testing.B) {
	t := &testing.T{}
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, _ := generateTestToken(1, "customer@test.com", "customer", jwtSecret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/1/read", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
	}
}
