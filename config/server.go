// config/server.go
package config

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	appConfig "barber-booking-system/internal/config"
)

// ServerManager manages HTTP server configuration and lifecycle
type ServerManager struct {
	Server *http.Server
	Config appConfig.ServerConfig
}

// NewServerManager creates a new server manager with the given configuration
func NewServerManager(config appConfig.ServerConfig, handler http.Handler) *ServerManager {
	// Set Gin mode based on configuration
	gin.SetMode(config.GinMode)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", config.Port),
		Handler:        handler,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return &ServerManager{
		Server: server,
		Config: config,
	}
}

// Start starts the HTTP server
func (sm *ServerManager) Start() error {
	return sm.Server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (sm *ServerManager) Shutdown(ctx context.Context) error {
	return sm.Server.Shutdown(ctx)
}

// GetAddress returns the server address
func (sm *ServerManager) GetAddress() string {
	return sm.Server.Addr
}

// GetFullAddress returns the full server address with protocol
func (sm *ServerManager) GetFullAddress() string {
	protocol := "http"
	if sm.Server.TLSConfig != nil {
		protocol = "https"
	}

	host := sm.Config.Host
	if host == "" {
		host = "localhost"
	}

	return fmt.Sprintf("%s://%s:%s", protocol, host, sm.Config.Port)
}

// SetupMiddleware sets up all middleware based on configuration
func SetupMiddleware(router *gin.Engine, config *appConfig.Config) {
	// Recovery middleware
	router.Use(gin.Recovery())

	// Logging middleware
	router.Use(CreateLoggingMiddleware(config.Logging))

	// CORS middleware
	router.Use(CreateCORSMiddleware(config.CORS))

	// Security headers
	router.Use(SecurityHeadersMiddleware())
}

// CreateCORSMiddleware creates CORS middleware based on configuration
func CreateCORSMiddleware(config appConfig.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Origin", getAllowedOrigin(c.Request, config.AllowedOrigins))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CreateLoggingMiddleware creates a logging middleware based on configuration
func CreateLoggingMiddleware(config appConfig.LoggingConfig) gin.HandlerFunc {
	// Configure Gin's default logger based on format
	if config.Format == "json" {
		return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			return fmt.Sprintf(`{"time":"%s","method":"%s","path":"%s","status":%d,"latency":"%s","ip":"%s","user_agent":"%s","error":"%s"}%s`,
				param.TimeStamp.Format(time.RFC3339),
				param.Method,
				param.Path,
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Request.UserAgent(),
				param.ErrorMessage,
				"\n",
			)
		})
	}

	// Default text format
	return gin.Logger()
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}

// CreateHealthCheckHandler returns a health check handler
func CreateHealthCheckHandler(dbManager *DatabaseManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := gin.H{
			"status":    "healthy",
			"service":   "barbershop-api",
			"timestamp": time.Now().Format(time.RFC3339),
		}

		// Check database health if database manager is provided
		if dbManager != nil {
			dbHealth := dbManager.Health()
			response["database"] = dbHealth

			// If database is unhealthy, mark overall status as unhealthy
			if dbHealth.Status == "unhealthy" {
				response["status"] = "unhealthy"
				c.JSON(http.StatusServiceUnavailable, response)
				return
			}
		}

		c.JSON(http.StatusOK, response)
	}
}

// getAllowedOrigin determines the appropriate allowed origin based on the request
func getAllowedOrigin(r *http.Request, allowedOrigins []string) string {
	origin := r.Header.Get("Origin")

	// If no specific origins configured or wildcard is present, allow all
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return "*"
		}
		if allowed == origin {
			return origin
		}
	}

	// If origin doesn't match any allowed origins, return the first one or empty
	if len(allowedOrigins) > 0 && allowedOrigins[0] != "*" {
		return allowedOrigins[0]
	}

	return ""
}
