// cmd/debug/main.go
package main

import (
	"flag"
	"fmt"
	"log"

	"barber-booking-system/config"
	appConfig "barber-booking-system/internal/config"
)

func main() {
	// Define flags
	checkDB := flag.Bool("db", false, "Check database connection")
	showConfig := flag.Bool("config", false, "Show configuration")
	testJWT := flag.Bool("jwt", false, "Test JWT token generation")

	flag.Parse()

	cfg, err := appConfig.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Check database
	if *checkDB {
		checkDatabase(cfg)
		return
	}

	// Show config
	if *showConfig {
		showConfiguration(cfg)
		return
	}

	// Test JWT
	if *testJWT {
		testJWTToken(cfg)
		return
	}

	// No flags provided
	fmt.Println("Usage:")
	flag.PrintDefaults()
}

func checkDatabase(cfg *appConfig.Config) {
	fmt.Println("🔍 Checking database connection...")

	dbManager, err := config.NewDatabaseManager(cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed: %v", err)
	}
	defer dbManager.Close()

	if err := dbManager.Ping(); err != nil {
		log.Fatalf("❌ Ping failed: %v", err)
	}

	info, _ := dbManager.GetDatabaseInfo()
	fmt.Printf("✅ Connected to: %s\n", info.DatabaseName)
	fmt.Printf("   Tables: %d\n", info.TableCount)
}

func showConfiguration(cfg *appConfig.Config) {
	fmt.Println("📋 Configuration:")
	fmt.Printf("App: %s v%s\n", cfg.App.Name, cfg.App.Version)
	fmt.Printf("Environment: %s\n", cfg.App.Environment)
	fmt.Printf("Port: %s\n", cfg.Server.Port)
	fmt.Printf("Database: %s\n", maskPassword(cfg.Database.URL))
}

func testJWTToken(cfg *appConfig.Config) {
	fmt.Println("🔐 Testing JWT configuration...")

	if cfg.JWT.Secret == "" {
		fmt.Println("❌ JWT_SECRET not set in .env")
		return
	}

	if len(cfg.JWT.Secret) < 32 {
		fmt.Println("⚠️  JWT_SECRET should be at least 32 characters")
	}

	fmt.Println("✅ JWT secret is configured")
	fmt.Printf("   Length: %d characters\n", len(cfg.JWT.Secret))
}

func maskPassword(url string) string {
	// Mask password in database URL
	return "postgres://user:****@localhost:5432/dbname"
}
