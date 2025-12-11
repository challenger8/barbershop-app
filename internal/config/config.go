// internal/config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the complete application configuration
type Config struct {
	App      AppConfig      `json:"app"`
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
	Redis    RedisConfig    `json:"redis"`
	Upload   UploadConfig   `json:"upload"`
	SMTP     SMTPConfig     `json:"smtp"`
	API      APIConfig      `json:"api"`
	CORS     CORSConfig     `json:"cors"`
	Logging LoggingConfig `yaml:"logging"`
}

// AppConfig represents application-level configuration
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	GinMode      string        `json:"gin_mode"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	URL             string        `json:"url"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret     string        `json:"-"` // Don't include secret in JSON output
	Expiration time.Duration `json:"expiration"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	URL      string `json:"url"`
	Password string `json:"-"` // Don't include password in JSON output
	DB       int    `json:"db"`
}

// UploadConfig represents file upload configuration
type UploadConfig struct {
	Directory   string `json:"directory"`
	MaxFileSize int64  `json:"max_file_size"`
}

// SMTPConfig represents email configuration
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"-"` // Don't include password in JSON output
	From     string `json:"from"`
}

// APIConfig represents API configuration
type APIConfig struct {
	RateLimit int           `json:"rate_limit"`
	Timeout   time.Duration `json:"timeout"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL" default:"info"`
	Format string `yaml:"format" env:"LOG_FORMAT" default:"json"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// Load loads configuration from environment variables and .env files
func Load() (*Config, error) {
	// Load environment-specific .env file first
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Try to load environment-specific .env file
	envFile := fmt.Sprintf(".env.%s", env)
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			log.Printf("Warning: Could not load %s: %v", envFile, err)
		}
	}

	// Load default .env file as fallback
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	config := &Config{
		App:      loadAppConfig(),
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
		JWT:      loadJWTConfig(),
		Redis:    loadRedisConfig(),
		Upload:   loadUploadConfig(),
		SMTP:     loadSMTPConfig(),
		API:      loadAPIConfig(),
		Logging:  loadLoggingConfig(),
		CORS:     loadCORSConfig(),
	}

	// Validate required configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// loadAppConfig loads application configuration
func loadAppConfig() AppConfig {
	return AppConfig{
		Name:        getEnv("APP_NAME", "Barbershop API"),
		Version:     getEnv("APP_VERSION", "1.0.0"),
		Environment: getEnv("APP_ENV", "development"),
	}
}

// loadServerConfig loads server configuration
func loadServerConfig() ServerConfig {
	return ServerConfig{
		Port:         getEnv("PORT", "8080"),
		Host:         getEnv("HOST", "localhost"),
		GinMode:      getEnv("GIN_MODE", "debug"),
		ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
		WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
	}
}

// loadDatabaseConfig loads database configuration
func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		URL:             getEnv("DATABASE_URL", ""),
		MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
	}
}

// loadJWTConfig loads JWT configuration
func loadJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:     getEnv("JWT_SECRET", ""),
		Expiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
	}
}

// loadRedisConfig loads Redis configuration
func loadRedisConfig() RedisConfig {
	return RedisConfig{
		URL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getIntEnv("REDIS_DB", 0),
	}
}

// loadUploadConfig loads upload configuration
func loadUploadConfig() UploadConfig {
	return UploadConfig{
		Directory:   getEnv("UPLOAD_DIR", "./uploads"),
		MaxFileSize: getInt64Env("MAX_UPLOAD_SIZE", 10485760), // 10MB
	}
}

// loadSMTPConfig loads SMTP configuration
func loadSMTPConfig() SMTPConfig {
	return SMTPConfig{
		Host:     getEnv("SMTP_HOST", ""),
		Port:     getIntEnv("SMTP_PORT", 587),
		Username: getEnv("SMTP_USERNAME", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
		From:     getEnv("SMTP_FROM", "noreply@barbershop.com"),
	}
}

// loadAPIConfig loads API configuration
func loadAPIConfig() APIConfig {
	return APIConfig{
		RateLimit: getIntEnv("API_RATE_LIMIT", 100),
		Timeout:   getDurationEnv("API_TIMEOUT", 30*time.Second),
	}
}

// loadLoggingConfig loads logging configuration
func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:  getEnv("LOG_LEVEL", "info"),
		Format: getEnv("LOG_FORMAT", "json"),
	}
}

// loadCORSConfig loads CORS configuration
func loadCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: getSliceEnv("CORS_ALLOWED_ORIGINS", []string{"*"}),
		AllowedMethods: getSliceEnv("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}),
		AllowedHeaders: getSliceEnv("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
	}
}

// validateConfig validates required configuration fields
func validateConfig(config *Config) error {
	var errors []string

	// Required fields
	if config.Database.URL == "" {
		errors = append(errors, "DATABASE_URL is required")
	}

	if config.JWT.Secret == "" {
		errors = append(errors, "JWT_SECRET is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

// Helper functions for environment variable parsing

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getIntEnv gets an integer environment variable with a fallback value
func getIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using fallback: %d", key, value, fallback)
	}
	return fallback
}

// getInt64Env gets an int64 environment variable with a fallback value
func getInt64Env(key string, fallback int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid int64 value for %s: %s, using fallback: %d", key, value, fallback)
	}
	return fallback
}

// getDurationEnv gets a duration environment variable with a fallback value
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Warning: Invalid duration value for %s: %s, using fallback: %v", key, value, fallback)
	}
	return fallback
}

// getSliceEnv gets a comma-separated slice environment variable with a fallback value
func getSliceEnv(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, len(parts))
		for i, part := range parts {
			result[i] = strings.TrimSpace(part)
		}
		return result
	}
	return fallback
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsStaging returns true if running in staging environment
func (c *Config) IsStaging() bool {
	return c.App.Environment == "staging"
}

// IsTest returns true if running in test environment
func (c *Config) IsTest() bool {
	return c.App.Environment == "test"
}

// IsProductionLike returns true if production or staging
func (c *Config) IsProductionLike() bool {
	return c.IsProduction() || c.IsStaging()
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}
