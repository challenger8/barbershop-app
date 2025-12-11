// internal/logger/logger.go
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// ========================================================================
// STRUCTURED LOGGER - Production-grade logging
// ========================================================================
// Features:
// - Log levels (DEBUG, INFO, WARN, ERROR, FATAL)
// - JSON output for production (parseable by log aggregators)
// - Pretty console output for development
// - Zero-allocation logging (high performance)
// - Structured fields for queryable logs
// ========================================================================

// Config holds logger configuration
type Config struct {
	// Level is the minimum log level (debug, info, warn, error)
	Level string
	// Format is the output format (json, console)
	Format string
	// Output is the writer to output logs to (default: os.Stdout)
	Output io.Writer
	// ServiceName is included in every log entry
	ServiceName string
	// Environment is included in every log entry
	Environment string
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		Level:       "info",
		Format:      "json",
		Output:      os.Stdout,
		ServiceName: "barbershop-api",
		Environment: "development",
	}
}

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	zl zerolog.Logger
}

// New creates a new Logger with the given configuration
func New(cfg Config) *Logger {
	// Set global time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Configure output
	var output io.Writer = cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// Use pretty console output for development
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: "15:04:05",
		}
	}

	// Create base logger with common fields
	zl := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("environment", cfg.Environment).
		Logger()

	return &Logger{zl: zl}
}

// ========================================================================
// LOG METHODS
// ========================================================================

// Debug logs a debug message with optional fields
func (l *Logger) Debug(msg string) *zerolog.Event {
	return l.zl.Debug().Str("level", "debug")
}

// Info logs an info message with optional fields
func (l *Logger) Info(msg string) *zerolog.Event {
	return l.zl.Info().Str("level", "info")
}

// Warn logs a warning message with optional fields
func (l *Logger) Warn(msg string) *zerolog.Event {
	return l.zl.Warn().Str("level", "warn")
}

// Error logs an error message with the error and optional fields
func (l *Logger) Error(err error) *zerolog.Event {
	return l.zl.Error().Err(err).Str("level", "error")
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(err error) *zerolog.Event {
	return l.zl.Fatal().Err(err).Str("level", "fatal")
}

// ========================================================================
// CONVENIENCE METHODS (for simpler logging)
// ========================================================================

// DebugMsg logs a simple debug message
func (l *Logger) DebugMsg(msg string) {
	l.zl.Debug().Msg(msg)
}

// InfoMsg logs a simple info message
func (l *Logger) InfoMsg(msg string) {
	l.zl.Info().Msg(msg)
}

// WarnMsg logs a simple warning message
func (l *Logger) WarnMsg(msg string) {
	l.zl.Warn().Msg(msg)
}

// ErrorMsg logs an error with message
func (l *Logger) ErrorMsg(err error, msg string) {
	l.zl.Error().Err(err).Msg(msg)
}

// ========================================================================
// CHILD LOGGER METHODS
// ========================================================================

// With creates a child logger with additional fields
func (l *Logger) With() *LoggerContext {
	return &LoggerContext{ctx: l.zl.With()}
}

// WithRequestID creates a child logger with request ID
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		zl: l.zl.With().Str("request_id", requestID).Logger(),
	}
}

// WithUserID creates a child logger with user ID
func (l *Logger) WithUserID(userID int) *Logger {
	return &Logger{
		zl: l.zl.With().Int("user_id", userID).Logger(),
	}
}

// WithComponent creates a child logger for a specific component
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		zl: l.zl.With().Str("component", component).Logger(),
	}
}

// ========================================================================
// LOGGER CONTEXT (for building child loggers)
// ========================================================================

// LoggerContext wraps zerolog.Context for building child loggers
type LoggerContext struct {
	ctx zerolog.Context
}

// Str adds a string field
func (lc *LoggerContext) Str(key, val string) *LoggerContext {
	lc.ctx = lc.ctx.Str(key, val)
	return lc
}

// Int adds an int field
func (lc *LoggerContext) Int(key string, val int) *LoggerContext {
	lc.ctx = lc.ctx.Int(key, val)
	return lc
}

// Float64 adds a float64 field
func (lc *LoggerContext) Float64(key string, val float64) *LoggerContext {
	lc.ctx = lc.ctx.Float64(key, val)
	return lc
}

// Bool adds a bool field
func (lc *LoggerContext) Bool(key string, val bool) *LoggerContext {
	lc.ctx = lc.ctx.Bool(key, val)
	return lc
}

// Err adds an error field
func (lc *LoggerContext) Err(err error) *LoggerContext {
	lc.ctx = lc.ctx.Err(err)
	return lc
}

// Logger finalizes and returns the child logger
func (lc *LoggerContext) Logger() *Logger {
	return &Logger{zl: lc.ctx.Logger()}
}

// ========================================================================
// GLOBAL LOGGER INSTANCE
// ========================================================================

var globalLogger *Logger

// Init initializes the global logger
func Init(cfg Config) {
	globalLogger = New(cfg)
}

// Global returns the global logger instance
func Global() *Logger {
	if globalLogger == nil {
		// Return a default logger if not initialized
		globalLogger = New(DefaultConfig())
	}
	return globalLogger
}

// ========================================================================
// PACKAGE-LEVEL CONVENIENCE FUNCTIONS
// ========================================================================

// Debug logs a debug message using the global logger
func Debug() *zerolog.Event {
	return Global().zl.Debug()
}

// Info logs an info message using the global logger
func Info() *zerolog.Event {
	return Global().zl.Info()
}

// Warn logs a warning message using the global logger
func Warn() *zerolog.Event {
	return Global().zl.Warn()
}

// Error logs an error using the global logger
func Error(err error) *zerolog.Event {
	return Global().zl.Error().Err(err)
}
