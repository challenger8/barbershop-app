// internal/repository/review_repository.go
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

// ========================================================================
// REVIEW REPOSITORY - Data Access Layer for Reviews
// ========================================================================

// ReviewRepository handles review data operations
type ReviewRepository struct {
	*BaseRepository[models.Review]
	db *sqlx.DB
}

// NewReviewRepository creates a new review repository
func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{
		BaseRepository: NewBaseRepository[models.Review](db, ErrReviewNotFound),
		db:             db,
	}
}

// ========================================================================
// FILTER STRUCTS
// ========================================================================

type ReviewFilters struct {
	// Identity filters
	CustomerID int `form:"customer_id"`
	BarberID   int `form:"barber_id"`
	BookingID  int `form:"booking_id"`

	// Rating filters
	MinRating int `form:"min_rating"`
	MaxRating int `form:"max_rating"`

	// Status filters
	ModerationStatus string   `form:"moderation_status"`
	Statuses         []string `form:"statuses"`
	IsPublished      *bool    `form:"is_published"`
	IsVerified       *bool    `form:"is_verified"`

	// Content filters
	HasComment  *bool `form:"has_comment"`
	HasImages   *bool `form:"has_images"`
	HasResponse *bool `form:"has_response"`

	// Date range filters
	CreatedFrom time.Time `form:"created_from" time_format:"2006-01-02T15:04:05Z07:00"`
	CreatedTo   time.Time `form:"created_to" time_format:"2006-01-02T15:04:05Z07:00"`

	// Search
	Search string `form:"search"`

	// Recommendation filters
	WouldRecommend *bool `form:"would_recommend"`

	// Sorting and pagination
	SortBy string `form:"sort_by"`
	Order  string `form:"order"`
	Limit  int    `form:"limit,default=50"`
	Offset int    `form:"offset,default=0"`

	// Relation loading flags (prevents N+1 queries)
	IncludeCustomer bool `form:"include_customer"`
	IncludeBarber   bool `form:"include_barber"`
	IncludeBooking  bool `form:"include_booking"`
}

// ReviewStats represents review statistics
type ReviewStats struct {
	TotalReviews     int     `json:"total_reviews" db:"total_reviews"`
	AverageRating    float64 `json:"average_rating" db:"average_rating"`
	FiveStarCount    int     `json:"five_star_count" db:"five_star_count"`
	FourStarCount    int     `json:"four_star_count" db:"four_star_count"`
	ThreeStarCount   int     `json:"three_star_count" db:"three_star_count"`
	TwoStarCount     int     `json:"two_star_count" db:"two_star_count"`
	OneStarCount     int     `json:"one_star_count" db:"one_star_count"`
	RecommendPercent float64 `json:"recommend_percent" db:"recommend_percent"`
}

// NOTE: Error variables are defined in errors.go to avoid duplication
// ErrReviewNotFound, ErrDuplicateReview, ErrInvalidRating,
// ErrBookingNotCompleted, ErrReviewAlreadyExists, ErrCannotModifyReview, ErrInvalidModeration

// ========================================================================
// VALID STATUS VALUES
// ========================================================================

// ValidModerationStatuses defines allowed moderation statuses - using config constants
var ValidModerationStatuses = []string{
	config.ReviewModerationPending,
	config.ReviewModerationApproved,
	config.ReviewModerationRejected,
	config.ReviewModerationFlagged,
}

// IsValidModerationStatus checks if a moderation status is valid
func IsValidModerationStatus(status string) bool {
	return IsValidValue(status, ValidModerationStatuses)
}

// ========================================================================
// CREATE OPERATIONS
// ========================================================================

