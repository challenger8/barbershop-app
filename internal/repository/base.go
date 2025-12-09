// internal/repository/base.go
package repository

import (
	"time"
)

// ========================================================================
// COMMON REPOSITORY HELPERS - Reduce repetition across repositories
// ========================================================================

// SetDefaultFloat sets a default value if the field is zero
func SetDefaultFloat(field *float64, defaultValue float64) {
	if *field == 0 {
		*field = defaultValue
	}
}

// ========================================================================
// QUERY BUILDER HELPERS - Simplify dynamic query building
// ========================================================================

// QueryBuilder helps build dynamic SQL queries with conditions

// AddCondition adds a WHERE condition
func (qb *QueryBuilder) AddCondition(condition string, value interface{}) {
	qb.conditions = append(qb.conditions, condition)
	qb.args = append(qb.args, value)
	qb.argCount++
}

// AddOptionalCondition adds a condition only if value is not zero/nil
func (qb *QueryBuilder) AddOptionalCondition(condition string, value interface{}, checkFunc func(interface{}) bool) {
	if checkFunc(value) {
		qb.AddCondition(condition, value)
	}
}

// ========================================================================
// VALIDATION HELPERS
// ========================================================================

// IsNotEmpty returns true if string is not empty
func IsNotEmpty(v interface{}) bool {
	if s, ok := v.(string); ok {
		return s != ""
	}
	return false
}

// IsPositive returns true if integer is greater than zero
func IsPositive(v interface{}) bool {
	if i, ok := v.(int); ok {
		return i > 0
	}
	return false
}

// IsNotZeroTime returns true if time is not zero
func IsNotZeroTime(v interface{}) bool {
	if t, ok := v.(time.Time); ok {
		return !t.IsZero()
	}
	return false
}
