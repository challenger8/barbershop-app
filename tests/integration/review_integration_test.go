// tests/integration/review_integration_test.go
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
// REVIEW INTEGRATION TESTS - TABLE DRIVEN
// =============================================================================

// TestCreateReview consolidates all create review tests
func TestCreateReview(t *testing.T) {
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
			name: "Success",
			payload: map[string]interface{}{
				"booking_id":      1,
				"overall_rating":  5,
				"title":           "Great haircut!",
				"comment":         "The barber was very professional and skilled.",
				"would_recommend": true,
			},
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusNotFound, http.StatusConflict},
		},
		{
			name: "InvalidRating_TooHigh",
			payload: map[string]interface{}{
				"booking_id":     1,
				"overall_rating": 6, // Invalid: must be 1-5
			},
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusUnprocessableEntity},
		},
		{
			name: "InvalidRating_TooLow",
			payload: map[string]interface{}{
				"booking_id":     1,
				"overall_rating": 0, // Invalid: must be 1-5
			},
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusUnprocessableEntity},
		},
		{
			name: "MissingBookingID",
			payload: map[string]interface{}{
				"overall_rating": 5,
			},
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest},
		},
		{
			name: "MissingRating",
			payload: map[string]interface{}{
				"booking_id": 1,
			},
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest},
		},
		{
			name: "Unauthorized",
			payload: map[string]interface{}{
				"booking_id":     1,
				"overall_rating": 5,
			},
			userType:       "customer",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name: "WithAllRatings",
			payload: map[string]interface{}{
				"booking_id":             1,
				"overall_rating":         5,
				"service_quality_rating": 4,
				"punctuality_rating":     5,
				"cleanliness_rating":     5,
				"value_for_money_rating": 4,
				"professionalism_rating": 5,
				"title":                  "Excellent service!",
				"comment":                "Very professional",
				"would_recommend":        true,
			},
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusCreated, http.StatusBadRequest, http.StatusNotFound, http.StatusConflict},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(body))
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

