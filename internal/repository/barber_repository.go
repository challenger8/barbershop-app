// internal/repository/barber_repository.go
package repository

import (
	"barber-booking-system/internal/config"
	"barber-booking-system/internal/models"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// BarberRepository handles barber data operations
type BarberRepository struct {
	db *sqlx.DB
}

// NewBarberRepository creates a new barber repository
func NewBarberRepository(db *sqlx.DB) *BarberRepository {
	return &BarberRepository{db: db}
}

// FindAll retrieves all barbers with optional filters
func (r *BarberRepository) FindAll(ctx context.Context, filters BarberFilters) ([]models.Barber, error) {
	// Define sort column mappings
	sortMap := map[string]string{
		"rating":         "b.rating DESC",
		"total_bookings": "b.total_bookings DESC",
		"shop_name":      "b.shop_name ASC",
		"user_name":      "u.name ASC",
		"default":        "b.created_at DESC",
	}

	// Build query using QueryBuilder
	qb := BuildBarberQuery().
		WhereIf(filters.Status != "", "b.status = ?", filters.Status).
		WhereIf(filters.City != "", "LOWER(b.city) = LOWER(?)", filters.City).
		WhereIf(filters.State != "", "LOWER(b.state) = LOWER(?)", filters.State).
		WhereIf(filters.MinRating > 0, "b.rating >= ?", filters.MinRating)

	// Handle pointer fields safely
	if filters.IsVerified != nil {
		qb.Where("b.is_verified = ?", *filters.IsVerified)
	}

	// Add search across multiple fields
	if filters.Search != "" {
		qb.Search([]string{
			"b.shop_name",
			"b.description",
			"u.name",
			"b.address",
			"b.city",
			"b.state",
		}, filters.Search).
			SearchILike([]string{
				"b.specialties",
			}, filters.Search)
	}

	// Add sorting and pagination
	query, args := qb.
		OrderByWithDefault(filters.SortBy, "default", sortMap).
		Paginate(filters.Limit, filters.Offset).
		Build()

	// Execute query
	var barbers []models.Barber
	err := r.db.SelectContext(ctx, &barbers, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch barbers: %w", err)
	}

	return barbers, nil
}

// FindByID retrieves a barber by ID
func (r *BarberRepository) FindByID(ctx context.Context, id int) (*models.Barber, error) {
	query := `
		SELECT b.*, u.name as user_name, u.email as user_email
		FROM barbers b
		LEFT JOIN users u ON b.user_id = u.id
		WHERE b.id = $1 AND b.deleted_at IS NULL
	`

	var barber models.Barber
	err := r.db.GetContext(ctx, &barber, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBarberNotFound
		}
		return nil, fmt.Errorf("failed to fetch barber: %w", err)
	}

	return &barber, nil
}

// FindByUUID retrieves a barber by UUID
func (r *BarberRepository) FindByUUID(ctx context.Context, uuid string) (*models.Barber, error) {
	query := `
		SELECT b.*, u.name as user_name, u.email as user_email
		FROM barbers b
		LEFT JOIN users u ON b.user_id = u.id
		WHERE b.uuid = $1 AND b.deleted_at IS NULL
	`

	var barber models.Barber
	err := r.db.GetContext(ctx, &barber, query, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBarberNotFound
		}
		return nil, fmt.Errorf("failed to fetch barber: %w", err)
	}

	return &barber, nil
}

