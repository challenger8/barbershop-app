// internal/routes/routes.go
package routes

import (
	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/handlers"
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Setup configures all application routes with cache support
func Setup(router *gin.Engine, db *sqlx.DB, jwtSecret string, cacheService *cache.CacheService) {
	// Initialize repositories
	barberRepo := repository.NewBarberRepository(db)

	// Initialize services (with cache if available)
	barberService := services.NewBarberService(barberRepo, cacheService)

	// Initialize handlers
	barberHandler := handlers.NewBarberHandler(barberService)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Barber routes
		barbers := v1.Group("/barbers")
		{
			// Public routes (no auth required)
			barbers.GET("", barberHandler.GetAllBarbers)
			barbers.GET("/search", barberHandler.SearchBarbers)
			barbers.GET("/:id", barberHandler.GetBarber)
			barbers.GET("/uuid/:uuid", barberHandler.GetBarberByUUID)
			barbers.GET("/:id/statistics", barberHandler.GetBarberStatistics)
		}

		// Protected barber routes (auth required)
		barbersProtected := v1.Group("/barbers")
		barbersProtected.Use(middleware.RequireAuth(jwtSecret))
		{
			barbersProtected.POST("", barberHandler.CreateBarber)
			barbersProtected.PUT("/:id", barberHandler.UpdateBarber)
			barbersProtected.DELETE("/:id", barberHandler.DeleteBarber)
			barbersProtected.PATCH("/:id/status", barberHandler.UpdateBarberStatus)
		}
	}
}
