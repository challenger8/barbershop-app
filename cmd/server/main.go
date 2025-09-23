// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database connection
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection established")

	// Initialize Gin router
	router := setupRouter(db)

	// Server configuration
	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited gracefully")
}

// initDatabase initializes database connection
func initDatabase() (*sqlx.DB, error) {
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/barbershop?sslmode=disable")

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// setupRouter configures all routes
func setupRouter(db *sqlx.DB) *gin.Engine {
	// Set Gin mode
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Global middleware
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware())

	// Health check endpoint
	router.GET("/health", healthCheckHandler(db))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Barber routes (we'll implement these step by step)
		barbers := v1.Group("/barbers")
		{
			barbers.GET("", placeholderHandler("List Barbers"))
			barbers.GET("/:id", placeholderHandler("Get Barber"))
			barbers.POST("", placeholderHandler("Create Barber"))
			barbers.PUT("/:id", placeholderHandler("Update Barber"))
			barbers.DELETE("/:id", placeholderHandler("Delete Barber"))
		}
	}

	return router
}

// healthCheckHandler returns health status
func healthCheckHandler(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database connection
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "error",
				"message":   "Database connection failed",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "barbershop-api",
			"version":   "1.0.0",
			"database":  "connected",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// placeholderHandler returns a placeholder response
func placeholderHandler(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"action":  action,
			"status":  "not_implemented",
			"message": "This endpoint will be implemented in the next step",
		})
	}
}

// corsMiddleware handles CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// loggingMiddleware logs requests
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[%s] %s %s - Status: %d - Duration: %v",
			method, path, c.ClientIP(), status, duration)
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
