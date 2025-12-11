// internal/repository/base_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ========================================================================
// BASE REPOSITORY WITH GENERICS
// ========================================================================
// Provides common CRUD operations to reduce code duplication.
// Uses Go 1.18+ generics for type safety.
// ========================================================================

// Entity is a constraint for database models
type Entity interface {
	TableName() string
	GetID() int
}

// BaseRepository provides common CRUD operations for any entity
type BaseRepository[T Entity] struct {
	db          *sqlx.DB
	tableName   string
	notFoundErr error
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T Entity](db *sqlx.DB, notFoundErr error) *BaseRepository[T] {
	var entity T
	return &BaseRepository[T]{
		db:          db,
		tableName:   entity.TableName(),
		notFoundErr: notFoundErr,
	}
}

// DB returns the underlying database connection
func (r *BaseRepository[T]) DB() *sqlx.DB {
	return r.db
}

// ========================================================================
// READ OPERATIONS
// ========================================================================

// FindByID retrieves an entity by ID
func (r *BaseRepository[T]) FindByID(ctx context.Context, id int) (*T, error) {
	var entity T
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND deleted_at IS NULL", r.tableName)

	err := r.db.GetContext(ctx, &entity, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, r.notFoundErr
		}
		return nil, fmt.Errorf("failed to find %s by id: %w", r.tableName, err)
	}
	return &entity, nil
}

// FindByUUID retrieves an entity by UUID
func (r *BaseRepository[T]) FindByUUID(ctx context.Context, uuid string) (*T, error) {
	var entity T
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid = $1 AND deleted_at IS NULL", r.tableName)

	err := r.db.GetContext(ctx, &entity, query, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, r.notFoundErr
		}
		return nil, fmt.Errorf("failed to find %s by uuid: %w", r.tableName, err)
	}
	return &entity, nil
}

// Exists checks if an entity exists by ID
func (r *BaseRepository[T]) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND deleted_at IS NULL)", r.tableName)

	err := r.db.GetContext(ctx, &exists, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return exists, nil
}

// ExistsByUUID checks if an entity exists by UUID
func (r *BaseRepository[T]) ExistsByUUID(ctx context.Context, uuid string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE uuid = $1 AND deleted_at IS NULL)", r.tableName)

	err := r.db.GetContext(ctx, &exists, query, uuid)
	if err != nil {
		return false, fmt.Errorf("failed to check existence by uuid: %w", err)
	}
	return exists, nil
}

// Count returns the total count of non-deleted entities
func (r *BaseRepository[T]) Count(ctx context.Context) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL", r.tableName)

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count %s: %w", r.tableName, err)
	}
	return count, nil
}

// ========================================================================
// DELETE OPERATIONS
// ========================================================================

// SoftDelete soft-deletes an entity by setting deleted_at
func (r *BaseRepository[T]) SoftDelete(ctx context.Context, id int) error {
	query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL", r.tableName)

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete %s: %w", r.tableName, err)
	}
	return CheckRowsAffected(result, r.notFoundErr)
}

// HardDelete permanently deletes an entity
func (r *BaseRepository[T]) HardDelete(ctx context.Context, id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", r.tableName)

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete %s: %w", r.tableName, err)
	}
	return CheckRowsAffected(result, r.notFoundErr)
}

// ========================================================================
// TRANSACTION SUPPORT
// ========================================================================

// BeginTx starts a new transaction
func (r *BaseRepository[T]) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// FindByIDTx retrieves an entity by ID within a transaction (with FOR UPDATE lock)
func (r *BaseRepository[T]) FindByIDForUpdate(ctx context.Context, tx *sqlx.Tx, id int) (*T, error) {
	var entity T
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND deleted_at IS NULL FOR UPDATE", r.tableName)

	err := tx.GetContext(ctx, &entity, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, r.notFoundErr
		}
		return nil, fmt.Errorf("failed to find %s by id for update: %w", r.tableName, err)
	}
	return &entity, nil
}
