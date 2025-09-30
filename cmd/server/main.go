// cmd/server/main.go
package main

import (
	"barber-booking-system/config"
	appConfig "barber-booking-system/internal/config"
	"barber-booking-system/internal/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Print startup banner
	printBanner()

	// Load configuration (includes JWT secret validation)
	cfg, err := appConfig.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Log configuration summary (with sensitive data masked)
	logConfigSummary(cfg)

	// Initialize database connection
	dbManager, err := config.NewDatabaseManager(cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer dbManager.Close()

	// Test database connection
	if err := dbManager.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection established")

	// Get database info
	dbInfo, err := dbManager.GetDatabaseInfo()
	if err != nil {
		log.Printf("⚠️  Warning: Could not get database info: %v", err)
	} else {
		log.Printf("📊 Database: %s (%d tables)", dbInfo.DatabaseName, dbInfo.TableCount)
	}

	// Initialize Gin router
	router := setupRouter(cfg, dbManager)

	// Setup request limits BEFORE other middleware
	setupRequestLimits(router, cfg)

	// Setup all middleware
	setupMiddleware(router, cfg)

	// Setup routes
	SetupRoutes(router, dbManager.DB, cfg)

	// Create server manager
	serverManager := config.NewServerManager(cfg.Server, router)

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on %s", serverManager.GetFullAddress())
		log.Printf("📝 Environment: %s", cfg.App.Environment)
		log.Printf("🔧 Gin mode: %s", cfg.Server.GinMode)
		if err := serverManager.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
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
		log.Fatal("❌ Server forced to shutdown:", err)
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

// setupRequestLimits configures request body size limits
func setupRequestLimits(router *gin.Engine, cfg *appConfig.Config) {
	maxSize := cfg.Upload.MaxFileSize

	// Set multipart form memory limit
	router.MaxMultipartMemory = maxSize

	// Add request body limit middleware
	router.Use(middleware.DefaultRequestBodyLimit(maxSize))

	log.Printf("📦 Request body limit: %.2f MB", float64(maxSize)/(1024*1024))
}

// printBanner prints application startup banner
func printBanner() {
	banner := `
╔════════════════════════════════════════════════════════╗
║                                                        ║
║           💈 BARBERSHOP BOOKING API 💈                 ║
║                                                        ║
╚════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

// logConfigSummary logs a summary of the configuration
func logConfigSummary(cfg *appConfig.Config) {
	log.Println("📋 Configuration Summary:")
	log.Printf("   App: %s v%s", cfg.App.Name, cfg.App.Version)
	log.Printf("   Environment: %s", cfg.App.Environment)
	log.Printf("   Server: %s (mode: %s)", cfg.GetServerAddress(), cfg.Server.GinMode)
	log.Printf("   Database: Connected")
	log.Printf("   JWT Expiration: %v", cfg.JWT.Expiration)
	log.Printf("   Rate Limit: %d req/min", cfg.API.RateLimit)
	log.Printf("   Upload Max Size: %.2f MB", float64(cfg.Upload.MaxFileSize)/(1024*1024))
	log.Printf("   CORS Origins: %v", cfg.CORS.AllowedOrigins)

	// Warning for development
	if cfg.IsDevelopment() {
		log.Println("⚠️  Running in DEVELOPMENT mode")
	} else if cfg.IsProduction() {
		log.Println("🔒 Running in PRODUCTION mode - security enhanced")
	}
}
