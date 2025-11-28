// tests/integration/setup_test.go
package integration

import (
	"barber-booking-system/internal/middleware"
	"os"
	"path/filepath"
	"testing"
	"time"

	"barber-booking-system/config"
	appConfig "barber-booking-system/internal/config"
	"barber-booking-system/internal/routes"

	"github.com/gin-gonic/gin"
)

func getTestConfig(t *testing.T) *appConfig.Config {
	// Find and load .env file from project root
	// Go up from tests/integration to project root
	projectRoot := filepath.Join("..", "..")
	envPath := filepath.Join(projectRoot, ".env")

	// Check if .env exists
	if _, err := os.Stat(envPath); err == nil {
		// Set working directory context for config loading
		originalWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(originalWd)
	}

	// Load configuration (it will use DATABASE_URL from .env)
	cfg, err := appConfig.Load()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Override only if TEST_DATABASE_URL is set
	if testDBURL := os.Getenv("TEST_DATABASE_URL"); testDBURL != "" {
		cfg.Database.URL = testDBURL
	}

	return cfg
}

func setupTestDatabase(t *testing.T, cfg *appConfig.Config) *config.DatabaseManager {
	dbManager, err := config.NewDatabaseManager(cfg.Database)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Don't clean up data in tests for now
	// cleanupTestData(t, dbManager)

	return dbManager
}

func setupTestRouter(t *testing.T) (*gin.Engine, *config.DatabaseManager, string) {
	gin.SetMode(gin.TestMode)

	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)

	router := gin.New()
	routes.Setup(router, dbManager.DB, cfg.JWT.Secret, cfg.JWT.Expiration, nil)

	return router, dbManager, cfg.JWT.Secret
}

// tests/integration/test_helpers.go

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
