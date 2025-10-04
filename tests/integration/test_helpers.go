// tests/integration/test_helpers.go
package integration

import (
	"time"

	"barber-booking-system/internal/middleware"
)

// Pointer helper functions for tests
// These make it easy to create pointer values for optional fields

// String pointer helper
func stringPtr(s string) *string {
	return &s
}

// Int pointer helper
func intPtr(i int) *int {
	return &i
}

// Int64 pointer helper
func int64Ptr(i int64) *int64 {
	return &i
}

// Float64 pointer helper
func float64Ptr(f float64) *float64 {
	return &f
}

// Bool pointer helper
func boolPtr(b bool) *bool {
	return &b
}

// Time pointer helper
func timePtr(t time.Time) *time.Time {
	return &t
}

// generateTestToken creates a JWT token for testing
func generateTestToken(userID int, email string, userType string, jwtSecret string) (string, error) {
	return middleware.GenerateToken(userID, email, userType, jwtSecret, 24*time.Hour)
}

// Usage examples:
//
// Creating a barber with optional fields:
//
// barber := &models.Barber{
//     ShopName:    "Test Shop",
//     Description: stringPtr("Great barbershop"),     // *string
//     Latitude:    float64Ptr(40.7128),               // *float64
//     Longitude:   float64Ptr(-74.0060),              // *float64
//     Phone:       stringPtr("+1234567890"),          // *string
//     YearsExperience: intPtr(5),                     // *int
//     IsVerified:  boolPtr(true),                     // *bool (if needed)
// }
//
// Creating an auth token for tests:
//
// token, _ := generateTestToken(123, "test@example.com", "admin", cfg.JWT.Secret)
