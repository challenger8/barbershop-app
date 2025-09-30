// cmd/server/routes.go
package main

import (
	"barber-booking-system/internal/cache"
	appConfig "barber-booking-system/internal/config"
	"barber-booking-system/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// SetupRoutes configures all application routes with cache support
func SetupRoutes(router *gin.Engine, db *sqlx.DB, cfg *appConfig.Config, cacheService *cache.CacheService) {
	routes.Setup(router, db, cfg.JWT.Secret, cacheService)
}