// Create inserts a new review into the database
func (r *ReviewRepository) Create(ctx context.Context, review *models.Review) error {
	// Check for duplicate review
	exists, err := r.ExistsByBookingID(ctx, review.BookingID)
	if err != nil {
		return fmt.Errorf("failed to check existing review: %w", err)
	}
	if exists {
		return ErrDuplicateReview
	}

	query := `
		INSERT INTO reviews (
			booking_id, customer_id, barber_id,
			overall_rating, service_quality_rating, punctuality_rating,
			cleanliness_rating, value_for_money_rating, professionalism_rating,
			title, comment, pros, cons,
			would_recommend, would_book_again, service_as_expected, duration_accurate,
			images,
			is_verified, is_published, moderation_status,
			created_at, updated_at
		) VALUES (
			:booking_id, :customer_id, :barber_id,
			:overall_rating, :service_quality_rating, :punctuality_rating,
			:cleanliness_rating, :value_for_money_rating, :professionalism_rating,
			:title, :comment, :pros, :cons,
			:would_recommend, :would_book_again, :service_as_expected, :duration_accurate,
			:images,
			:is_verified, :is_published, :moderation_status,
			:created_at, :updated_at
		) RETURNING id
	`

	// Set timestamps and defaults using helpers
	SetCreateTimestamps(&review.CreatedAt, &review.UpdatedAt)
	SetDefaultString(&review.ModerationStatus, config.ReviewModerationPending)

	rows, err := r.db.NamedQueryContext(ctx, query, review)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return ErrDuplicateReview
		}
		return fmt.Errorf("failed to create review: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&review.ID); err != nil {
			return fmt.Errorf("failed to scan review id: %w", err)
		}
	}

	return nil
}

// ========================================================================
// READ OPERATIONS - FindByID
// ========================================================================

// FindByID retrieves a review by its ID
func (r *ReviewRepository) FindByID(ctx context.Context, id int) (*models.Review, error) {
	query := `SELECT * FROM reviews WHERE id = $1`

	var review models.Review
	err := r.db.GetContext(ctx, &review, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("failed to find review by id: %w", err)
	}

	return &review, nil
}

// FindByBookingID retrieves a review by booking ID
func (r *ReviewRepository) FindByBookingID(ctx context.Context, bookingID int) (*models.Review, error) {
	query := `SELECT * FROM reviews WHERE booking_id = $1`

	var review models.Review
	err := r.db.GetContext(ctx, &review, query, bookingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("failed to find review by booking id: %w", err)
	}

	return &review, nil
}

// ExistsByBookingID checks if a review exists for a booking
func (r *ReviewRepository) ExistsByBookingID(ctx context.Context, bookingID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM reviews WHERE booking_id = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, bookingID)
	if err != nil {
		return false, fmt.Errorf("failed to check review existence: %w", err)
	}

	return exists, nil
}

// ========================================================================
// READ OPERATIONS - FindAll with Filters
// ========================================================================

