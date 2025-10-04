// internal/routes/routes.go
package routes

import (
	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/handlers"
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Setup configures all application routes
func Setup(router *gin.Engine, db *sqlx.DB, jwtSecret string, jwtExpiration time.Duration, cacheService *cache.CacheService) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	barberRepo := repository.NewBarberRepository(db)

	// Initialize services (with cache support)
	userService := services.NewUserService(userRepo, jwtSecret, jwtExpiration)
	barberService := services.NewBarberService(barberRepo, cacheService) // Pass cache here

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	barberHandler := handlers.NewBarberHandler(barberService)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			// Public auth routes (no authentication required)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)

			// Protected auth routes (authentication required)
			authProtected := auth.Group("")
			authProtected.Use(middleware.RequireAuth(jwtSecret))
			{
				authProtected.GET("/me", authHandler.GetMe)
				authProtected.PUT("/profile", authHandler.UpdateProfile)
				authProtected.POST("/change-password", authHandler.ChangePassword)
				authProtected.POST("/logout", authHandler.Logout)
			}
		}

		// Barber routes
		barbers := v1.Group("/barbers")
		{
			// Public barber routes (no auth required)
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
