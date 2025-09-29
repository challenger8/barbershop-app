// cmd/debug/main.go
package main

import (
	"log"

	"barber-booking-system/config"
	appConfig "barber-booking-system/internal/config"
	"barber-booking-system/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := appConfig.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	dbManager, err := config.NewDatabaseManager(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbManager.Close()

	router := gin.Default()
	routes.Setup(router, dbManager.DB)

	// Print all routes for debugging
	log.Println("📋 Registered routes:")
	allRoutes := router.Routes()
	for _, route := range allRoutes {
		log.Printf("  %s %s", route.Method, route.Path)
	}
	log.Printf("Total routes registered: %d\n", len(allRoutes))
}
