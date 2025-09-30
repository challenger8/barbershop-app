// tests/integration/server_test.go
package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"barber-booking-system/config"
	"barber-booking-system/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)
	defer dbManager.Close()

	router := gin.New()
	router.GET("/health", config.CreateHealthCheckHandler(dbManager))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "barbershop-api", response["service"])
	assert.NotNil(t, response["timestamp"])
	assert.NotNil(t, response["database"])
}

func TestAllRoutesRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)
	defer dbManager.Close()

	router := gin.New()
	// ✅ Pass nil for cacheService since tests don't need Redis
	routes.Setup(router, dbManager.DB, cfg.JWT.Secret, nil)

	allRoutes := router.Routes()

	expectedRoutes := []string{
		"GET /api/v1/barbers",
		"GET /api/v1/barbers/search",
		"GET /api/v1/barbers/:id",
		"GET /api/v1/barbers/uuid/:uuid",
		"POST /api/v1/barbers",
		"PUT /api/v1/barbers/:id",
		"DELETE /api/v1/barbers/:id",
		"PATCH /api/v1/barbers/:id/status",
		"GET /api/v1/barbers/:id/statistics",
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
