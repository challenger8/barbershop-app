// cmd/server/routes.go
package main

import (
	"barber-booking-system/internal/handlers"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// SetupRoutes configures all application routes
func SetupRoutes(router *gin.Engine, db *sqlx.DB) {
	// Initialize repositories
	barberRepo := repository.NewBarberRepository(db)

	// Initialize services
	barberService := services.NewBarberService(barberRepo)

	// Initialize handlers
	barberHandler := handlers.NewBarberHandler(barberService)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Barber routes
		barbers := v1.Group("/barbers")
		{
			// List and search
			barbers.GET("", barberHandler.GetAllBarbers)
			barbers.GET("/search", barberHandler.SearchBarbers)

			// CRUD operations
			barbers.GET("/:id", barberHandler.GetBarber)
			barbers.GET("/uuid/:uuid", barberHandler.GetBarberByUUID)
			barbers.POST("", barberHandler.CreateBarber)
			barbers.PUT("/:id", barberHandler.UpdateBarber)
			barbers.DELETE("/:id", barberHandler.DeleteBarber)

			// Additional actions
			barbers.PATCH("/:id/status", barberHandler.UpdateBarberStatus)
			barbers.GET("/:id/statistics", barberHandler.GetBarberStatistics)
		}
	}
}
