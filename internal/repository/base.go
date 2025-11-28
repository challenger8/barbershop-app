// internal/repository/base.go
package repository

import (
	"time"
)

// ========================================================================
// COMMON REPOSITORY HELPERS - Reduce repetition across repositories
// ========================================================================

// SetCreateTimestamps sets CreatedAt and UpdatedAt to current time
// Use this in all Create methods to eliminate repetition
func SetCreateTimestamps(createdAt, updatedAt *time.Time) {
	now := time.Now()
	*createdAt = now
	*updatedAt = now
}

// SetUpdateTimestamp sets UpdatedAt to current time
// Use this in all Update methods
func SetUpdateTimestamp(updatedAt *time.Time) {
	*updatedAt = time.Now()
}

// SetDefaultString sets a default value if the field is empty
func SetDefaultString(field *string, defaultValue string) {
	if *field == "" {
		*field = defaultValue
	}
}

// SetDefaultInt sets a default value if the field is zero
func SetDefaultInt(field *int, defaultValue int) {
	if *field == 0 {
		*field = defaultValue
	}
}

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
type QueryBuilder struct {
	baseQuery string
	args      []interface{}
	argCount  int
	conditions []string
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery: baseQuery,
		args:      make([]interface{}, 0),
		argCount:  1,
		conditions: make([]string, 0),
	}
}

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

// Build constructs the final query string
func (qb *QueryBuilder) Build() (string, []interface{}) {
	query := qb.baseQuery
	if len(qb.conditions) > 0 {
		query += " WHERE " + qb.conditions[0]
		for i := 1; i < len(qb.conditions); i++ {
			query += " AND " + qb.conditions[i]
		}
	}
	return query, qb.args
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