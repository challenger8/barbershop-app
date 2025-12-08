// internal/repository/user_repository.go
package repository

import (
	"barber-booking-system/internal/config"
	"barber-booking-system/internal/models"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// UserRepository handles user data operations
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	query := `SELECT ` + USER_COLUMNS + ` FROM users ` + WHERE_ACTIVE + ` AND id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id %d: %w", id, err)
	}

	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT ` + USER_COLUMNS + ` FROM users ` + WHERE_ACTIVE + ` AND email = $1`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

// FindByUUID retrieves a user by UUID
func (r *UserRepository) FindByUUID(ctx context.Context, uuid string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, uuid, email, password_hash, name, phone, user_type, status,
		       email_verified, phone_verified, two_factor_enabled, 
		       failed_login_attempts, locked_until,
		       date_of_birth, gender, profile_picture_url,
		       address, city, state, country, postal_code, latitude, longitude,
		       preferences, notification_settings,
		       created_at, updated_at, last_login_at, created_by, deleted_at
		FROM users
		WHERE uuid = $1 AND deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, &user, query, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by uuid: %w", err)
	}

	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			uuid, email, password_hash, name, phone, user_type, status,
			email_verified, phone_verified, two_factor_enabled,
			date_of_birth, gender, profile_picture_url,
			address, city, state, country, postal_code, latitude, longitude,
			preferences, notification_settings,
			created_at, updated_at
		) VALUES (
			:uuid, :email, :password_hash, :name, :phone, :user_type, :status,
			:email_verified, :phone_verified, :two_factor_enabled,
			:date_of_birth, :gender, :profile_picture_url,
			:address, :city, :state, :country, :postal_code, :latitude, :longitude,
			:preferences, :notification_settings,
			:created_at, :updated_at
		) RETURNING id
	`

	// Single line replaces 3 lines
	SetCreateTimestamps(&user.CreatedAt, &user.UpdatedAt)

	// Single line replaces if blocks
	SetDefaultString(&user.Status, config.UserStatusActive)
	SetDefaultString(&user.UserType, config.UserTypeCustomer)

	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		// Check for duplicate email constraint violation
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique constraint") {
			if strings.Contains(strings.ToLower(err.Error()), "email") {
				return ErrDuplicateEmail
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&user.ID); err != nil {
			return fmt.Errorf("failed to scan user id: %w", err)
		}
	}

	return nil
}

// Update updates user information
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET
			name = :name,
			phone = :phone,
			date_of_birth = :date_of_birth,
			gender = :gender,
			profile_picture_url = :profile_picture_url,
			address = :address,
			city = :city,
			state = :state,
			country = :country,
			postal_code = :postal_code,
			latitude = :latitude,
			longitude = :longitude,
			preferences = :preferences,
			notification_settings = :notification_settings,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		// Check for duplicate email
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique constraint") {
			if strings.Contains(strings.ToLower(err.Error()), "email") {
				return ErrDuplicateEmail
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int, hashedPassword string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET last_login_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// IncrementFailedLoginAttempts increments failed login attempts counter
func (r *UserRepository) IncrementFailedLoginAttempts(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET failed_login_attempts = failed_login_attempts + 1,
		    updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to increment failed login attempts: %w", err)
	}

	return nil
}

// ResetFailedLoginAttempts resets failed login attempts counter
func (r *UserRepository) ResetFailedLoginAttempts(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET failed_login_attempts = 0,
		    locked_until = NULL,
		    updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to reset failed login attempts: %w", err)
	}

	return nil
}

// LockAccount locks user account temporarily
func (r *UserRepository) LockAccount(ctx context.Context, userID int, duration time.Duration) error {
	lockedUntil := time.Now().Add(duration)

	query := `
		UPDATE users
		SET locked_until = $1,
		    status = 'suspended',
		    updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, lockedUntil, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to lock account: %w", err)
	}

	return nil
}

// IsAccountLocked checks if account is currently locked
func (r *UserRepository) IsAccountLocked(ctx context.Context, userID int) (bool, error) {
	var lockedUntil *time.Time

	query := `
		SELECT locked_until
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, &lockedUntil, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrUserNotFound
		}
		return false, fmt.Errorf("failed to check account lock status: %w", err)
	}

	// If locked_until is NULL or in the past, account is not locked
	if lockedUntil == nil || lockedUntil.Before(time.Now()) {
		// If it was locked but time has passed, unlock it
		if lockedUntil != nil && lockedUntil.Before(time.Now()) {
			_ = r.ResetFailedLoginAttempts(ctx, userID)
		}
		return false, nil
	}

	return true, nil
}

// EmailExists checks if email already exists in database
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users 
			WHERE email = $1 AND deleted_at IS NULL
		)
	`

	err := r.db.GetContext(ctx, &exists, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET deleted_at = $1,
		    status = 'deleted',
		    updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
