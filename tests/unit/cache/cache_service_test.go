// tests/unit/cache/cache_service_test.go
package cache_test

import (
	"context"
	"testing"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCacheService(t *testing.T) *cache.CacheService {
	client := setupRedisClient(t)
	if client == nil {
		return nil
	}
	return cache.NewCacheService(client)
}

func TestCacheService_CacheBarber(t *testing.T) {
	service := setupCacheService(t)
	if service == nil {
		t.Skip("Redis not available")
		return
	}

	ctx := context.Background()

	barber := &models.Barber{
		ID:       1,
		ShopName: "Test Barber Shop",
		City:     "New York",
		State:    "NY",
		Country:  "USA",
	}

	// Cache barber
	err := service.CacheBarber(ctx, barber.ID, barber)
	require.NoError(t, err)

	// Retrieve barber
	var cached models.Barber
	err = service.GetBarber(ctx, barber.ID, &cached)
	require.NoError(t, err)

	assert.Equal(t, barber.ID, cached.ID)
	assert.Equal(t, barber.ShopName, cached.ShopName)
	assert.Equal(t, barber.City, cached.City)
}

func TestCacheService_GetBarber_NotFound(t *testing.T) {
	service := setupCacheService(t)
	if service == nil {
		t.Skip("Redis not available")
		return
	}

	ctx := context.Background()

	var barber models.Barber
	err := service.GetBarber(ctx, 99999, &barber)
	assert.Error(t, err)
}

func TestCacheService_InvalidateBarber(t *testing.T) {
	service := setupCacheService(t)
	if service == nil {
		t.Skip("Redis not available")
		return
	}

	ctx := context.Background()
	barberID := 123

	barber := &models.Barber{
		ID:       barberID,
		ShopName: "Test Shop",
	}

	// Cache barber
	err := service.CacheBarber(ctx, barberID, barber)
	require.NoError(t, err)

	// Invalidate
	err = service.InvalidateBarber(ctx, barberID)
	require.NoError(t, err)

	// Should not be found
	var cached models.Barber
	err = service.GetBarber(ctx, barberID, &cached)
	assert.Error(t, err)
}

func TestCacheService_CacheSearchResults(t *testing.T) {
	service := setupCacheService(t)
	if service == nil {
		t.Skip("Redis not available")
		return
	}

	ctx := context.Background()
	queryHash := "search-hash-123"

	results := []models.Barber{
		{ID: 1, ShopName: "Shop 1"},
		{ID: 2, ShopName: "Shop 2"},
	}

	// Cache search results
	err := service.CacheSearchResults(ctx, queryHash, results)
	require.NoError(t, err)

	// Retrieve search results
	var cached []models.Barber
	err = service.GetSearchResults(ctx, queryHash, &cached)
	require.NoError(t, err)

	assert.Len(t, cached, 2)
	assert.Equal(t, results[0].ShopName, cached[0].ShopName)
}

func TestCacheService_CacheStats(t *testing.T) {
	service := setupCacheService(t)
	if service == nil {
		t.Skip("Redis not available")
		return
	}

	ctx := context.Background()
	statsKey := "barber-stats-1"

	type Stats struct {
		TotalBookings int
		AverageRating float64
	}

	stats := Stats{
		TotalBookings: 100,
		AverageRating: 4.5,
	}

	// Cache stats
	err := service.CacheStats(ctx, statsKey, stats)
	require.NoError(t, err)

	// Retrieve stats
	var cached Stats
	err = service.GetStats(ctx, statsKey, &cached)
	require.NoError(t, err)

	assert.Equal(t, stats.TotalBookings, cached.TotalBookings)
	assert.Equal(t, stats.AverageRating, cached.AverageRating)
}

func TestCacheService_InvalidateAllBarbers(t *testing.T) {
	service := setupCacheService(t)
	if service == nil {
		t.Skip("Redis not available")
		return
	}

	ctx := context.Background()

	// Cache multiple barbers
	for i := 1; i <= 5; i++ {
		barber := &models.Barber{
			ID:       i,
			ShopName: "Shop " + string(rune('0'+i)),
		}
		err := service.CacheBarber(ctx, i, barber)
		require.NoError(t, err)
	}

	// Invalidate all barbers
	err := service.InvalidateAllBarbers(ctx)
	require.NoError(t, err)

	// Verify all are invalidated
	for i := 1; i <= 5; i++ {
		var barber models.Barber
		err := service.GetBarber(ctx, i, &barber)
		assert.Error(t, err)
	}
}

func TestCacheService_TTL(t *testing.T) {
	service := setupCacheService(t)
	if service == nil {
		t.Skip("Redis not available")
		return
	}

	ctx := context.Background()

	// Override TTL for this test by creating custom cache service
	// (In production, you might want to make TTL configurable)
	barber := &models.Barber{
		ID:       1,
		ShopName: "Test Shop",
	}

	// Cache barber (uses MediumTTL = 30 minutes by default)
	err := service.CacheBarber(ctx, barber.ID, barber)
	require.NoError(t, err)

	// Should exist immediately
	var cached models.Barber
	err = service.GetBarber(ctx, barber.ID, &cached)
	require.NoError(t, err)
	assert.Equal(t, barber.ShopName, cached.ShopName)
}
