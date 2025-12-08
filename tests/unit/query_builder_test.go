// tests/unit/repository/query_builder_test.go
package repository

import (
	"strings"
	"testing"

	"barber-booking-system/internal/repository"
)

// ========================================================================
// QUERY BUILDER UNIT TESTS
// ========================================================================

func TestQueryBuilder_BasicWhere(t *testing.T) {
	qb := repository.NewQueryBuilder("SELECT * FROM users")
	query, args := qb.Where("id = ?", 123).Build()

	expectedQuery := "SELECT * FROM users WHERE id = $1"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 1 || args[0] != 123 {
		t.Errorf("Expected args [123], got %v", args)
	}
}

func TestQueryBuilder_MultipleWhere(t *testing.T) {
	qb := repository.NewQueryBuilder("SELECT * FROM users")
	query, args := qb.
		Where("id = ?", 123).
		Where("status = ?", "active").
		Where("email = ?", "test@example.com").
		Build()

	expectedQuery := "SELECT * FROM users WHERE id = $1 AND status = $2 AND email = $3"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
}

func TestQueryBuilder_WhereIf(t *testing.T) {
	// Test with condition true
	qb := repository.NewQueryBuilder("SELECT * FROM users")
	query, args := qb.
		WhereIf(true, "id = ?", 123).
		WhereIf(false, "status = ?", "active").
		Build()

	expectedQuery := "SELECT * FROM users WHERE id = $1"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

func TestQueryBuilder_WhereIn(t *testing.T) {
	qb := repository.NewQueryBuilder("SELECT * FROM users")
	query, args := qb.
		WhereIn("id", []interface{}{1, 2, 3, 4, 5}).
		Build()

	expectedQuery := "SELECT * FROM users WHERE id IN ($1, $2, $3, $4, $5)"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 5 {
		t.Errorf("Expected 5 args, got %d", len(args))
	}
}

func TestQueryBuilder_WhereBetween(t *testing.T) {
	qb := repository.NewQueryBuilder("SELECT * FROM bookings")
	query, args := qb.
		WhereBetween("scheduled_start_time", "2024-01-01", "2024-12-31").
		Build()

	expectedQuery := "SELECT * FROM bookings WHERE scheduled_start_time BETWEEN $1 AND $2"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

func TestQueryBuilder_WhereNull(t *testing.T) {
	qb := repository.NewQueryBuilder("SELECT * FROM barbers")
	query, args := qb.
		WhereNull("deleted_at").
		Build()

	expectedQuery := "SELECT * FROM barbers WHERE deleted_at IS NULL"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 args, got %d", len(args))
	}
}

func TestQueryBuilder_Search(t *testing.T) {
	qb := repository.NewQueryBuilder("SELECT * FROM services")
	query, args := qb.
		Search([]string{"name", "description", "keywords"}, "haircut").
		Build()

	// Query should have LIKE conditions for all columns
	if !strings.Contains(query, "LOWER(name) LIKE") {
		t.Error("Expected query to contain LOWER(name) LIKE")
	}
	if !strings.Contains(query, "LOWER(description) LIKE") {
		t.Error("Expected query to contain LOWER(description) LIKE")
	}
	if !strings.Contains(query, "LOWER(keywords) LIKE") {
		t.Error("Expected query to contain LOWER(keywords) LIKE")
	}

	// Should have 3 args with wildcards
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}

	for _, arg := range args {
		strArg, ok := arg.(string)
		if !ok || !strings.Contains(strArg, "%") {
			t.Errorf("Expected wildcard in arg, got %v", arg)
		}
	}
}

func TestQueryBuilder_OrderBy(t *testing.T) {
	qb := repository.NewQueryBuilder("SELECT * FROM users")
	query, _ := qb.
		OrderBy("created_at", "DESC").
		Build()

	expectedQuery := "SELECT * FROM users ORDER BY created_at DESC"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}
}

func TestQueryBuilder_Paginate(t *testing.T) {
	qb := repository.NewQueryBuilder("SELECT * FROM users")
	query, args := qb.
		Paginate(20, 10).
		Build()

	expectedQuery := "SELECT * FROM users LIMIT $1 OFFSET $2"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 2 || args[0] != 20 || args[1] != 10 {
		t.Errorf("Expected args [20, 10], got %v", args)
	}
}

