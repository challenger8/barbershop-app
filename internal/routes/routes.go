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
	serviceRepo := repository.NewServiceRepository(db)

	// Initialize services (with cache support)
	userService := services.NewUserService(userRepo, jwtSecret, jwtExpiration)
	barberService := services.NewBarberService(barberRepo, cacheService)
	serviceService := services.NewServiceService(serviceRepo, cacheService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	barberHandler := handlers.NewBarberHandler(barberService)
	serviceHandler := handlers.NewServiceHandler(serviceService)

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

		// Service routes
		services := v1.Group("/services")
		{
			// Public service routes (no auth required)
			services.GET("", serviceHandler.GetAllServices)
			services.GET("/search", serviceHandler.SearchServices)
			services.GET("/categories", serviceHandler.GetAllCategories)
			services.GET("/categories/:id", serviceHandler.GetCategory)
			services.GET("/:id", serviceHandler.GetService)
			services.GET("/slug/:slug", serviceHandler.GetServiceBySlug)
			services.GET("/:service_id/barbers", serviceHandler.GetBarbersOfferingService)
		}

		// Protected service routes (auth required - admin only)
		servicesProtected := v1.Group("/services")
		servicesProtected.Use(middleware.RequireAuth(jwtSecret))
		{
			servicesProtected.POST("", serviceHandler.CreateService)
			servicesProtected.PUT("/:id", serviceHandler.UpdateService)
			servicesProtected.DELETE("/:id", serviceHandler.DeleteService)
			servicesProtected.POST("/categories", serviceHandler.CreateCategory)
			servicesProtected.PUT("/categories/:id", serviceHandler.UpdateCategory)
			servicesProtected.DELETE("/categories/:id", serviceHandler.DeleteCategory)
		}

		// Barber services routes (services offered by barbers)
		barberServices := v1.Group("/barber-services")
		{
			// Public routes
			barberServices.GET("/:id", serviceHandler.GetBarberServiceByID)
		}

		// Protected barber services routes
		barberServicesProtected := v1.Group("/barber-services")
		barberServicesProtected.Use(middleware.RequireAuth(jwtSecret))
		{
			barberServicesProtected.POST("", serviceHandler.AddServiceToBarber)
			barberServicesProtected.PUT("/:id", serviceHandler.UpdateBarberService)
			barberServicesProtected.DELETE("/:id", serviceHandler.RemoveServiceFromBarber)
		}

		// Get barber's services (nested under barbers)
		v1.GET("/barbers/:barber_id/services", serviceHandler.GetBarberServices)
	}
}
