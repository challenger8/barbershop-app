// internal/middleware/setup.go
package middleware

import (
	"barber-booking-system/internal/cache"
	appConfig "barber-booking-system/internal/config"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// ========================================================================
// SIMPLIFIED MIDDLEWARE SETUP - Following KISS Principle
// ========================================================================

// SetupConfig holds all configuration needed for middleware setup
type SetupConfig struct {
	Config      *appConfig.Config
	RedisClient *cache.RedisClient
}

// SetupAll configures all middleware in the correct order
// This centralizes middleware setup logic following KISS principle
func SetupAll(router *gin.Engine, cfg SetupConfig) {
	// Log setup start
	log.Println("🔧 Configuring middleware stack...")

	// 1. Recovery - must be first to catch panics
	router.Use(RecoveryHandler())
	log.Println("   ✓ Recovery handler")

	// 2. Request ID - for tracing
	router.Use(RequestIDMiddleware())
	log.Println("   ✓ Request ID tracking")

	// 3. CORS - handle cross-origin requests
	setupCORS(router, cfg.Config)
	log.Println("   ✓ CORS configured")

	// 4. Security Headers
	router.Use(SecurityHeaders())
	log.Println("   ✓ Security headers")

	// 5. Logging
	setupLogging(router, cfg.Config)
	log.Println("   ✓ Request logging")

	// 6. Rate Limiting
	setupRateLimiting(router, cfg.Config, cfg.RedisClient)

	// 7. Error Handler - must be last
	router.Use(ErrorHandler())
	log.Println("   ✓ Error handler")

	log.Println("✅ Middleware stack ready")
}

// setupCORS configures CORS based on environment
func setupCORS(router *gin.Engine, cfg *appConfig.Config) {
	var corsConfig CORSConfig
	if cfg.IsDevelopment() {
		corsConfig = DevelopmentCORSConfig()
	} else {
		corsConfig = ProductionCORSConfig(cfg.CORS.AllowedOrigins)
	}
	router.Use(CORS(corsConfig))
}

// setupLogging configures logging middleware
func setupLogging(router *gin.Engine, cfg *appConfig.Config) {
	logConfig := LoggerConfig{
		Format:          getLogFormat(cfg),
		SkipPaths:       []string{"/health", "/metrics"},
		LogRequestBody:  cfg.IsDevelopment(),
		LogResponseBody: false,
		MaxBodySize:     1024,
	}
	router.Use(Logger(logConfig))
}

// setupRateLimiting configures rate limiting with Redis or in-memory fallback
func setupRateLimiting(router *gin.Engine, cfg *appConfig.Config, redisClient *cache.RedisClient) {
	if redisClient != nil {
		// Redis-based distributed rate limiting
		log.Println("   ✓ Rate limiting (Redis-based)")
		rateLimiter := NewRateLimiter(
			redisClient,
			cfg.API.RateLimit,
			time.Minute,
		)
		router.Use(rateLimiter.Middleware())
	} else {
		// Fallback to in-memory rate limiting
		log.Println("   ✓ Rate limiting (in-memory)")
		rateLimitConfig := DefaultRateLimitConfig()
		rateLimitConfig.Limit = cfg.API.RateLimit
		router.Use(RateLimitMiddleware(rateLimitConfig))
	}
}

// getLogFormat returns the appropriate log format based on config
func getLogFormat(cfg *appConfig.Config) LogFormat {
	if cfg.Logging.Format == "json" {
		return JSONFormat
	}
	return TextFormat
}

// SetupRequestLimits configures request body size limits
func SetupRequestLimits(router *gin.Engine, maxSize int64) {
	router.MaxMultipartMemory = maxSize
	router.Use(DefaultRequestBodyLimit(maxSize))
	log.Printf("📦 Request body limit: %.2f MB", float64(maxSize)/(1024*1024))
}
