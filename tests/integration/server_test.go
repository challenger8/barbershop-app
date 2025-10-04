package integration

import (
	"barber-booking-system/config"
	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/routes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedisClient(t *testing.T) *cache.RedisClient {
	t.Helper()

	// Use config.RedisConfig (not cache.RedisConfig)
	client, err := cache.NewRedisClient(config.RedisConfig{
		URL:      "redis://localhost:6379",
		Password: "",
		DB:       1, // Use test database (different from production DB 0)
	})
	if err != nil {
		t.Logf("Redis not available: %v", err)
		return nil
	}

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

func setupCacheService(t *testing.T) *cache.CacheService {
	t.Helper()

	client := setupRedisClient(t)
	if client == nil {
		return nil
	}

	return cache.NewCacheService(client)
}

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

	cacheService := setupCacheService(t) // Fixed: proper variable declaration
	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)
	defer dbManager.Close()

	router := gin.New()
	routes.Setup(router, dbManager.DB, cfg.JWT.Secret, cfg.JWT.Expiration, cacheService)

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
