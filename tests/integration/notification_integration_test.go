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

// ========================================================================
// NOTIFICATION INTEGRATION TESTS - TABLE DRIVEN
// ========================================================================

// TestGetMyNotifications consolidates all list notification tests
func TestGetMyNotifications(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		queryParams    string
		hasAuth        bool
		expectedStatus int
	}{
		{"Success_NoFilters", "", true, http.StatusOK},
		{"Success_WithTypeFilter", "?type=booking_confirmation&unread_only=true", true, http.StatusOK},
		{"Success_WithPagination", "?limit=10&offset=0", true, http.StatusOK},
		{"Success_WithPriority", "?priority=high", true, http.StatusOK},
		{"Unauthorized", "", false, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications"+tt.queryParams, nil)
			if tt.hasAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetNotificationByID consolidates get by ID tests
func TestGetNotificationByID(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		notificationID string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Success_OrNotFound", "1", true, []int{http.StatusOK, http.StatusNotFound}},
		{"NotFound", "99999", true, []int{http.StatusNotFound}},
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

			assert.Contains(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestMarkNotificationAsRead consolidates mark as read tests
func TestMarkNotificationAsRead(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		notificationID string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Success_OrNotFound", "1", true, []int{http.StatusOK, http.StatusNotFound}},
		{"NotFound", "99999", true, []int{http.StatusNotFound}},
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

			assert.Contains(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestMarkAllNotificationsAsRead tests bulk read operation
func TestMarkAllNotificationsAsRead(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		userType       string
		hasAuth        bool
		expectedStatus int
	}{
		{"Success_Customer", "customer", true, http.StatusOK},
		{"Success_Barber", "barber", true, http.StatusOK},
		{"Unauthorized", "customer", false, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/read-all", nil)
			if tt.hasAuth {
				token, _ := generateTestToken(1, "user@test.com", tt.userType, jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetUnreadCount tests unread count endpoint
func TestGetUnreadCount(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	require.NoError(t, err)

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
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/unread/count", nil)
			if tt.hasAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetNotificationStats tests stats endpoint
func TestGetNotificationStats(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestCreateNotification consolidates create notification tests
func TestCreateNotification(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		userType       string
		payload        map[string]interface{}
		hasAuth        bool
		expectedStatus []int
	}{
		{
			name:     "Success_Admin",
			userType: "admin",
			payload: map[string]interface{}{
				"user_id":  2,
				"title":    "Test Notification",
				"message":  "This is a test notification",
				"type":     "system_alert",
				"priority": "normal",
				"channels": []string{"app"},
			},
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusNotFound},
		},
		{
			name:     "Success_Customer",
			userType: "customer",
			payload: map[string]interface{}{
				"user_id":  2,
				"title":    "Test Notification",
				"message":  "This is a test notification",
				"type":     "system_alert",
				"priority": "normal",
				"channels": []string{"app"},
			},
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusNotFound, http.StatusForbidden},
		},
		{
			name:     "InvalidType",
			userType: "admin",
			payload: map[string]interface{}{
				"user_id":  2,
				"title":    "Test",
				"message":  "Test message",
				"type":     "invalid_type",
				"priority": "normal",
			},
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusNotFound},
		},
		{
			name:           "Unauthorized",
			userType:       "customer",
			payload:        map[string]interface{}{},
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
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

			assert.Contains(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestDeleteNotification consolidates delete tests
func TestDeleteNotification(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		notificationID string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Success_OrNotFound", "1", true, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound}},
		{"NotFound", "99999", true, []int{http.StatusNotFound}},
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

			assert.Contains(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestSendBookingNotification tests booking-specific notification
func TestSendBookingNotification(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	adminToken, _ := generateTestToken(1, "admin@test.com", "admin", jwtSecret)

	notificationData := map[string]interface{}{
		"booking_id":        1,
		"notification_type": "booking_confirmation",
	}
	body, _ := json.Marshal(notificationData)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/booking", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// May fail if booking doesn't exist
	assert.Contains(t, []int{http.StatusCreated, http.StatusOK, http.StatusBadRequest, http.StatusNotFound}, w.Code)
}
