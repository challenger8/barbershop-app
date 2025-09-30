// tests/integration/redis_integration_test.go
package integration

import (
	"context"
	"testing"
	"time"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/models"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBarberService_WithCache(t *testing.T) {
	// Setup database
	cfg := getTestConfig(t)
	dbManager := setupTestDatabase(t, cfg)
	defer dbManager.Close()

	// Setup Redis
	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		t.Skip("Redis not available:", err)
		return
	}
	defer redisClient.Close()

	cacheService := cache.NewCacheService(redisClient)

	// Setup service with cache
	barberRepo := repository.NewBarberRepository(dbManager.DB)
	barberService := services.NewBarberService(barberRepo, cacheService)

	ctx := context.Background()

	// Create a test barber
	barber := &models.Barber{
		UserID:     1,
		ShopName:   "Cached Barber Shop",
		Address:    "123 Test St",
		City:       "Test City",
		State:      "TS",
		Country:    "Test Country",
		PostalCode: "12345",
	}

	err = barberRepo.Create(ctx, barber)
	require.NoError(t, err)
	require.NotZero(t, barber.ID)

	// First call - should hit database
	start := time.Now()
	retrieved1, err := barberService.GetBarber(ctx, barber.ID)
	dbDuration := time.Since(start)
	require.NoError(t, err)
	assert.Equal(t, barber.ShopName, retrieved1.ShopName)

	// Second call - should hit cache (faster)
	start = time.Now()
	retrieved2, err := barberService.GetBarber(ctx, barber.ID)
	cacheDuration := time.Since(start)
	require.NoError(t, err)
	assert.Equal(t, barber.ShopName, retrieved2.ShopName)

	// Cache should be faster (usually 10-100x faster)
	t.Logf("DB query: %v, Cache query: %v", dbDuration, cacheDuration)
	assert.Less(t, cacheDuration, dbDuration)

	// Update barber - should invalidate cache
	barber.ShopName = "Updated Shop Name"
	err = barberService.Update(ctx, barber)
	require.NoError(t, err)

	// Next call should get updated data
	retrieved3, err := barberService.GetBarber(ctx, barber.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Shop Name", retrieved3.ShopName)

	// Cleanup
	err = barberRepo.Delete(ctx, barber.ID)
	require.NoError(t, err)
}
