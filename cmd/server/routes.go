// cmd/server/routes.go
package main

import (
	appConfig "barber-booking-system/internal/config"
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// SetupRoutes configures all application routes
func SetupRoutes(router *gin.Engine, db *sqlx.DB, cfg *appConfig.Config) {
	routes.Setup(router, db, cfg.JWT.Secret)
}

// setupMiddleware configures all middleware in correct order
func setupMiddleware(router *gin.Engine, cfg *appConfig.Config) {
	// 1. Recovery - must be first to catch panics
	router.Use(middleware.RecoveryHandler())

	// 2. Request ID - for tracing
	router.Use(middleware.RequestIDMiddleware())

	// 3. CORS - handle cross-origin requests
	var corsConfig middleware.CORSConfig
	if cfg.IsDevelopment() {
		corsConfig = middleware.DevelopmentCORSConfig()
	} else {
		corsConfig = middleware.ProductionCORSConfig(cfg.CORS.AllowedOrigins)
	}
	router.Use(middleware.CORS(corsConfig))

	// 4. Security Headers
	router.Use(middleware.SecurityHeaders())

	// 5. Logging - log all requests
	logConfig := middleware.LoggerConfig{
		Format:          getLogFormat(cfg),
		SkipPaths:       []string{"/health", "/metrics"},
		LogRequestBody:  cfg.IsDevelopment(),
		LogResponseBody: false,
		MaxBodySize:     1024,
	}
	router.Use(middleware.Logger(logConfig))

	// 6. Rate Limiting - prevent abuse
	if cfg.IsProduction() || cfg.IsStaging() {
		// Stricter rate limiting in production
		router.Use(middleware.RateLimit(middleware.StrictRateLimit()))
	} else {
		router.Use(middleware.DefaultRateLimit())
	}

	// 7. Error Handler - must be last
	router.Use(middleware.ErrorHandler())
}

// getLogFormat returns the appropriate log format based on config
func getLogFormat(cfg *appConfig.Config) middleware.LogFormat {
	if cfg.Logging.Format == "json" {
		return middleware.JSONFormat
	}
	return middleware.TextFormat
}
