// config/database.go
package config

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	appConfig "barber-booking-system/internal/config"
)

// DatabaseManager manages database connections and operations
type DatabaseManager struct {
	DB     *sqlx.DB
	Config appConfig.DatabaseConfig
}

// NewDatabaseManager creates a new database manager using the provided configuration
func NewDatabaseManager(config appConfig.DatabaseConfig) (*DatabaseManager, error) {
	// Validate database URL
	if config.URL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseManager{
		DB:     db,
		Config: config,
	}, nil
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	return dm.DB.Close()
}

// Ping tests the database connection
func (dm *DatabaseManager) Ping() error {
	return dm.DB.Ping()
}

// Health checks database health
func (dm *DatabaseManager) Health() DatabaseHealth {
	health := DatabaseHealth{
		Status: "healthy",
	}

	// Test connection
	if err := dm.Ping(); err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		return health
	}

	// Get connection stats
	stats := dm.DB.Stats()
	health.Stats = &ConnectionStats{
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}

	// Check for connection pool issues
	if stats.OpenConnections >= dm.Config.MaxOpenConns {
		health.Status = "degraded"
		health.Warning = "Connection pool is at maximum capacity"
	} else if float64(stats.InUse)/float64(stats.OpenConnections) > 0.8 {
		health.Status = "degraded"
		health.Warning = "High connection utilization"
	}

	return health
}

// GetDatabaseInfo returns information about the database
func (dm *DatabaseManager) GetDatabaseInfo() (DatabaseInfo, error) {
	var info DatabaseInfo

	// Get database version
	err := dm.DB.Get(&info.Version, "SELECT version()")
	if err != nil {
		return info, fmt.Errorf("failed to get database version: %w", err)
	}

	// Get current database name
	err = dm.DB.Get(&info.DatabaseName, "SELECT current_database()")
	if err != nil {
		return info, fmt.Errorf("failed to get database name: %w", err)
	}

	// Get current user
	err = dm.DB.Get(&info.CurrentUser, "SELECT current_user")
	if err != nil {
		return info, fmt.Errorf("failed to get current user: %w", err)
	}

	// Get table count
	query := `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`
	err = dm.DB.Get(&info.TableCount, query)
	if err != nil {
		return info, fmt.Errorf("failed to get table count: %w", err)
	}

	return info, nil
}

// Transaction executes a function within a database transaction
func (dm *DatabaseManager) Transaction(fn func(*sqlx.Tx) error) error {
	tx, err := dm.DB.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// ConnectionStats represents database connection statistics
type ConnectionStats struct {
	OpenConnections   int           `json:"open_connections"`
	InUse             int           `json:"in_use"`
	Idle              int           `json:"idle"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
}

// DatabaseHealth represents database health status
type DatabaseHealth struct {
	Status  string           `json:"status"`
	Error   string           `json:"error,omitempty"`
	Warning string           `json:"warning,omitempty"`
	Stats   *ConnectionStats `json:"stats,omitempty"`
}

// DatabaseInfo represents database information
type DatabaseInfo struct {
	Version      string `json:"version" db:"version"`
	DatabaseName string `json:"database_name" db:"current_database"`
	CurrentUser  string `json:"current_user" db:"current_user"`
	TableCount   int    `json:"table_count" db:"count"`
}
