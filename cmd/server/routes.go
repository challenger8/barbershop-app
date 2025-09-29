// cmd/server/routes.go
package main

import (
	"barber-booking-system/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// SetupRoutes configures all application routes
func SetupRoutes(router *gin.Engine, db *sqlx.DB) {
	routes.Setup(router, db)
}
