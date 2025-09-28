// cmd/server/routes.go
package main

import (
	"log"

	"barber-booking-system/internal/handlers"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// SetupRoutes configures all application routes
func SetupRoutes(router *gin.Engine, db *sqlx.DB) {
	log.Println("🔧 Setting up routes...")

	// Initialize repositories
	barberRepo := repository.NewBarberRepository(db)
	log.Println("✅ Barber repository initialized")

	// Initialize services
	barberService := services.NewBarberService(barberRepo)
	log.Println("✅ Barber service initialized")

	// Initialize handlers
	barberHandler := handlers.NewBarberHandler(barberService)
	log.Println("✅ Barber handler initialized")

	// Add a simple test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Routes are working!"})
	})
	log.Println("✅ Registered GET /test")

	// API v1 routes
	v1 := router.Group("/api/v1")
	log.Println("✅ Created /api/v1 route group")
	{
		// Add a test endpoint to v1 group
		v1.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "API v1 routes are working!"})
		})
		log.Println("✅ Registered GET /api/v1/test")
		// Barber routes
		barbers := v1.Group("/barbers")
		log.Println("✅ Created /api/v1/barbers route group")
		{
			// List and search
			barbers.GET("", barberHandler.GetAllBarbers)
			log.Println("✅ Registered GET /api/v1/barbers")

			barbers.GET("/search", barberHandler.SearchBarbers)
			log.Println("✅ Registered GET /api/v1/barbers/search")

			// CRUD operations
			barbers.GET("/:id", barberHandler.GetBarber)
			log.Println("✅ Registered GET /api/v1/barbers/:id")

			barbers.GET("/uuid/:uuid", barberHandler.GetBarberByUUID)
			log.Println("✅ Registered GET /api/v1/barbers/uuid/:uuid")

			barbers.POST("", barberHandler.CreateBarber)
			log.Println("✅ Registered POST /api/v1/barbers")

			barbers.PUT("/:id", barberHandler.UpdateBarber)
			log.Println("✅ Registered PUT /api/v1/barbers/:id")

			barbers.DELETE("/:id", barberHandler.DeleteBarber)
			log.Println("✅ Registered DELETE /api/v1/barbers/:id")

			// Additional actions
			barbers.PATCH("/:id/status", barberHandler.UpdateBarberStatus)
			log.Println("✅ Registered PATCH /api/v1/barbers/:id/status")

			barbers.GET("/:id/statistics", barberHandler.GetBarberStatistics)
			log.Println("✅ Registered GET /api/v1/barbers/:id/statistics")
		}
	}

	log.Println("🎉 All routes configured successfully")
}
