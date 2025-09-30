// config/health.go
package config

import (
	"context"
	"net/http"
	"time"

	"barber-booking-system/internal/cache"

	"github.com/gin-gonic/gin"
)

// CreateHealthCheckHandlerWithRedis creates a health check handler with Redis support
func CreateHealthCheckHandlerWithRedis(dbManager *DatabaseManager, redisClient *cache.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		health := gin.H{
			"status":    "healthy",
			"service":   "barbershop-api",
			"timestamp": time.Now().Format(time.RFC3339),
		}

		// Check database
		if err := dbManager.Ping(); err != nil {
			health["status"] = "unhealthy"
			health["database"] = map[string]interface{}{
				"status": "disconnected",
				"error":  err.Error(),
			}
			c.JSON(http.StatusServiceUnavailable, health)
			return
		}
		health["database"] = map[string]interface{}{
			"status": "connected",
		}

		// Check Redis if available
		if redisClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_, err := redisClient.Get(ctx, "health-check")
			if err != nil && err.Error() != "key not found" {
				health["redis"] = map[string]interface{}{
					"status": "disconnected",
					"error":  err.Error(),
				}
			} else {
				health["redis"] = map[string]interface{}{
					"status": "connected",
				}
			}
		} else {
			health["redis"] = map[string]interface{}{
				"status": "disabled",
			}
		}

		c.JSON(http.StatusOK, health)
	}
}
