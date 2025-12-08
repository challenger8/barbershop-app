// internal/repository/query_builder.go
package repository

import (
	"fmt"
	"strings"

	"barber-booking-system/internal/config"
)

// ========================================================================
// QUERY BUILDER - Flexible SQL Query Construction
// ========================================================================
//
// Purpose: Eliminate repetitive query building code in repositories
// Benefits:
//   - No manual argCount tracking
//   - Type-safe query construction
//   - Reusable patterns
//   - Easy to test
//   - DRY principle
//
// Usage Example:
//   qb := NewQueryBuilder("SELECT * FROM services").
//       Where("category_id = ?", categoryID).
//       Where("is_active = ?", true).
//       Search([]string{"name", "description"}, searchTerm).
//       OrderBy("created_at", "DESC").
//       Paginate(20, 0).
//       Build()
// ========================================================================

// QueryBuilder helps construct SQL queries safely and cleanly
type QueryBuilder struct {
	baseQuery  string
	joins      []string
	conditions []string
	args       []interface{}
	orderBy    string
	limit      int
	offset     int
	argCount   int
}

// NewQueryBuilder creates a new query builder with base query
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery:  baseQuery,
		joins:      []string{},
		conditions: []string{},
		args:       []interface{}{},
		argCount:   1,
	}
}

// ========================================================================
// JOIN OPERATIONS
// ========================================================================

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(joinType, table, condition string) *QueryBuilder {
	joinClause := fmt.Sprintf("%s JOIN %s ON %s", joinType, table, condition)
	qb.joins = append(qb.joins, joinClause)
	return qb
}

// LeftJoin adds a LEFT JOIN clause
func (qb *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	return qb.Join("LEFT", table, condition)
}

// InnerJoin adds an INNER JOIN clause
func (qb *QueryBuilder) InnerJoin(table, condition string) *QueryBuilder {
	return qb.Join("INNER", table, condition)
}

// ========================================================================
// WHERE CONDITIONS
// ========================================================================

// Where adds a WHERE condition with argument
// Automatically handles argCount tracking
func (qb *QueryBuilder) Where(condition string, arg interface{}) *QueryBuilder {
	// Replace ? with $N for PostgreSQL
	condition = strings.Replace(condition, "?", fmt.Sprintf("$%d", qb.argCount), 1)
	qb.conditions = append(qb.conditions, condition)
	qb.args = append(qb.args, arg)
	qb.argCount++
	return qb
}

// WhereIf adds a WHERE condition only if shouldAdd is true
func (qb *QueryBuilder) WhereIf(shouldAdd bool, condition string, arg interface{}) *QueryBuilder {
	if shouldAdd {
		return qb.Where(condition, arg)
	}
	return qb
}

// WhereIn adds a WHERE IN condition
func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}

	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = fmt.Sprintf("$%d", qb.argCount)
		qb.args = append(qb.args, values[i])
		qb.argCount++
	}

	condition := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereBetween adds a BETWEEN condition
func (qb *QueryBuilder) WhereBetween(column string, min, max interface{}) *QueryBuilder {
	condition := fmt.Sprintf("%s BETWEEN $%d AND $%d", column, qb.argCount, qb.argCount+1)
	qb.conditions = append(qb.conditions, condition)
	qb.args = append(qb.args, min, max)
	qb.argCount += 2
	return qb
}

// WhereGreaterThan adds a > condition
func (qb *QueryBuilder) WhereGreaterThan(column string, value interface{}) *QueryBuilder {
	return qb.Where(column+" > ?", value)
}

// WhereLessThan adds a < condition
func (qb *QueryBuilder) WhereLessThan(column string, value interface{}) *QueryBuilder {
	return qb.Where(column+" < ?", value)
}

// ========================================================================
// SEARCH OPERATIONS
// ========================================================================

// Search adds LIKE conditions across multiple columns
// Automatically handles the repetitive search pattern
func (qb *QueryBuilder) Search(columns []string, searchTerm string) *QueryBuilder {
	if searchTerm == "" || len(columns) == 0 {
		return qb
	}

	// Add % wildcards for LIKE
	searchPattern := "%" + strings.ToLower(searchTerm) + "%"

	// Build OR conditions for each column
	searchConditions := make([]string, len(columns))
	for i, col := range columns {
		searchConditions[i] = fmt.Sprintf("LOWER(%s) LIKE $%d", col, qb.argCount)
		qb.args = append(qb.args, searchPattern)
		qb.argCount++
	}

	// Combine with OR
	combinedCondition := "(" + strings.Join(searchConditions, " OR ") + ")"
	qb.conditions = append(qb.conditions, combinedCondition)
	return qb
}