// FindAll retrieves reviews with optional filters
func (r *ReviewRepository) FindAll(ctx context.Context, filters ReviewFilters) ([]models.Review, error) {
	query := `SELECT * FROM reviews WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Customer filter
	if filters.CustomerID > 0 {
		query += fmt.Sprintf(" AND customer_id = $%d", argCount)
		args = append(args, filters.CustomerID)
		argCount++
	}

	// Barber filter
	if filters.BarberID > 0 {
		query += fmt.Sprintf(" AND barber_id = $%d", argCount)
		args = append(args, filters.BarberID)
		argCount++
	}

	// Booking filter
	if filters.BookingID > 0 {
		query += fmt.Sprintf(" AND booking_id = $%d", argCount)
		args = append(args, filters.BookingID)
		argCount++
	}

	// Rating filters
	if filters.MinRating > 0 {
		query += fmt.Sprintf(" AND overall_rating >= $%d", argCount)
		args = append(args, filters.MinRating)
		argCount++
	}
	if filters.MaxRating > 0 {
		query += fmt.Sprintf(" AND overall_rating <= $%d", argCount)
		args = append(args, filters.MaxRating)
		argCount++
	}

	// Moderation status filter
	if filters.ModerationStatus != "" {
		query += fmt.Sprintf(" AND moderation_status = $%d", argCount)
		args = append(args, filters.ModerationStatus)
		argCount++
	}

	// Multiple statuses filter
	if len(filters.Statuses) > 0 {
		placeholders := make([]string, len(filters.Statuses))
		for i, status := range filters.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, status)
			argCount++
		}
		query += fmt.Sprintf(" AND moderation_status IN (%s)", strings.Join(placeholders, ", "))
	}

	// Published filter
	if filters.IsPublished != nil {
		query += fmt.Sprintf(" AND is_published = $%d", argCount)
		args = append(args, *filters.IsPublished)
		argCount++
	}

	// Verified filter
	if filters.IsVerified != nil {
		query += fmt.Sprintf(" AND is_verified = $%d", argCount)
		args = append(args, *filters.IsVerified)
		argCount++
	}

	// Has comment filter
	if filters.HasComment != nil {
		if *filters.HasComment {
			query += " AND comment IS NOT NULL AND comment != ''"
		} else {
			query += " AND (comment IS NULL OR comment = '')"
		}
	}

	// Has images filter
	if filters.HasImages != nil {
		if *filters.HasImages {
			query += " AND images IS NOT NULL AND array_length(images, 1) > 0"
		} else {
			query += " AND (images IS NULL OR array_length(images, 1) = 0)"
		}
	}

	// Has barber response filter
	if filters.HasResponse != nil {
		if *filters.HasResponse {
			query += " AND barber_response IS NOT NULL AND barber_response != ''"
		} else {
			query += " AND (barber_response IS NULL OR barber_response = '')"
		}
	}

	// Date range filters
	if !filters.CreatedFrom.IsZero() {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, filters.CreatedFrom)
		argCount++
	}
	if !filters.CreatedTo.IsZero() {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, filters.CreatedTo)
		argCount++
	}

	// Search filter
	if filters.Search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR comment ILIKE $%d)", argCount, argCount+1)
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argCount += 2
	}

	// Would recommend filter
	if filters.WouldRecommend != nil {
		query += fmt.Sprintf(" AND would_recommend = $%d", argCount)
		args = append(args, *filters.WouldRecommend)
		argCount++
	}

	// Sorting
	orderBy := "created_at DESC" // Default sort
	if filters.SortBy != "" {
		order := "DESC"
		if filters.Order != "" && (filters.Order == "ASC" || filters.Order == "asc") {
			order = "ASC"
		}
		switch filters.SortBy {
		case "overall_rating":
			orderBy = fmt.Sprintf("overall_rating %s", order)
		case "helpful_votes":
			orderBy = fmt.Sprintf("helpful_votes %s", order)
		case "created_at":
			orderBy = fmt.Sprintf("created_at %s", order)
		}
	}
	query += " ORDER BY " + orderBy

	// Pagination
	limit := 50
	if filters.Limit > 0 {
		limit = filters.Limit
	}
	offset := 0
	if filters.Offset > 0 {
		offset = filters.Offset
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	var reviews []models.Review
	err := r.db.SelectContext(ctx, &reviews, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews: %w", err)
	}

	return reviews, nil
}
// ========================================================================
// READ OPERATIONS - FindAll with Relations (prevents N+1 queries)
// ========================================================================

// ReviewWithRelations represents a review with optional loaded relations
type ReviewWithRelations struct {
	models.Review
	CustomerName  *string `db:"customer_name"`
	CustomerEmail *string `db:"customer_email"`
	BarberName    *string `db:"barber_shop_name"`
	BarberCity    *string `db:"barber_city"`
	BookingNumber *string `db:"booking_number"`
	ServiceName   *string `db:"service_name"`
}

// FindAllWithRelations retrieves reviews with optional relation loading
// Use this instead of FindAll when you need customer/barber/booking info
func (r *ReviewRepository) FindAllWithRelations(ctx context.Context, filters ReviewFilters) ([]ReviewWithRelations, error) {
	// Build SELECT columns
	selectCols := `r.*`
	joins := ""

	if filters.IncludeCustomer {
		selectCols += `, u.name as customer_name, u.email as customer_email`
		joins += ` LEFT JOIN users u ON r.customer_id = u.id`
	}

	if filters.IncludeBarber {
		selectCols += `, b.shop_name as barber_shop_name, b.city as barber_city`
		joins += ` LEFT JOIN barbers b ON r.barber_id = b.id`
	}

	if filters.IncludeBooking {
		selectCols += `, bk.booking_number as booking_number, bk.service_name as service_name`
		joins += ` LEFT JOIN bookings bk ON r.booking_id = bk.id`
	}

	// Base query
	query := fmt.Sprintf(`SELECT %s FROM reviews r%s WHERE 1=1`, selectCols, joins)
	args := []interface{}{}
	argCount := 1

	// Apply filters (same as FindAll)
	if filters.CustomerID > 0 {
		query += fmt.Sprintf(" AND r.customer_id = $%d", argCount)
		args = append(args, filters.CustomerID)
		argCount++
	}

	if filters.BarberID > 0 {
		query += fmt.Sprintf(" AND r.barber_id = $%d", argCount)
		args = append(args, filters.BarberID)
		argCount++
	}

	if filters.BookingID > 0 {
		query += fmt.Sprintf(" AND r.booking_id = $%d", argCount)
		args = append(args, filters.BookingID)
		argCount++
	}

	if filters.MinRating > 0 {
		query += fmt.Sprintf(" AND r.overall_rating >= $%d", argCount)
		args = append(args, filters.MinRating)
		argCount++
	}

	if filters.MaxRating > 0 {
		query += fmt.Sprintf(" AND r.overall_rating <= $%d", argCount)
		args = append(args, filters.MaxRating)
		argCount++
	}

	if filters.ModerationStatus != "" {
		query += fmt.Sprintf(" AND r.moderation_status = $%d", argCount)
		args = append(args, filters.ModerationStatus)
		argCount++
	}

	if filters.IsPublished != nil {
		query += fmt.Sprintf(" AND r.is_published = $%d", argCount)
		args = append(args, *filters.IsPublished)
		argCount++
	}

	if filters.IsVerified != nil {
		query += fmt.Sprintf(" AND r.is_verified = $%d", argCount)
		args = append(args, *filters.IsVerified)
		argCount++
	}

	if filters.Search != "" {
		query += fmt.Sprintf(" AND (r.title ILIKE $%d OR r.comment ILIKE $%d)", argCount, argCount+1)
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argCount += 2
	}

	// Sorting
	orderBy := "r.created_at DESC"
	if filters.SortBy != "" {
		order := "DESC"
		if filters.Order == "ASC" || filters.Order == "asc" {
			order = "ASC"
		}
		switch filters.SortBy {
		case "created_at":
			orderBy = fmt.Sprintf("r.created_at %s", order)
		case "overall_rating":
			orderBy = fmt.Sprintf("r.overall_rating %s", order)
		case "helpful_votes":
			orderBy = fmt.Sprintf("r.helpful_votes %s", order)
		}
	}
	query += " ORDER BY " + orderBy

	// Pagination
	limit := 50
	if filters.Limit > 0 {
		limit = filters.Limit
	}
	offset := 0
	if filters.Offset > 0 {
		offset = filters.Offset
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	// Execute query
	var reviews []ReviewWithRelations
	err := r.db.SelectContext(ctx, &reviews, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews with relations: %w", err)
	}

	return reviews, nil
}
// ========================================================================
// READ OPERATIONS - Specific Queries
// ========================================================================

// FindByCustomerID retrieves all reviews by a customer
func (r *ReviewRepository) FindByCustomerID(ctx context.Context, customerID int, filters ReviewFilters) ([]models.Review, error) {
	filters.CustomerID = customerID
	return r.FindAll(ctx, filters)
}

// FindByBarberID retrieves all reviews for a barber
func (r *ReviewRepository) FindByBarberID(ctx context.Context, barberID int, filters ReviewFilters) ([]models.Review, error) {
	filters.BarberID = barberID
	return r.FindAll(ctx, filters)
}

// GetPublishedReviews retrieves only published and approved reviews
func (r *ReviewRepository) GetPublishedReviews(ctx context.Context, barberID int, filters ReviewFilters) ([]models.Review, error) {
	isPublished := true
	filters.BarberID = barberID
	filters.IsPublished = &isPublished
	filters.ModerationStatus = config.ReviewModerationApproved
	return r.FindAll(ctx, filters)
}

// ========================================================================
// UPDATE OPERATIONS
// ========================================================================

// Update updates a review
func (r *ReviewRepository) Update(ctx context.Context, review *models.Review) error {
	SetUpdateTimestamp(&review.UpdatedAt)

	query := `
		UPDATE reviews SET
			overall_rating = :overall_rating,
			service_quality_rating = :service_quality_rating,
			punctuality_rating = :punctuality_rating,
			cleanliness_rating = :cleanliness_rating,
			value_for_money_rating = :value_for_money_rating,
			professionalism_rating = :professionalism_rating,
			title = :title,
			comment = :comment,
			pros = :pros,
			cons = :cons,
			would_recommend = :would_recommend,
			would_book_again = :would_book_again,
			service_as_expected = :service_as_expected,
			duration_accurate = :duration_accurate,
			images = :images,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, review)
	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}

	return CheckRowsAffected(result, ErrReviewNotFound)
}

