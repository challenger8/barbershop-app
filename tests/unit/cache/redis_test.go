// tests/unit/cache/redis_test.go
package cache_test

import (
	"context"
	"testing"
	"time"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestRedisConfig returns a test Redis configuration
func getTestRedisConfig() config.RedisConfig {
	return config.RedisConfig{
		URL:      "redis://localhost:6379",
		Password: "",
		DB:       1, // Use DB 1 for tests to avoid conflicts
	}
}

// setupRedisClient creates a Redis client for testing
func setupRedisClient(t *testing.T) *cache.RedisClient {
	cfg := getTestRedisConfig()
	client, err := cache.NewRedisClient(cfg)
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return nil
	}
	return client
}

// cleanupRedis cleans up test data
func cleanupRedis(t *testing.T, client *cache.RedisClient, keys ...string) {
	if client == nil {
		return
	}
	ctx := context.Background()
	for _, key := range keys {
		_ = client.Delete(ctx, key)
	}
}

func TestNewRedisClient(t *testing.T) {
	cfg := getTestRedisConfig()
	client, err := cache.NewRedisClient(cfg)

	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return
	}

	require.NoError(t, err)
	require.NotNil(t, client)

	defer client.Close()

	// Test connection
	ctx := context.Background()
	err = client.Set(ctx, "test-connection", "ok", time.Minute)
	assert.NoError(t, err)

	// Cleanup
	_ = client.Delete(ctx, "test-connection")
}