func TestQueryBuilder_ComplexQuery(t *testing.T) {
	qb := repository.NewQueryBuilder("SELECT * FROM bookings")
	query, args := qb.
		Where("customer_id = ?", 123).
		Where("status = ?", "confirmed").
		WhereBetween("total_price", 50.0, 200.0).
		WhereGreaterThan("rating", 4.0).
		Search([]string{"notes", "special_requests"}, "allergy").
		OrderBy("scheduled_start_time", "DESC").
		Paginate(20, 0).
		Build()

	// Check structure
	if !strings.Contains(query, "WHERE") {
		t.Error("Expected WHERE clause")
	}
	if !strings.Contains(query, "ORDER BY") {
		t.Error("Expected ORDER BY clause")
	}
	if !strings.Contains(query, "LIMIT") {
		t.Error("Expected LIMIT clause")
	}

	// Check arg count: 4 basic wheres + 2 between + 1 greater + 2 search + 2 pagination = 11
	expectedArgCount := 11
	if len(args) != expectedArgCount {
		t.Errorf("Expected %d args, got %d", expectedArgCount, len(args))
	}
}

func TestQueryBuilder_Join(t *testing.T) {
	qb := repository.NewQueryBuilder("SELECT * FROM users u")
	query, _ := qb.
		LeftJoin("profiles p", "p.user_id = u.id").
		InnerJoin("roles r", "r.id = u.role_id").
		Build()

	if !strings.Contains(query, "LEFT JOIN profiles p ON p.user_id = u.id") {
		t.Error("Expected LEFT JOIN clause")
	}
	if !strings.Contains(query, "INNER JOIN roles r ON r.id = u.role_id") {
		t.Error("Expected INNER JOIN clause")
	}
}

func TestQueryBuilder_EmptySearch(t *testing.T) {
	// Empty search should not add conditions
	qb := repository.NewQueryBuilder("SELECT * FROM services")
	query, args := qb.
		Search([]string{"name", "description"}, "").
		Build()

	// Should not have WHERE clause
	if strings.Contains(query, "WHERE") {
		t.Error("Empty search should not add WHERE conditions")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 args for empty search, got %d", len(args))
	}
}

func TestQueryBuilder_BuildServiceQuery(t *testing.T) {
	qb := repository.BuildServiceQuery()
	query, _ := qb.Build()

	// Should have base query with join
	if !strings.Contains(query, "FROM services s") {
		t.Error("Expected services table in query")
	}
	if !strings.Contains(query, "LEFT JOIN service_categories c") {
		t.Error("Expected category join in query")
	}
}

func TestQueryBuilder_BuildBarberQuery(t *testing.T) {
	qb := repository.BuildBarberQuery()
	query, _ := qb.Build()

	// Should have base query with join and deleted_at filter
	if !strings.Contains(query, "FROM barbers b") {
		t.Error("Expected barbers table in query")
	}
	if !strings.Contains(query, "LEFT JOIN users u") {
		t.Error("Expected users join in query")
	}
	if !strings.Contains(query, "deleted_at IS NULL") {
		t.Error("Expected deleted_at IS NULL condition")
	}
}

// ========================================================================
// BENCHMARK TESTS
// ========================================================================

func BenchmarkQueryBuilder_SimpleQuery(b *testing.B) {
	for i := 0; i < b.N; i++ {
		qb := repository.NewQueryBuilder("SELECT * FROM users")
		_, _ = qb.
			Where("id = ?", 123).
			Where("status = ?", "active").
			Build()
	}
}

func BenchmarkQueryBuilder_ComplexQuery(b *testing.B) {
	for i := 0; i < b.N; i++ {
		qb := repository.NewQueryBuilder("SELECT * FROM bookings")
		_, _ = qb.
			Where("customer_id = ?", 123).
			WhereBetween("total_price", 50.0, 200.0).
			Search([]string{"notes", "special_requests"}, "allergy").
			OrderBy("scheduled_start_time", "DESC").
			Paginate(20, 0).
			Build()
	}
}