// UpdateModerationStatus updates the moderation status of a review
func (r *ReviewRepository) UpdateModerationStatus(ctx context.Context, id int, status string, moderatorID int, notes *string) error {
	if !IsValidModerationStatus(status) {
		return ErrInvalidModeration
	}

	now := time.Now()
	query := `
		UPDATE reviews SET
			moderation_status = $1,
			moderated_by = $2,
			moderation_notes = $3,
			moderated_at = $4,
			is_published = $5,
			updated_at = $6
		WHERE id = $7
	`

	// Auto-publish if approved
	isPublished := status == "approved"

	result, err := r.db.ExecContext(ctx, query, status, moderatorID, notes, now, isPublished, now, id)
	if err != nil {
		return fmt.Errorf("failed to update moderation status: %w", err)
	}

	return CheckRowsAffected(result, ErrReviewNotFound)
}

// AddBarberResponse adds a barber's response to a review
func (r *ReviewRepository) AddBarberResponse(ctx context.Context, id int, response string) error {
	now := time.Now()
	query := `
		UPDATE reviews SET
			barber_response = $1,
			barber_response_at = $2,
			updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, response, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to add barber response: %w", err)
	}

	return CheckRowsAffected(result, ErrReviewNotFound)
}

// IncrementHelpfulVotes increments the helpful votes counter
func (r *ReviewRepository) IncrementHelpfulVotes(ctx context.Context, id int, isHelpful bool) error {
	var query string
	if isHelpful {
		query = `UPDATE reviews SET helpful_votes = helpful_votes + 1, total_votes = total_votes + 1 WHERE id = $1`
	} else {
		query = `UPDATE reviews SET total_votes = total_votes + 1 WHERE id = $1`
	}

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment votes: %w", err)
	}

	return CheckRowsAffected(result, ErrReviewNotFound)
}

// ========================================================================
// DELETE OPERATIONS
// ========================================================================

// Delete removes a review (soft delete by setting moderation status)
func (r *ReviewRepository) Delete(ctx context.Context, id int) error {
	query := `
		UPDATE reviews SET
			is_published = false,
			moderation_status = 'rejected',
			updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	return CheckRowsAffected(result, ErrReviewNotFound)
}