// TestGetReview consolidates all get review tests
func TestGetReview(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		reviewID       string
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
			req := httptest.NewRequest(http.MethodGet, "/api/v1/reviews/"+tt.reviewID, nil)

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

// TestUpdateReview consolidates all update review tests
func TestUpdateReview(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		reviewID       string
		payload        map[string]interface{}
		userID         int
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{
			name:     "Success_OrNotFound",
			reviewID: "1",
			payload: map[string]interface{}{
				"overall_rating": 4,
				"comment":        "Updated review comment",
			},
			userID:         1,
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound, http.StatusForbidden, http.StatusBadRequest},
		},
		{
			name:     "NotFound",
			reviewID: "99999",
			payload: map[string]interface{}{
				"overall_rating": 4,
			},
			userID:         1,
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusNotFound},
		},
		{
			name:     "NotOwner",
			reviewID: "1",
			payload: map[string]interface{}{
				"overall_rating": 4,
			},
			userID:         999,
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusForbidden, http.StatusNotFound},
		},
		{
			name:     "Unauthorized",
			reviewID: "1",
			payload: map[string]interface{}{
				"overall_rating": 4,
			},
			userID:         1,
			userType:       "customer",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name:     "InvalidRating",
			reviewID: "1",
			payload: map[string]interface{}{
				"overall_rating": 10,
			},
			userID:         1,
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/reviews/"+tt.reviewID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.hasAuth {
				token, _ := generateTestToken(tt.userID, tt.userType+"@test.com", tt.userType, jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestDeleteReview consolidates all delete review tests
func TestDeleteReview(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		reviewID       string
		userID         int
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Success_OrNotFound", "1", 1, "customer", true, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound}},
		{"NotFound", "99999", 1, "customer", true, []int{http.StatusNotFound}},
		{"NotOwner", "1", 999, "customer", true, []int{http.StatusForbidden, http.StatusNotFound}},
		{"Unauthorized", "1", 1, "customer", false, []int{http.StatusUnauthorized}},
		{"Admin_CanDelete", "1", 1, "admin", true, []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/"+tt.reviewID, nil)

			if tt.hasAuth {
				token, _ := generateTestToken(tt.userID, tt.userType+"@test.com", tt.userType, jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestGetBarberReviews consolidates all barber reviews tests
func TestGetBarberReviews(t *testing.T) {
	router, dbManager, _ := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus []int
	}{
		// Get reviews
		{"Success", "/api/v1/barbers/1/reviews", []int{http.StatusOK, http.StatusNotFound}},
		{"NotFound", "/api/v1/barbers/99999/reviews", []int{http.StatusOK, http.StatusNotFound}},
		{"WithMinRating", "/api/v1/barbers/1/reviews?min_rating=4", []int{http.StatusOK, http.StatusNotFound}},
		{"WithSorting", "/api/v1/barbers/1/reviews?sort_by=created_at&order=desc", []int{http.StatusOK, http.StatusNotFound}},
		{"WithPagination", "/api/v1/barbers/1/reviews?limit=10&offset=0", []int{http.StatusOK, http.StatusNotFound}},
		{"WithAllFilters", "/api/v1/barbers/1/reviews?min_rating=3&sort_by=rating&order=desc&limit=5", []int{http.StatusOK, http.StatusNotFound}},

		// Get stats
		{"Stats_Success", "/api/v1/barbers/1/reviews/stats", []int{http.StatusOK, http.StatusNotFound}},
		{"Stats_NotFound", "/api/v1/barbers/99999/reviews/stats", []int{http.StatusOK, http.StatusNotFound}},

		// Invalid barber ID
		{"InvalidBarberID", "/api/v1/barbers/abc/reviews", []int{http.StatusBadRequest, http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code,
				"Expected one of %v, got %d", tt.expectedStatus, w.Code)
		})
	}
}

// TestModerateReview consolidates moderation tests
func TestModerateReview(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		reviewID       string
		payload        map[string]interface{}
		userType       string
		hasAuth        bool
		expectedStatus []int
	}{
		{
			name:     "Admin_Approve",
			reviewID: "1",
			payload: map[string]interface{}{
				"status": "approved",
				"notes":  "Review meets guidelines",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
		{
			name:     "Admin_Reject",
			reviewID: "1",
			payload: map[string]interface{}{
				"status": "rejected",
				"notes":  "Inappropriate content",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
		{
			name:     "Customer_Forbidden",
			reviewID: "1",
			payload: map[string]interface{}{
				"status": "approved",
			},
			userType:       "customer",
			hasAuth:        true,
			expectedStatus: []int{http.StatusForbidden, http.StatusNotFound},
		},
		{
			name:     "NotFound",
			reviewID: "99999",
			payload: map[string]interface{}{
				"status": "approved",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusNotFound},
		},
		{
			name:     "Unauthorized",
			reviewID: "1",
			payload: map[string]interface{}{
				"status": "approved",
			},
			userType:       "admin",
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
		{
			name:     "InvalidStatus",
			reviewID: "1",
			payload: map[string]interface{}{
				"status": "invalid_status",
			},
			userType:       "admin",
			hasAuth:        true,
			expectedStatus: []int{http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews/"+tt.reviewID+"/moderate", bytes.NewReader(body))
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

// TestVoteReview consolidates vote tests
func TestVoteReview(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		reviewID       string
		payload        map[string]interface{}
		hasAuth        bool
		expectedStatus []int
	}{
		{
			name:           "Helpful",
			reviewID:       "1",
			payload:        map[string]interface{}{"is_helpful": true},
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
		{
			name:           "NotHelpful",
			reviewID:       "1",
			payload:        map[string]interface{}{"is_helpful": false},
			hasAuth:        true,
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
		{
			name:           "NotFound",
			reviewID:       "99999",
			payload:        map[string]interface{}{"is_helpful": true},
			hasAuth:        true,
			expectedStatus: []int{http.StatusNotFound},
		},
		{
			name:           "Unauthorized",
			reviewID:       "1",
			payload:        map[string]interface{}{"is_helpful": true},
			hasAuth:        false,
			expectedStatus: []int{http.StatusUnauthorized},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews/"+tt.reviewID+"/vote", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

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

// TestGetMyReviews tests the my reviews endpoint
func TestGetMyReviews(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	tests := []struct {
		name           string
		queryParams    string
		userType       string
		hasAuth        bool
		expectedStatus int
	}{
		{"Success", "", "customer", true, http.StatusOK},
		{"WithPagination", "?limit=10&offset=0", "customer", true, http.StatusOK},
		{"Unauthorized", "", "customer", false, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/reviews/me"+tt.queryParams, nil)

			if tt.hasAuth {
				token, _ := generateTestToken(1, tt.userType+"@test.com", tt.userType, jwtSecret)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestCanReviewBooking tests the can-review check endpoint
func TestCanReviewBooking(t *testing.T) {
	router, dbManager, jwtSecret := setupTestRouter(t)
	defer dbManager.Close()

	token, err := generateTestToken(1, "customer@test.com", "customer", jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		bookingID      string
		hasAuth        bool
		expectedStatus []int
	}{
		{"Success_OrNotFound", "1", true, []int{http.StatusOK, http.StatusNotFound}},
		{"NotFound", "99999", true, []int{http.StatusOK, http.StatusNotFound}},
		{"Unauthorized", "1", false, []int{http.StatusUnauthorized}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/reviews/can-review/"+tt.bookingID, nil)

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
