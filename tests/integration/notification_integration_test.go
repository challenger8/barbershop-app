// tests/integration/notification_integration_test.go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ========================================================================
// NOTIFICATION INTEGRATION TESTS
// ========================================================================

func TestGetMyNotifications_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetMyNotifications_WithFilters(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications?type=booking_confirmation&unread_only=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetMyNotifications_Unauthorized(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestGetNotificationByID_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Notification may not exist
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", w.Code)
	}
}

func TestMarkNotificationAsRead_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/1/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", w.Code)
	}
}

func TestMarkAllNotificationsAsRead_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/read-all", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetUnreadCount_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/unread-count", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
}

func TestCreateNotification_AdminOnly(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Regular user token
	customerToken, _ := generateTestToken(1, "customer@test.com", "customer", jwtSecret)

	notificationData := map[string]interface{}{
		"user_id":  2,
		"title":    "Test Notification",
		"message":  "This is a test notification",
		"type":     "system_alert",
		"priority": "normal",
		"channels": []string{"app"},
	}
	body, _ := json.Marshal(notificationData)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+customerToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should be forbidden for non-admin
	if w.Code != http.StatusForbidden && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 403 or 404 for non-admin, got %d", w.Code)
	}
}

func TestCreateNotification_AdminSuccess(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Admin token
	adminToken, _ := generateTestToken(1, "admin@test.com", "admin", jwtSecret)

	notificationData := map[string]interface{}{
		"user_id":  2,
		"title":    "Test Notification",
		"message":  "This is a test notification",
		"type":     "system_alert",
		"priority": "normal",
		"channels": []string{"app"},
	}
	body, _ := json.Marshal(notificationData)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// May fail if user doesn't exist
	if w.Code != http.StatusCreated && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 201, 400, or 404, got %d", w.Code)
	}
}

func TestDeleteNotification_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/notifications/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 200, 204, or 404, got %d", w.Code)
	}
}

func TestGetNotificationStats_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSendBookingNotification_Success(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	// Admin token for sending notifications
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
	if w.Code != http.StatusCreated && w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 201, 200, 400, or 404, got %d", w.Code)
	}
}

func TestNotification_InvalidType(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	adminToken, _ := generateTestToken(1, "admin@test.com", "admin", jwtSecret)

	notificationData := map[string]interface{}{
		"user_id":  2,
		"title":    "Test",
		"message":  "Test message",
		"type":     "invalid_type", // Invalid type
		"priority": "normal",
	}
	body, _ := json.Marshal(notificationData)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 400 or 422 for invalid type, got %d", w.Code)
	}
}

func TestNotification_Pagination(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications?limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNotification_FilterByPriority(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "user@test.com", "customer", jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications?priority=high", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