// HardDelete permanently removes a review
func (r *ReviewRepository) HardDelete(ctx context.Context, id int) error {
	query := `DELETE FROM reviews WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete review: %w", err)
	}

	return CheckRowsAffected(result, ErrReviewNotFound)
}

// ========================================================================
// STATISTICS
// ========================================================================

// GetBarberStats retrieves review statistics for a barber
func (r *ReviewRepository) GetBarberStats(ctx context.Context, barberID int) (*ReviewStats, error) {
	query := `
		SELECT
			COUNT(*) as total_reviews,
			COALESCE(AVG(overall_rating), 0) as average_rating,
			COUNT(CASE WHEN overall_rating = 5 THEN 1 END) as five_star_count,
			COUNT(CASE WHEN overall_rating = 4 THEN 1 END) as four_star_count,
			COUNT(CASE WHEN overall_rating = 3 THEN 1 END) as three_star_count,
			COUNT(CASE WHEN overall_rating = 2 THEN 1 END) as two_star_count,
			COUNT(CASE WHEN overall_rating = 1 THEN 1 END) as one_star_count,
			COALESCE(AVG(CASE WHEN would_recommend = true THEN 100.0 ELSE 0.0 END), 0) as recommend_percent
		FROM reviews
		WHERE barber_id = $1
		AND is_published = true
		AND moderation_status = 'approved'
	`

	var stats ReviewStats
	err := r.db.GetContext(ctx, &stats, query, barberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get barber review stats: %w", err)
	}

	return &stats, nil
}

// Count returns the total number of reviews matching the filters
func (r *ReviewRepository) Count(ctx context.Context, filters ReviewFilters) (int, error) {
	query := `SELECT COUNT(*) FROM reviews WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if filters.CustomerID > 0 {
		query += fmt.Sprintf(" AND customer_id = $%d", argCount)
		args = append(args, filters.CustomerID)
		argCount++
	}

	if filters.BarberID > 0 {
		query += fmt.Sprintf(" AND barber_id = $%d", argCount)
		args = append(args, filters.BarberID)
		argCount++
	}

	if filters.ModerationStatus != "" {
		query += fmt.Sprintf(" AND moderation_status = $%d", argCount)
		args = append(args, filters.ModerationStatus)
		argCount++
	}

	if filters.IsPublished != nil {
		query += fmt.Sprintf(" AND is_published = $%d", argCount)
		args = append(args, *filters.IsPublished)
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count reviews: %w", err)
	}

	return count, nil
}

