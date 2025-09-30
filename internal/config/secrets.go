// internal/config/secrets.go
package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// GenerateJWTSecret generates a cryptographically secure JWT secret
// Returns a base64-encoded string of 32 random bytes (256 bits)
func GenerateJWTSecret() (string, error) {
	bytes := make([]byte, 32) // 256 bits for HS256
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateJWTSecret validates JWT secret meets security requirements
func ValidateJWTSecret(secret string, environment string) error {
	if secret == "" {
		return fmt.Errorf("JWT secret cannot be empty")
	}

	// Minimum length check
	minLength := 32
	if environment == "production" {
		minLength = 43 // base64 encoded 32 bytes = 43 characters
	}

	if len(secret) < minLength {
		return fmt.Errorf("JWT secret must be at least %d characters (got %d)", minLength, len(secret))
	}

	// Check for common insecure secrets
	insecureSecrets := []string{
		"barbershop-super-secret-jwt-key-change-in-production",
		"your-secret-key",
		"secret",
		"change-me",
		"test-secret",
		"development-secret",
		"my-secret-key",
		"jwt-secret",
		"super-secret",
	}

	secretLower := strings.ToLower(secret)
	for _, insecure := range insecureSecrets {
		if secretLower == strings.ToLower(insecure) || strings.Contains(secretLower, insecure) {
			return fmt.Errorf("JWT secret must not contain common/default values")
		}
	}

	// In production, check for entropy
	if environment == "production" {
		if !hasGoodEntropy(secret) {
			return fmt.Errorf("JWT secret appears to have low entropy (not random enough)")
		}
	}

	return nil
}

// hasGoodEntropy checks if string has reasonable entropy
// Simple check: should have variety of characters
func hasGoodEntropy(s string) bool {
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, c := range s {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	// Should have at least 3 types of characters or be base64 encoded
	typeCount := 0
	if hasUpper {
		typeCount++
	}
	if hasLower {
		typeCount++
	}
	if hasDigit {
		typeCount++
	}
	if hasSpecial {
		typeCount++
	}

	// Base64 has upper, lower, digits, and +/= so it's good
	return typeCount >= 3
}

// EnsureJWTSecret gets JWT secret from env with environment-specific validation
func EnsureJWTSecret(environment string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	// Production: Strict requirements
	if environment == "production" {
		if secret == "" {
			return "", fmt.Errorf("JWT_SECRET is required in production environment")
		}

		if err := ValidateJWTSecret(secret, environment); err != nil {
			return "", fmt.Errorf("JWT secret validation failed: %w", err)
		}

		return secret, nil
	}

	// Staging: Same as production but with warnings
	if environment == "staging" {
		if secret == "" {
			return "", fmt.Errorf("JWT_SECRET is required in staging environment")
		}

		if err := ValidateJWTSecret(secret, "production"); err != nil {
			fmt.Printf("⚠️  WARNING: %v\n", err)
		}

		return secret, nil
	}

	// Development/Test: Generate if missing
	if secret == "" {
		generated, err := GenerateJWTSecret()
		if err != nil {
			return "", fmt.Errorf("failed to generate JWT secret: %w", err)
		}

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("⚠️  WARNING: JWT_SECRET not set in environment")
		fmt.Println("⚠️  Generated temporary secret for development")
		fmt.Println("⚠️  This secret will change on each restart!")
		fmt.Println("⚠️")
		fmt.Println("⚠️  Add this to your .env file:")
		fmt.Printf("⚠️  JWT_SECRET=%s\n", generated)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		return generated, nil
	}

	// Validate even in development (but just warn)
	if err := ValidateJWTSecret(secret, environment); err != nil {
		fmt.Printf("⚠️  WARNING: %v\n", err)
		fmt.Println("⚠️  Consider generating a new secure secret with: make generate-jwt-secret")
	}

	return secret, nil
}

// ValidateDatabaseURL validates database URL based on environment
func ValidateDatabaseURL(url string, environment string) error {
	if url == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	// Production checks
	if environment == "production" {
		// Must use SSL
		if !strings.Contains(url, "sslmode=require") && !strings.Contains(url, "sslmode=verify-") {
			return fmt.Errorf("database must use SSL in production (add sslmode=require)")
		}

		// Should not use localhost
		if strings.Contains(url, "localhost") || strings.Contains(url, "127.0.0.1") {
			return fmt.Errorf("database URL should not use localhost in production")
		}

		// Should not have default passwords
		insecurePasswords := []string{"password", "admin", "root", "123456"}
		urlLower := strings.ToLower(url)
		for _, pwd := range insecurePasswords {
			if strings.Contains(urlLower, ":"+pwd+"@") {
				return fmt.Errorf("database URL contains insecure password")
			}
		}
	}

	return nil
}

// ValidateProductionConfig performs comprehensive production validation
// This function accepts a pointer to Config struct
func ValidateProductionConfig(cfg *Config) error {
	var errors []string

	// JWT Secret
	if err := ValidateJWTSecret(cfg.JWT.Secret, cfg.App.Environment); err != nil {
		errors = append(errors, fmt.Sprintf("JWT: %v", err))
	}

	// Database URL
	if err := ValidateDatabaseURL(cfg.Database.URL, cfg.App.Environment); err != nil {
		errors = append(errors, fmt.Sprintf("Database: %v", err))
	}

	// Server Configuration
	if cfg.Server.GinMode != "release" {
		errors = append(errors, "Server: GIN_MODE must be 'release' in production")
	}

	// CORS Configuration
	if len(cfg.CORS.AllowedOrigins) > 0 {
		for _, origin := range cfg.CORS.AllowedOrigins {
			if origin == "*" {
				errors = append(errors, "CORS: wildcard origin (*) is not allowed in production")
				break
			}
		}
	}

	// SMTP Configuration (if used)
	if cfg.SMTP.Host != "" && cfg.SMTP.Password == "" {
		errors = append(errors, "SMTP: password is required when SMTP is configured")
	}

	// Redis Password (recommended in production)
	if cfg.App.Environment == "production" {
		if strings.Contains(cfg.Redis.URL, "localhost") {
			errors = append(errors, "Redis: should not use localhost in production")
		}
	}

	// Rate Limiting (should be reasonable)
	if cfg.API.RateLimit > 10000 {
		errors = append(errors, fmt.Sprintf("API: rate limit is too high (%d req/min)", cfg.API.RateLimit))
	}

	// Upload directory should be absolute path in production
	if cfg.App.Environment == "production" {
		if !strings.HasPrefix(cfg.Upload.Directory, "/") {
			errors = append(errors, "Upload: directory should be an absolute path in production")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("production validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// ValidateStagingConfig performs staging-specific validation
func ValidateStagingConfig(cfg *Config) error {
	// Staging uses same strict rules as production
	return ValidateProductionConfig(cfg)
}

// ValidateDevelopmentConfig performs development-specific validation
func ValidateDevelopmentConfig(cfg *Config) error {
	var warnings []string

	// Just warnings in development, not errors
	if cfg.JWT.Secret != "" {
		if err := ValidateJWTSecret(cfg.JWT.Secret, "development"); err != nil {
			warnings = append(warnings, fmt.Sprintf("JWT: %v", err))
		}
	}

	if cfg.Server.GinMode == "release" {
		warnings = append(warnings, "Server: Using release mode in development (consider using 'debug')")
	}

	if len(warnings) > 0 {
		fmt.Println("⚠️  Development configuration warnings:")
		for _, warning := range warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	return nil
}
