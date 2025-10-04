// cmd/server/routes.go
package main

import (
	"barber-booking-system/internal/cache"
	appConfig "barber-booking-system/internal/config"
	"barber-booking-system/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// SetupRoutes configures all application routes
func SetupRoutes(router *gin.Engine, db *sqlx.DB, cfg *appConfig.Config, cacheService *cache.CacheService) {
	// Pass cache service to routes setup
	routes.Setup(router, db, cfg.JWT.Secret, cfg.JWT.Expiration, cacheService)
}

// NOTE: Keep all other existing functions (setupMiddlewareWithRedis, getLogFormat, etc.) unchanged
// They are already correct in your file
