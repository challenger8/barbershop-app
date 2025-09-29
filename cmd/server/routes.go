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
// SetupRoutes configures all application routes
func SetupRoutes(router *gin.Engine, db *sqlx.DB, cfg *appConfig.Config) { // ✅ Add cfg parameter
	routes.Setup(router, db, cfg.JWT.Secret) // ✅ Pass JWT secret
}

func setupMiddleware(router *gin.Engine, cfg *appConfig.Config) {
	// 1. Recovery - must be first to catch panics
	router.Use(middleware.RecoveryHandler())

	// 2. CORS - handle cross-origin requests
	var corsConfig middleware.CORSConfig
	if cfg.App.Environment == "development" {
		corsConfig = middleware.DevelopmentCORSConfig()
	} else {
		corsConfig = middleware.ProductionCORSConfig(cfg.CORS.AllowedOrigins)
	}
	router.Use(middleware.CORS(corsConfig))

	// 3. Security Headers
	router.Use(middleware.SecurityHeaders())

	// 4. Logging - log all requests
	logConfig := middleware.LoggerConfig{
		Format:          middleware.JSONFormat,
		SkipPaths:       []string{"/health", "/metrics"},
		LogRequestBody:  cfg.App.Environment == "development",
		LogResponseBody: false,
		MaxBodySize:     1024,
	}
	router.Use(middleware.Logger(logConfig))

	// 5. Rate Limiting - prevent abuse
	router.Use(middleware.DefaultRateLimit())

	// 6. Error Handler - must be last
	router.Use(middleware.ErrorHandler())
}
