// tests/integration/setup_test.go
package integration

import (
	"os"
	"path/filepath"
	"testing"

	"barber-booking-system/config"
	appConfig "barber-booking-system/internal/config"
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

func cleanupTestData(t *testing.T, dbManager *config.DatabaseManager) {
	// Only clean test data (data created during tests)
	// Be very careful with this in production database
	queries := []string{
		"DELETE FROM bookings WHERE booking_number LIKE 'TEST%'",
		"DELETE FROM barbers WHERE shop_name LIKE 'Test%'",
		"DELETE FROM users WHERE email LIKE 'test%@test.com'",
	}

	for _, query := range queries {
		_, err := dbManager.DB.Exec(query)
		if err != nil {
			t.Logf("Warning: cleanup query failed: %v", err)
		}
	}
}
