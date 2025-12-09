// internal/repository/helpers.go
package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ========================================================================
// REPOSITORY HELPER FUNCTIONS - Eliminate DRY Violations
// ========================================================================
//
// These helpers consolidate repetitive patterns found across all repositories:
// 1. CheckRowsAffected - Repeated 15+ times
// 2. IsDuplicateError - Repeated 8+ times
// 3. IsFieldDuplicate - Helper for specific field duplicates
//
// Benefits:
// - Removes ~100 lines of duplicated code
// - Single source of truth for error handling
// - Easy to maintain and extend
// - Consistent error messages
// ========================================================================

// CheckRowsAffected checks if any rows were affected by a database operation
// and returns the appropriate "not found" error if no rows were affected.
//
// Usage:
//
//	result, err := r.db.ExecContext(ctx, query, args...)
//	if err != nil {
//	    return fmt.Errorf("failed to update: %w", err)
//	}
//	return CheckRowsAffected(result, ErrUserNotFound)
func CheckRowsAffected(result sql.Result, notFoundErr error) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return notFoundErr
	}

	return nil
}

// IsDuplicateError checks if an error is caused by a unique constraint violation.
// Works with PostgreSQL duplicate key errors.
//
// Usage:
//
//	if err != nil {
//	    if IsDuplicateError(err) {
//	        return ErrDuplicateEmail
//	    }
//	    return fmt.Errorf("failed to create: %w", err)
//	}
func IsDuplicateError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "duplicate") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "violates unique")
}

// IsFieldDuplicate checks if a specific field caused the duplicate error.
// Useful for returning specific error messages.
//
// Usage:
//
//	if err != nil {
//	    if IsFieldDuplicate(err, "email") {
//	        return ErrDuplicateEmail
//	    }
//	    if IsFieldDuplicate(err, "slug") {
//	        return ErrDuplicateSlug
//	    }
//	    return fmt.Errorf("failed to create: %w", err)
//	}
func IsFieldDuplicate(err error, fieldName string) bool {
	if !IsDuplicateError(err) {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, strings.ToLower(fieldName))
}

// ========================================================================
// ADDITIONAL HELPERS (Future Expansion)
// ========================================================================

// SetCreateTimestamps sets created_at and updated_at to now
func SetCreateTimestamps(createdAt, updatedAt *time.Time) {
	now := time.Now()
	*createdAt = now
	*updatedAt = now
}

// SetUpdateTimestamp sets updated_at to now
func SetUpdateTimestamp(updatedAt *time.Time) {
	*updatedAt = time.Now()
}

// SetDefaultString sets a default value if the string is empty
func SetDefaultString(target *string, defaultValue string) {
	if *target == "" {
		*target = defaultValue
	}
}

// SetDefaultInt sets a default value if the int is zero
func SetDefaultInt(target *int, defaultValue int) {
	if *target == 0 {
		*target = defaultValue
	}
}