// ========================================================================
// TRANSACTION SUPPORT
// ========================================================================

// CreateTx inserts a new review within a transaction
func (r *ReviewRepository) CreateTx(ctx context.Context, tx *sqlx.Tx, review *models.Review) error {
	// Check for duplicate review within transaction
	var exists bool
	err := tx.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM reviews WHERE booking_id = $1)`, review.BookingID)
	if err != nil {
		return fmt.Errorf("failed to check existing review: %w", err)
	}
	if exists {
		return ErrDuplicateReview
	}

	query := `
		INSERT INTO reviews (
			booking_id, customer_id, barber_id,
			overall_rating, service_quality_rating, punctuality_rating,
			cleanliness_rating, value_for_money_rating, professionalism_rating,
			title, comment, pros, cons,
			would_recommend, would_book_again, service_as_expected, duration_accurate,
			images,
			is_verified, is_published, moderation_status,
			created_at, updated_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6,
			$7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16, $17,
			$18,
			$19, $20, $21,
			$22, $23
		) RETURNING id
	`

	// Set timestamps and defaults
	SetCreateTimestamps(&review.CreatedAt, &review.UpdatedAt)
	SetDefaultString(&review.ModerationStatus, config.ReviewModerationPending)

	err = tx.QueryRowContext(ctx, query,
		review.BookingID, review.CustomerID, review.BarberID,
		review.OverallRating, review.ServiceQualityRating, review.PunctualityRating,
		review.CleanlinessRating, review.ValueForMoneyRating, review.ProfessionalismRating,
		review.Title, review.Comment, review.Pros, review.Cons,
		review.WouldRecommend, review.WouldBookAgain, review.ServiceAsExpected, review.DurationAccurate,
		review.Images,
		review.IsVerified, review.IsPublished, review.ModerationStatus,
		review.CreatedAt, review.UpdatedAt,
	).Scan(&review.ID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return ErrDuplicateReview
		}
		return fmt.Errorf("failed to create review: %w", err)
	}

	return nil
}

// BeginTx starts a new database transaction
func (r *ReviewRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}