// SearchILike adds case-insensitive search for JSONB columns
func (qb *QueryBuilder) SearchILike(columns []string, searchTerm string) *QueryBuilder {
	if searchTerm == "" || len(columns) == 0 {
		return qb
	}

	searchPattern := "%" + searchTerm + "%"

	searchConditions := make([]string, len(columns))
	for i, col := range columns {
		searchConditions[i] = fmt.Sprintf("%s::text ILIKE $%d", col, qb.argCount)
		qb.args = append(qb.args, searchPattern)
		qb.argCount++
	}

	combinedCondition := "(" + strings.Join(searchConditions, " OR ") + ")"
	qb.conditions = append(qb.conditions, combinedCondition)
	return qb
}

// ========================================================================
// SORTING
// ========================================================================

// OrderBy sets the ORDER BY clause
func (qb *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	direction = strings.ToUpper(direction)
	if direction != "ASC" && direction != "DESC" {
		direction = "DESC"
	}
	qb.orderBy = fmt.Sprintf("%s %s", column, direction)
	return qb
}

// OrderByWithDefault sets ORDER BY with a default if sortBy is empty
func (qb *QueryBuilder) OrderByWithDefault(sortBy, defaultSort string, sortMap map[string]string) *QueryBuilder {
	// Use provided sortBy or default
	if sortBy == "" {
		sortBy = defaultSort
	}

	// Look up column mapping
	if column, exists := sortMap[sortBy]; exists {
		qb.orderBy = column
	} else {
		qb.orderBy = sortMap[defaultSort]
	}

	return qb
}

// ========================================================================
// PAGINATION
// ========================================================================

// Paginate adds LIMIT and OFFSET
func (qb *QueryBuilder) Paginate(limit, offset int) *QueryBuilder {
	if limit <= 0 {
		limit = config.DefaultPageLimit
	}
	if offset < 0 {
		offset = 0
	}

	qb.limit = limit
	qb.offset = offset
	return qb
}

// ========================================================================
// BUILD QUERY
// ========================================================================

func (qb *QueryBuilder) Build() (string, []interface{}) {
	query := qb.baseQuery

	// Add JOINs
	if len(qb.joins) > 0 {
		query += " " + strings.Join(qb.joins, " ")
	}

	// Add WHERE conditions
	if len(qb.conditions) > 0 {
		query += " WHERE " + strings.Join(qb.conditions, " AND ")
	}

	// Add ORDER BY
	if qb.orderBy != "" {
		query += " ORDER BY " + qb.orderBy
	}

	// Add LIMIT and OFFSET
	// IMPORTANT: Always add OFFSET when LIMIT is present (even if offset is 0)
	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", qb.argCount)
		qb.args = append(qb.args, qb.limit)
		qb.argCount++

		// Always add OFFSET when we have LIMIT (PostgreSQL best practice)
		query += fmt.Sprintf(" OFFSET $%d", qb.argCount)
		qb.args = append(qb.args, qb.offset)
		qb.argCount++
	}

	return query, qb.args
}

// ========================================================================
// HELPER FUNCTIONS FOR COMMON PATTERNS
// ========================================================================

// BuildServiceQuery creates a pre-configured builder for services
func BuildServiceQuery() *QueryBuilder {
	baseQuery := `
		SELECT s.*, c.name as category_name
		FROM services s
		LEFT JOIN service_categories c ON s.category_id = c.id
	`
	return NewQueryBuilder(baseQuery)
}

// WhereNull adds a WHERE column IS NULL condition
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	condition := fmt.Sprintf("%s IS NULL", column)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereNotNull adds a WHERE column IS NOT NULL condition
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	condition := fmt.Sprintf("%s IS NOT NULL", column)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// BuildBarberQuery creates a pre-configured builder for barbers
func BuildBarberQuery() *QueryBuilder {
	baseQuery := `
		SELECT b.*, u.name as user_name, u.email as user_email
		FROM barbers b
		LEFT JOIN users u ON b.user_id = u.id
	`
	return NewQueryBuilder(baseQuery).WhereNull("b.deleted_at")
}