// FindByUserID retrieves a barber by user ID
func (r *BarberRepository) FindByUserID(ctx context.Context, userID int) (*models.Barber, error) {
	query := `
		SELECT * FROM barbers 
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	var barber models.Barber
	err := r.db.GetContext(ctx, &barber, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBarberNotFound
		}
		return nil, fmt.Errorf("failed to fetch barber: %w", err)
	}

	return &barber, nil
}

// Create creates a new barber
func (r *BarberRepository) Create(ctx context.Context, barber *models.Barber) error {
	query := `
		INSERT INTO barbers (
			user_id, uuid, shop_name, business_name, business_registration_number,
			tax_id, address, address_line_2, city, state, country, postal_code,
			latitude, longitude, phone, business_email, website_url, description,
			years_experience, specialties, certifications, languages_spoken,
			profile_image_url, cover_image_url, gallery_images, working_hours,
			rating, total_reviews, total_bookings, response_time_minutes,
			acceptance_rate, cancellation_rate, status, is_verified,
			advance_booking_days, min_booking_notice_hours, auto_accept_bookings,
			instant_booking_enabled, commission_rate, payout_method, payout_details,
			created_at, updated_at, last_active_at
		) VALUES (
			:user_id, :uuid, :shop_name, :business_name, :business_registration_number,
			:tax_id, :address, :address_line_2, :city, :state, :country, :postal_code,
			:latitude, :longitude, :phone, :business_email, :website_url, :description,
			:years_experience, :specialties, :certifications, :languages_spoken,
			:profile_image_url, :cover_image_url, :gallery_images, :working_hours,
			:rating, :total_reviews, :total_bookings, :response_time_minutes,
			:acceptance_rate, :cancellation_rate, :status, :is_verified,
			:advance_booking_days, :min_booking_notice_hours, :auto_accept_bookings,
			:instant_booking_enabled, :commission_rate, :payout_method, :payout_details,
			:created_at, :updated_at, :last_active_at
		) RETURNING id
	`

	// Set timestamps
	now := time.Now()
	barber.CreatedAt = now
	barber.UpdatedAt = now
	barber.LastActiveAt = now

	// Set default values
	if barber.Status == "" {
		barber.Status = config.BarberStatusPending
	}
	if barber.Rating == 0 {
		barber.Rating = 0.0
	}

	rows, err := r.db.NamedQueryContext(ctx, query, barber)
	if err != nil {
		// Check for duplicate user_id (one barber profile per user)
		if IsFieldDuplicate(err, "user_id") {
			return ErrDuplicateBarber
		}
		return fmt.Errorf("failed to create barber: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&barber.ID); err != nil {
			return fmt.Errorf("failed to scan barber id: %w", err)
		}
	}

	return nil
}

// Update updates a barber
func (r *BarberRepository) Update(ctx context.Context, barber *models.Barber) error {
	barber.UpdatedAt = time.Now()

	query := `
		UPDATE barbers SET
			shop_name = :shop_name,
			business_name = :business_name,
			business_registration_number = :business_registration_number,
			tax_id = :tax_id,
			address = :address,
			address_line_2 = :address_line_2,
			city = :city,
			state = :state,
			country = :country,
			postal_code = :postal_code,
			latitude = :latitude,
			longitude = :longitude,
			phone = :phone,
			business_email = :business_email,
			website_url = :website_url,
			description = :description,
			years_experience = :years_experience,
			specialties = :specialties,
			certifications = :certifications,
			languages_spoken = :languages_spoken,
			profile_image_url = :profile_image_url,
			cover_image_url = :cover_image_url,
			gallery_images = :gallery_images,
			working_hours = :working_hours,
			advance_booking_days = :advance_booking_days,
			min_booking_notice_hours = :min_booking_notice_hours,
			auto_accept_bookings = :auto_accept_bookings,
			instant_booking_enabled = :instant_booking_enabled,
			payout_method = :payout_method,
			payout_details = :payout_details,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	result, err := r.db.NamedExecContext(ctx, query, barber)
	if err != nil {
		return fmt.Errorf("failed to update barber: %w", err)
	}

	return CheckRowsAffected(result, ErrBarberNotFound)
}

// Delete soft deletes a barber
func (r *BarberRepository) Delete(ctx context.Context, id int) error {
	query := `
		UPDATE barbers 
		SET deleted_at = $1, updated_at = $1, status = 'inactive'
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete barber: %w", err)
	}

	return CheckRowsAffected(result, ErrBarberNotFound)
}

// UpdateStatus updates barber status
func (r *BarberRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE barbers 
		SET status = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return CheckRowsAffected(result, ErrBarberNotFound)
}

// UpdateLastActive updates last active timestamp
func (r *BarberRepository) UpdateLastActive(ctx context.Context, id int) error {
	query := `UPDATE barbers SET last_active_at = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last active: %w", err)
	}
	return CheckRowsAffected(result, ErrBarberNotFound)
}

// GetStatistics retrieves barber statistics
func (r *BarberRepository) GetStatistics(ctx context.Context, id int) (*BarberStatistics, error) {
	query := `
		SELECT 
			COALESCE(COUNT(DISTINCT b.id), 0) as total_bookings,
			COALESCE(COUNT(DISTINCT CASE WHEN b.status = 'completed' THEN b.id END), 0) as completed_bookings,
			COALESCE(COUNT(DISTINCT CASE WHEN b.status IN ('cancelled_by_customer', 'cancelled_by_barber') THEN b.id END), 0) as cancelled_bookings,
			COALESCE(COUNT(DISTINCT r.id), 0) as total_reviews,
			COALESCE(AVG(r.overall_rating), 0) as average_rating,
			COALESCE(SUM(CASE WHEN b.status = 'completed' THEN b.total_price ELSE 0 END), 0) as total_revenue
		FROM barbers bar
		LEFT JOIN bookings b ON bar.id = b.barber_id
		LEFT JOIN reviews r ON bar.id = r.barber_id AND r.is_published = true
		WHERE bar.id = $1
		GROUP BY bar.id
	`

	var stats BarberStatistics
	err := r.db.GetContext(ctx, &stats, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return &stats, nil
}

// BarberFilters represents filter options for barbers
type BarberFilters struct {
	Status     string
	Name       string
	IsVerified *bool
	City       string
	State      string
	MinRating  float64
	Search     string
	SortBy     string
	Limit      int
	Offset     int
}

// BarberStatistics represents barber statistics
type BarberStatistics struct {
	TotalBookings     int     `db:"total_bookings" json:"total_bookings"`
	CompletedBookings int     `db:"completed_bookings" json:"completed_bookings"`
	CancelledBookings int     `db:"cancelled_bookings" json:"cancelled_bookings"`
	TotalReviews      int     `db:"total_reviews" json:"total_reviews"`
	AverageRating     float64 `db:"average_rating" json:"average_rating"`
	TotalRevenue      float64 `db:"total_revenue" json:"total_revenue"`
}
