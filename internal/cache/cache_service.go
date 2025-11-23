// internal/cache/cache_service.go
package cache

import (
	"context"
	"fmt"
	"time"
)

// CacheService provides high-level caching operations
type CacheService struct {
	redis *RedisClient
}

// NewCacheService creates a new cache service
func NewCacheService(redis *RedisClient) *CacheService {
	return &CacheService{redis: redis}
}

// Cache key prefixes
const (
	BarberPrefix    = "barber:"
	UserPrefix      = "user:"
	ServicePrefix   = "service:"
	BookingPrefix   = "booking:"
	ReviewPrefix    = "review:"
	SearchPrefix    = "search:"
	StatsPrefix     = "stats:"
	RateLimitPrefix = "ratelimit:"
	SessionPrefix   = "session:"
)

// Default TTLs
const (
	ShortTTL  = 5 * time.Minute
	MediumTTL = 30 * time.Minute
	LongTTL   = 2 * time.Hour
	DayTTL    = 24 * time.Hour
)

// CacheBarber caches a barber object
func (s *CacheService) CacheBarber(ctx context.Context, barberID int, barber interface{}) error {
	key := fmt.Sprintf("%s%d", BarberPrefix, barberID)
	return s.redis.SetJSON(ctx, key, barber, MediumTTL)
}

// GetBarber retrieves a cached barber
func (s *CacheService) GetBarber(ctx context.Context, barberID int, dest interface{}) error {
	key := fmt.Sprintf("%s%d", BarberPrefix, barberID)
	return s.redis.GetJSON(ctx, key, dest)
}

// InvalidateBarber removes a barber from cache
func (s *CacheService) InvalidateBarber(ctx context.Context, barberID int) error {
	key := fmt.Sprintf("%s%d", BarberPrefix, barberID)
	return s.redis.Delete(ctx, key)
}

// CacheSearchResults caches search results
func (s *CacheService) CacheSearchResults(ctx context.Context, queryHash string, results interface{}) error {
	key := fmt.Sprintf("%s%s", SearchPrefix, queryHash)
	return s.redis.SetJSON(ctx, key, results, ShortTTL)
}

// GetSearchResults retrieves cached search results
func (s *CacheService) GetSearchResults(ctx context.Context, queryHash string, dest interface{}) error {
	key := fmt.Sprintf("%s%s", SearchPrefix, queryHash)
	return s.redis.GetJSON(ctx, key, dest)
}

// CacheStats caches statistics
func (s *CacheService) CacheStats(ctx context.Context, statsKey string, stats interface{}) error {
	key := fmt.Sprintf("%s%s", StatsPrefix, statsKey)
	return s.redis.SetJSON(ctx, key, stats, LongTTL)
}

// GetStats retrieves cached statistics
func (s *CacheService) GetStats(ctx context.Context, statsKey string, dest interface{}) error {
	key := fmt.Sprintf("%s%s", StatsPrefix, statsKey)
	return s.redis.GetJSON(ctx, key, dest)
}

// InvalidateAllBarbers clears all barber caches
func (s *CacheService) InvalidateAllBarbers(ctx context.Context) error {
	return s.redis.DeletePattern(ctx, BarberPrefix+"*")
}

// Generic cache operations

// Set stores any value in cache with medium TTL
func (s *CacheService) Set(ctx context.Context, key string, value interface{}) error {
	return s.redis.SetJSON(ctx, key, value, MediumTTL)
}

// SetWithTTL stores any value in cache with custom TTL
func (s *CacheService) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.redis.SetJSON(ctx, key, value, ttl)
}

// Get retrieves a value from cache
func (s *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	return s.redis.GetJSON(ctx, key, dest)
}

// Delete removes a value from cache
func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.redis.Delete(ctx, key)
}

// Exists checks if a key exists in cache
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	return s.redis.Exists(ctx, key)
}
