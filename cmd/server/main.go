// cmd/server/main.go
package main

import (
	"barber-booking-system/config"
	_ "barber-booking-system/docs"
	"barber-booking-system/internal/cache"
	appConfig "barber-booking-system/internal/config"
	"barber-booking-system/internal/logger"
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

	// Swagger imports

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Barbershop Booking API
// @version 1.0.0
// @description A production-ready RESTful API for barbershop booking system.
// @description Features include: User authentication, Barber management, Service catalog, Booking system, Review & Rating system, and Notification system.

// @contact.name API Support
// @contact.email support@barbershop.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

// @tag.name Authentication
// @tag.description User registration, login, and authentication management

// @tag.name Barbers
// @tag.description Barber profile management and discovery

// @tag.name Services
// @tag.description Service catalog and category management

// @tag.name Barber-Services
// @tag.description Manage services offered by barbers

// @tag.name Bookings
// @tag.description Appointment booking and scheduling

// @tag.name Reviews
// @tag.description Customer reviews and ratings for barbers

// @tag.name Notifications
// @tag.description User notification management and delivery

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
	initLogger(cfg)

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

	// Initialize Redis connection (optional)
	var redisClient *cache.RedisClient
	var cacheService *cache.CacheService

	redisClient, err = cache.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Printf("⚠️  Warning: Redis connection failed: %v. Continuing without cache.", err)
		redisClient = nil
	} else {
		defer redisClient.Close()
		log.Println("✅ Redis connection established")
		cacheService = cache.NewCacheService(redisClient)
	}

	// Initialize Gin router
	router := setupRouter(cfg, dbManager)

	// Setup request limits BEFORE other middleware
	setupRequestLimits(router, cfg)

	// Setup all middleware (including Redis rate limiting if available)
	setupMiddlewareWithRedis(router, cfg, redisClient)

	// Setup routes (pass cache service)
	SetupRoutes(router, dbManager.DB, cfg, cacheService)

	// Setup Swagger
	setupSwagger(router)

	// Create server manager
	serverManager := config.NewServerManager(cfg.Server, router)

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on %s", serverManager.GetFullAddress())
		log.Printf("📝 Environment: %s", cfg.App.Environment)
		log.Printf("🔧 Gin mode: %s", cfg.Server.GinMode)
		log.Printf("📚 Swagger UI: http://%s/swagger/index.html", serverManager.GetFullAddress())
		if redisClient != nil {
			log.Printf("🔴 Redis: Enabled (caching & rate limiting)")
		} else {
			log.Printf("⚪ Redis: Disabled (in-memory rate limiting)")
		}
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

// setupSwagger configures Swagger documentation route
func setupSwagger(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Println("📚 Swagger documentation enabled at /swagger/index.html")
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
	router.MaxMultipartMemory = maxSize
	router.Use(middleware.DefaultRequestBodyLimit(maxSize))
	log.Printf("📦 Request body limit: %.2f MB", float64(maxSize)/(1024*1024))
}

// setupMiddlewareWithRedis configures all middleware with optional Redis support
func setupMiddlewareWithRedis(router *gin.Engine, cfg *appConfig.Config, redisClient *cache.RedisClient) {
	middleware.SetupRequestLimits(router, cfg.Upload.MaxFileSize)
	middleware.SetupAll(router, middleware.SetupConfig{
		Config:      cfg,
		RedisClient: redisClient,
	})
}

// printBanner prints application startup banner
func printBanner() {
	banner := `
╔════════════════════════════════════════════════════════╗
║                                                        ║
║           💈 BARBERSHOP BOOKING API 💈                 ║
║                  with Redis Cache                      ║
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
	log.Printf("   Redis: %s", cfg.Redis.URL)
	log.Printf("   JWT Expiration: %v", cfg.JWT.Expiration)
	log.Printf("   Rate Limit: %d req/min", cfg.API.RateLimit)
	log.Printf("   Upload Max Size: %.2f MB", float64(cfg.Upload.MaxFileSize)/(1024*1024))
	log.Printf("   CORS Origins: %v", cfg.CORS.AllowedOrigins)

	if cfg.IsDevelopment() {
		log.Println("⚠️  Running in DEVELOPMENT mode")
	} else if cfg.IsProduction() {
		log.Println("🔒 Running in PRODUCTION mode - security enhanced")
	}
}
func initLogger(cfg *appConfig.Config) {
	logFormat := "json"
	logLevel := "info"

	if cfg.IsDevelopment() {
		logFormat = "console"
		logLevel = "debug"
	}

	// Override from config if set
	if cfg.Logging.Format != "" {
		logFormat = cfg.Logging.Format
	}
	if cfg.Logging.Level != "" {
		logLevel = cfg.Logging.Level
	}

	logger.Init(logger.Config{
		Level:       logLevel,
		Format:      logFormat,
		ServiceName: cfg.App.Name,
		Environment: cfg.App.Environment,
	})

	logger.Info().
		Str("log_level", logLevel).
		Str("log_format", logFormat).
		Msg("Logger initialized")
}