func TestNewRedisClient_InvalidURL(t *testing.T) {
	cfg := config.RedisConfig{
		URL:      "invalid-url",
		Password: "",
		DB:       0,
	}

	client, err := cache.NewRedisClient(cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestRedisClient_SetAndGet(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-set-get"
	value := "test-value"

	// Set value
	err := client.Set(ctx, key, value, time.Minute)
	require.NoError(t, err)

	// Get value
	result, err := client.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, result)

	// Cleanup
	cleanupRedis(t, client, key)
}

func TestRedisClient_Get_NotFound(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Try to get non-existent key
	_, err := client.Get(ctx, "non-existent-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestRedisClient_SetJSON(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-set-json"

	type TestData struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	data := TestData{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	// Set JSON
	err := client.SetJSON(ctx, key, data, time.Minute)
	require.NoError(t, err)

	// Get JSON
	var result TestData
	err = client.GetJSON(ctx, key, &result)
	require.NoError(t, err)

	assert.Equal(t, data.Name, result.Name)
	assert.Equal(t, data.Age, result.Age)
	assert.Equal(t, data.Email, result.Email)

	// Cleanup
	cleanupRedis(t, client, key)
}

func TestRedisClient_Delete(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-delete"

	// Set value
	err := client.Set(ctx, key, "value", time.Minute)
	require.NoError(t, err)

	// Delete
	err = client.Delete(ctx, key)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Get(ctx, key)
	assert.Error(t, err)
}

func TestRedisClient_DeleteMultiple(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	keys := []string{"test-del-1", "test-del-2", "test-del-3"}

	// Set multiple values
	for _, key := range keys {
		err := client.Set(ctx, key, "value", time.Minute)
		require.NoError(t, err)
	}

	// Delete all at once
	err := client.Delete(ctx, keys...)
	require.NoError(t, err)

	// Verify all deleted
	for _, key := range keys {
		_, err := client.Get(ctx, key)
		assert.Error(t, err)
	}
}

func TestRedisClient_Exists(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-exists"

	// Key should not exist
	exists, err := client.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Set value
	err = client.Set(ctx, key, "value", time.Minute)
	require.NoError(t, err)

	// Key should exist
	exists, err = client.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Cleanup
	cleanupRedis(t, client, key)
}

func TestRedisClient_Expire(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-expire"

	// Set value without expiration
	err := client.Set(ctx, key, "value", 0)
	require.NoError(t, err)

	// Set expiration - Redis minimum is 1 second
	err = client.Expire(ctx, key, 1*time.Second)
	require.NoError(t, err)

	// Key should exist
	exists, err := client.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(1200 * time.Millisecond)

	// Key should be expired
	exists, err = client.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRedisClient_Increment(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-increment"

	// First increment
	count, err := client.Increment(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Second increment
	count, err = client.Increment(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Third increment
	count, err = client.Increment(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Cleanup
	cleanupRedis(t, client, key)
}

func TestRedisClient_IncrementBy(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-increment-by"

	// Increment by 5
	count, err := client.IncrementBy(ctx, key, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)

	// Increment by 10
	count, err = client.IncrementBy(ctx, key, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(15), count)

	// Cleanup
	cleanupRedis(t, client, key)
}

func TestRedisClient_SetNX(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-setnx"

	// First SetNX should succeed
	success, err := client.SetNX(ctx, key, "value1", time.Minute)
	require.NoError(t, err)
	assert.True(t, success)

	// Second SetNX should fail (key already exists)
	success, err = client.SetNX(ctx, key, "value2", time.Minute)
	require.NoError(t, err)
	assert.False(t, success)

	// Value should still be the first one
	value, err := client.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Cleanup
	cleanupRedis(t, client, key)
}

func TestRedisClient_DeletePattern(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Set multiple keys with pattern
	keys := []string{
		"test-pattern:user:1",
		"test-pattern:user:2",
		"test-pattern:user:3",
		"test-other:data",
	}

	for _, key := range keys {
		err := client.Set(ctx, key, "value", time.Minute)
		require.NoError(t, err)
	}

	// Delete keys matching pattern
	err := client.DeletePattern(ctx, "test-pattern:user:*")
	require.NoError(t, err)

	// Verify user keys are deleted
	exists1, err := client.Exists(ctx, "test-pattern:user:1")
	require.NoError(t, err)
	assert.False(t, exists1)

	exists2, err := client.Exists(ctx, "test-pattern:user:2")
	require.NoError(t, err)
	assert.False(t, exists2)

	exists3, err := client.Exists(ctx, "test-pattern:user:3")
	require.NoError(t, err)
	assert.False(t, exists3)

	// Verify other key still exists
	exists, err := client.Exists(ctx, "test-other:data")
	require.NoError(t, err)
	assert.True(t, exists)

	// Cleanup
	cleanupRedis(t, client, "test-other:data")
}

func TestRedisClient_Expiration(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-expiration"

	// Set with 1s expiration (Redis minimum)
	err := client.Set(ctx, key, "value", 1*time.Second)
	require.NoError(t, err)

	// Key should exist
	value, err := client.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "value", value)

	// Wait for expiration
	time.Sleep(1200 * time.Millisecond)

	// Key should be gone
	_, err = client.Get(ctx, key)
	assert.Error(t, err)
}

func TestRedisClient_ConcurrentAccess(t *testing.T) {
	client := setupRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-concurrent"

	// Run 100 concurrent increments
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			_, err := client.Increment(ctx, key)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify final count
	value, err := client.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "100", value)

	// Cleanup
	cleanupRedis(t, client, key)
}

// Benchmark tests
func BenchmarkRedisClient_Set(b *testing.B) {
	cfg := getTestRedisConfig()
	client, err := cache.NewRedisClient(cfg)
	if err != nil {
		b.Skip("Redis not available:", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Set(ctx, "bench-key", "value", time.Minute)
	}
}

func BenchmarkRedisClient_Get(b *testing.B) {
	cfg := getTestRedisConfig()
	client, err := cache.NewRedisClient(cfg)
	if err != nil {
		b.Skip("Redis not available:", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Setup
	_ = client.Set(ctx, "bench-key", "value", time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Get(ctx, "bench-key")
	}
}

func BenchmarkRedisClient_SetJSON(b *testing.B) {
	cfg := getTestRedisConfig()
	client, err := cache.NewRedisClient(cfg)
	if err != nil {
		b.Skip("Redis not available:", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	type TestData struct {
		Name  string
		Age   int
		Email string
	}

	data := TestData{Name: "John", Age: 30, Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SetJSON(ctx, "bench-json", data, time.Minute)
	}
}

func BenchmarkRedisClient_Increment(b *testing.B) {
	cfg := getTestRedisConfig()
	client, err := cache.NewRedisClient(cfg)
	if err != nil {
		b.Skip("Redis not available:", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Increment(ctx, "bench-counter")
	}
}
