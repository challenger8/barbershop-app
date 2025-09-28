// cmd/server/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"barber-booking-system/config"
	appConfig "barber-booking-system/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration using the new configuration system
	cfg, err := appConfig.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("🚀 Starting %s v%s in %s mode", cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// Initialize database connection using configuration
	dbManager, err := config.NewDatabaseManager(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbManager.Close()

	// Test database connection
	if err := dbManager.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection established")

	// Initialize Gin router with configuration
	router := setupRouter(cfg, dbManager)

	// Create server manager using configuration
	serverManager := config.NewServerManager(cfg.Server, router)

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on %s", serverManager.GetFullAddress())
		if err := serverManager.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := serverManager.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited gracefully")
}

// setupRouter configures all routes using configuration
func setupRouter(cfg *appConfig.Config, dbManager *config.DatabaseManager) *gin.Engine {
	// Set Gin mode from configuration
	gin.SetMode(cfg.Server.GinMode)

	router := gin.Default()

	// Setup middleware using configuration
	config.SetupMiddleware(router, cfg)

	// Health check endpoint using the new health check handler
	router.GET("/health", config.CreateHealthCheckHandler(dbManager))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Barber routes (using your existing placeholder handlers for now)
		barbers := v1.Group("/barbers")
		{
			barbers.GET("", placeholderHandler("List Barbers"))
			barbers.GET("/:id", placeholderHandler("Get Barber"))
			barbers.POST("", placeholderHandler("Create Barber"))
			barbers.PUT("/:id", placeholderHandler("Update Barber"))
			barbers.DELETE("/:id", placeholderHandler("Delete Barber"))
		}

		// Add more API groups here as you build them
		// users := v1.Group("/users")
		// bookings := v1.Group("/bookings")
		// services := v1.Group("/services")
	}

	return router
}

// placeholderHandler returns a placeholder response (keeping your existing implementation)
func placeholderHandler(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"action":  action,
			"status":  "not_implemented",
			"message": "This endpoint will be implemented in the next step",
		})
	}
}
