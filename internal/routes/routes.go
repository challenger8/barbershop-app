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
	// ========================================================================
	// INITIALIZE REPOSITORIES
	// ========================================================================
	userRepo := repository.NewUserRepository(db)
	barberRepo := repository.NewBarberRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	bookingRepo := repository.NewBookingRepository(db) // NEW

	// ========================================================================
	// INITIALIZE SERVICES
	// ========================================================================
	userService := services.NewUserService(userRepo, jwtSecret, jwtExpiration)
	barberService := services.NewBarberService(barberRepo, cacheService)
	serviceService := services.NewServiceService(serviceRepo, cacheService)
	bookingService := services.NewBookingService(bookingRepo, barberRepo, serviceRepo, cacheService) // NEW

	// ========================================================================
	// INITIALIZE HANDLERS
	// ========================================================================
	authHandler := handlers.NewAuthHandler(userService)
	barberHandler := handlers.NewBarberHandler(barberService)
	serviceHandler := handlers.NewServiceHandler(serviceService)
	bookingHandler := handlers.NewBookingHandler(bookingService) // NEW

	// ========================================================================
	// API v1 ROUTES
	// ========================================================================
	v1 := router.Group("/api/v1")
	{
		// ────────────────────────────────────────────────────────────────
		// AUTHENTICATION ROUTES
		// ────────────────────────────────────────────────────────────────
		auth := v1.Group("/auth")
		{
			// Public auth routes
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)

			// Protected auth routes
			protected := auth.Group("")
			protected.Use(middleware.RequireAuth(jwtSecret))
			{
				protected.GET("/me", authHandler.GetMe)
				protected.PUT("/profile", authHandler.UpdateProfile)
				protected.POST("/change-password", authHandler.ChangePassword)
				protected.POST("/logout", authHandler.Logout)
			}
		}

		// ────────────────────────────────────────────────────────────────
		// BARBER ROUTES
		// ────────────────────────────────────────────────────────────────
		barbers := v1.Group("/barbers")
		{
			// Public barber routes
			barbers.GET("", barberHandler.GetAllBarbers)
			barbers.GET("/search", barberHandler.SearchBarbers)
			barbers.GET("/:id", barberHandler.GetBarber)
			barbers.GET("/uuid/:uuid", barberHandler.GetBarberByUUID)
			barbers.GET("/:id/statistics", barberHandler.GetBarberStatistics)
			barbers.GET("/:id/services", serviceHandler.GetBarberServices)

			// NEW: Barber booking routes (public - view bookings)
			barbers.GET("/:id/bookings", bookingHandler.GetBarberBookings)
			barbers.GET("/:id/bookings/today", bookingHandler.GetTodayBookings)
			barbers.GET("/:id/bookings/stats", bookingHandler.GetBarberBookingStats)

			// Protected barber routes
			protected := barbers.Group("")
			protected.Use(middleware.RequireAuth(jwtSecret))
			{
				protected.POST("", barberHandler.CreateBarber)
				protected.PUT("/:id", barberHandler.UpdateBarber)
				protected.DELETE("/:id", barberHandler.DeleteBarber)
				protected.PATCH("/:id/status", barberHandler.UpdateBarberStatus)
			}
		}

		// ────────────────────────────────────────────────────────────────
		// SERVICE ROUTES
		// ────────────────────────────────────────────────────────────────
		svcs := v1.Group("/services")
		{
			// Public service routes
			svcs.GET("", serviceHandler.GetAllServices)
			svcs.GET("/search", serviceHandler.SearchServices)
			svcs.GET("/:id", serviceHandler.GetService)
			svcs.GET("/slug/:slug", serviceHandler.GetServiceBySlug)
			svcs.GET("/categories", serviceHandler.GetAllCategories)
			svcs.GET("/categories/:id", serviceHandler.GetCategory)

			// Protected service routes (admin only)
			protected := svcs.Group("")
			protected.Use(middleware.RequireAuth(jwtSecret))
			{
				protected.POST("", serviceHandler.CreateService)
				protected.PUT("/:id", serviceHandler.UpdateService)
				protected.DELETE("/:id", serviceHandler.DeleteService)

				// Category management
				protected.POST("/categories", serviceHandler.CreateCategory)
				protected.PUT("/categories/:id", serviceHandler.UpdateCategory)
				protected.DELETE("/categories/:id", serviceHandler.DeleteCategory)
			}
		}

		// ────────────────────────────────────────────────────────────────
		// BARBER-SERVICE ROUTES (Junction table management)
		// ────────────────────────────────────────────────────────────────
		barberServices := v1.Group("/barber-services")
		{
			// Protected (all require auth)
			barberServices.Use(middleware.RequireAuth(jwtSecret))
			{
				barberServices.POST("", serviceHandler.AddServiceToBarber)
				barberServices.PUT("/:id", serviceHandler.UpdateBarberService)
				barberServices.DELETE("/:id", serviceHandler.RemoveServiceFromBarber)
			}
		}

		// ────────────────────────────────────────────────────────────────
		// BOOKING ROUTES (NEW!)
		// ────────────────────────────────────────────────────────────────
		bookings := v1.Group("/bookings")
		{
			// Public booking routes
			bookings.GET("/availability", bookingHandler.CheckAvailability)
			bookings.GET("/uuid/:uuid", bookingHandler.GetBookingByUUID)
			bookings.GET("/number/:number", bookingHandler.GetBookingByNumber)

			// Protected booking routes
			protected := bookings.Group("")
			protected.Use(middleware.RequireAuth(jwtSecret))
			{
				// Create booking
				protected.POST("", bookingHandler.CreateBooking)

				// Get bookings
				protected.GET("/me", bookingHandler.GetMyBookings)
				protected.GET("/:id", bookingHandler.GetBooking)
				protected.GET("/:id/history", bookingHandler.GetBookingHistory)

				// Update booking
				protected.PUT("/:id", bookingHandler.UpdateBooking)
				protected.PATCH("/:id/status", bookingHandler.UpdateBookingStatus)
				protected.PUT("/:id/reschedule", bookingHandler.RescheduleBooking)

				// Cancel booking
				protected.DELETE("/:id", bookingHandler.CancelBooking)
			}
		}
	}
}
