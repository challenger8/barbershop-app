// cmd/server/main.go
package main

import (
	"barber-booking-system/config"
	appConfig "barber-booking-system/internal/config"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := appConfig.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("🚀 Starting %s v%s in %s mode", cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// Initialize database connection
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

	// Initialize Gin router
	router := setupRouter(cfg, dbManager)
	setupMiddleware(router, cfg)

	// Setup routes
	SetupRoutes(router, dbManager.DB, cfg)
	// Create server manager
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

// setupRouter configures basic middleware and health check
func setupRouter(cfg *appConfig.Config, dbManager *config.DatabaseManager) *gin.Engine {
	gin.SetMode(cfg.Server.GinMode)
	router := gin.New()
	router.GET("/health", config.CreateHealthCheckHandler(dbManager))
	return router
}
