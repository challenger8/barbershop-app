// internal/repository/booking_repository.go
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
// BOOKING REPOSITORY - Data Access Layer for Bookings
// ========================================================================

// BookingRepository handles booking data operations
type BookingRepository struct {
	db *sqlx.DB
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(db *sqlx.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// ========================================================================
// FILTER STRUCTS
// ========================================================================

// BookingFilters represents filter options for booking queries
type BookingFilters struct {
	CustomerID    int       `form:"customer_id"`
	BarberID      int       `form:"barber_id"`
	Status        string    `form:"status"`
	Statuses      []string  `form:"statuses"`
	PaymentStatus string    `form:"payment_status"`
	BookingSource string    `form:"booking_source"`
	StartDateFrom time.Time `form:"start_date_from" time_format:"2006-01-02"`
	StartDateTo   time.Time `form:"start_date_to" time_format:"2006-01-02"`
	CreatedFrom   time.Time `form:"created_from" time_format:"2006-01-02T15:04:05Z07:00"`
	CreatedTo     time.Time `form:"created_to" time_format:"2006-01-02T15:04:05Z07:00"`
	Search        string    `form:"search"`
	SortBy        string    `form:"sort_by"`
	Order         string    `form:"order"`
	Limit         int       `form:"limit,default=50"`
	Offset        int       `form:"offset,default=0"`

	IncludeCustomer bool `form:"include_customer"`
	IncludeBarber   bool `form:"include_barber"`
}

// BookingHistoryFilters for audit trail queries
type BookingHistoryFilters struct {
	BookingID  int
	ChangeType string
	ChangedBy  int
	FromDate   time.Time
	ToDate     time.Time
	Limit      int
	Offset     int
}

// ========================================================================
// CUSTOM ERRORS
// ========================================================================

var (
	ErrInvalidStatusChange     = fmt.Errorf("invalid status transition")
	ErrBookingAlreadyCancelled = fmt.Errorf("booking is already cancelled")
	ErrCancellationNotAllowed  = fmt.Errorf("booking cannot be cancelled")
)

// ========================================================================
// VALID STATUS VALUES
// ========================================================================

// Valid booking statuses - using config constants
var ValidBookingStatuses = []string{
	config.BookingStatusPending,
	config.BookingStatusConfirmed,
	config.BookingStatusInProgress,
	config.BookingStatusCompleted,
	config.BookingStatusCancelledByCustomer,
	config.BookingStatusCancelledByBarber,
	config.BookingStatusNoShow,
}

// Valid payment statuses - using config constants
var ValidPaymentStatuses = []string{
	config.PaymentStatusPending,
	config.PaymentStatusPaid,
	config.PaymentStatusPartiallyPaid,
	config.PaymentStatusRefunded,
	config.PaymentStatusFailed,
}

// ValidStatusTransitions defines allowed status changes
// Key: current status, Value: allowed next statuses
var ValidStatusTransitions = map[string][]string{
	config.BookingStatusPending:             {config.BookingStatusConfirmed, config.BookingStatusCancelledByCustomer, config.BookingStatusCancelledByBarber},
	config.BookingStatusConfirmed:           {config.BookingStatusInProgress, config.BookingStatusCancelledByCustomer, config.BookingStatusCancelledByBarber, config.BookingStatusNoShow},
	config.BookingStatusInProgress:          {config.BookingStatusCompleted},
	config.BookingStatusCompleted:           {}, // Terminal state
	config.BookingStatusCancelledByCustomer: {}, // Terminal state
	config.BookingStatusCancelledByBarber:   {}, // Terminal state
	config.BookingStatusNoShow:              {}, // Terminal state
}

// IsValidStatusTransition checks if a status change is allowed
func IsValidStatusTransition(currentStatus, newStatus string) bool {
	allowedStatuses, exists := ValidStatusTransitions[currentStatus]
	if !exists {
		return false
	}
	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}
	return false
}

// ========================================================================
// CREATE OPERATIONS
// ========================================================================

// Create inserts a new booking into the database
func (r *BookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	query := `
		INSERT INTO bookings (
			uuid, booking_number, customer_id, barber_id, time_slot_id,
			service_name, service_category, estimated_duration_minutes,
			customer_name, customer_email, customer_phone,
			status, service_price, total_price, discount_amount, tax_amount, tip_amount, currency,
			payment_status, payment_method, payment_reference,
			notes, special_requests, internal_notes,
			scheduled_start_time, scheduled_end_time,
			booking_source, referral_source, utm_campaign,
			created_at, updated_at
		) VALUES (
			:uuid, :booking_number, :customer_id, :barber_id, :time_slot_id,
			:service_name, :service_category, :estimated_duration_minutes,
			:customer_name, :customer_email, :customer_phone,
			:status, :service_price, :total_price, :discount_amount, :tax_amount, :tip_amount, :currency,
			:payment_status, :payment_method, :payment_reference,
			:notes, :special_requests, :internal_notes,
			:scheduled_start_time, :scheduled_end_time,
			:booking_source, :referral_source, :utm_campaign,
			:created_at, :updated_at
		) RETURNING id
	`

	// Set timestamps using helper
	SetCreateTimestamps(&booking.CreatedAt, &booking.UpdatedAt)

	// Set defaults using helpers
	SetDefaultString(&booking.Status, config.BookingStatusPending)
	SetDefaultString(&booking.PaymentStatus, config.PaymentStatusPending)
	SetDefaultString(&booking.Currency, config.DefaultCurrency)
	SetDefaultString(&booking.BookingSource, "web_app")

	rows, err := r.db.NamedQueryContext(ctx, query, booking)
	if err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&booking.ID); err != nil {
			return fmt.Errorf("failed to scan booking id: %w", err)
		}
	}

	return nil
}

// ========================================================================
// READ OPERATIONS - FindByID
// ========================================================================

// FindByID retrieves a booking by its ID
func (r *BookingRepository) FindByID(ctx context.Context, id int) (*models.Booking, error) {
	query := `
		SELECT * FROM bookings
		WHERE id = $1
	`

	var booking models.Booking
	err := r.db.GetContext(ctx, &booking, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to find booking by id: %w", err)
	}

	return &booking, nil
}

// FindByUUID retrieves a booking by its UUID
func (r *BookingRepository) FindByUUID(ctx context.Context, uuid string) (*models.Booking, error) {
	query := `
		SELECT * FROM bookings
		WHERE uuid = $1
	`

	var booking models.Booking
	err := r.db.GetContext(ctx, &booking, query, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to find booking by uuid: %w", err)
	}

	return &booking, nil
}

// FindByBookingNumber retrieves a booking by its human-readable booking number
func (r *BookingRepository) FindByBookingNumber(ctx context.Context, bookingNumber string) (*models.Booking, error) {
	query := `
		SELECT * FROM bookings
		WHERE booking_number = $1
	`

	var booking models.Booking
	err := r.db.GetContext(ctx, &booking, query, bookingNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to find booking by booking number: %w", err)
	}

	return &booking, nil
}

// ========================================================================
// READ OPERATIONS - FindAll with Filters
// ========================================================================

// FindAll retrieves bookings with optional filters
func (r *BookingRepository) FindAll(ctx context.Context, filters BookingFilters) ([]models.Booking, error) {
	// Base query
	query := `SELECT * FROM bookings WHERE 1=1`
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

	// Single status filter
	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	// Multiple statuses filter (OR condition)
	if len(filters.Statuses) > 0 {
		placeholders := make([]string, len(filters.Statuses))
		for i, status := range filters.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, status)
			argCount++
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ", "))
	}

	// Payment status filter
	if filters.PaymentStatus != "" {
		query += fmt.Sprintf(" AND payment_status = $%d", argCount)
		args = append(args, filters.PaymentStatus)
		argCount++
	}

	// Date range - scheduled start time
	if !filters.StartDateFrom.IsZero() {
		query += fmt.Sprintf(" AND scheduled_start_time >= $%d", argCount)
		args = append(args, filters.StartDateFrom)
		argCount++
	}
	if !filters.StartDateTo.IsZero() {
		query += fmt.Sprintf(" AND scheduled_start_time <= $%d", argCount)
		args = append(args, filters.StartDateTo)
		argCount++
	}

	// Date range - created at
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
		query += fmt.Sprintf(" AND (booking_number ILIKE $%d OR customer_name ILIKE $%d OR customer_email ILIKE $%d)",
			argCount, argCount+1, argCount+2)
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
		argCount += 3
	}

	// Booking source filter
	if filters.BookingSource != "" {
		query += fmt.Sprintf(" AND booking_source = $%d", argCount)
		args = append(args, filters.BookingSource)
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
		case "scheduled_start_time":
			orderBy = fmt.Sprintf("scheduled_start_time %s", order)
		case "total_price":
			orderBy = fmt.Sprintf("total_price %s", order)
		case "created_at":
			orderBy = fmt.Sprintf("created_at %s", order)
		case "status":
			orderBy = fmt.Sprintf("status %s", order)
		}
	}
	query += " ORDER BY " + orderBy

	// Pagination
	limit := 50 // Default limit
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
	var bookings []models.Booking
	err := r.db.SelectContext(ctx, &bookings, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find bookings: %w", err)
	}

	return bookings, nil
}

// ========================================================================
// READ OPERATIONS - FindAll with Relations (prevents N+1 queries)
// ========================================================================

// BookingWithRelations represents a booking with optional loaded relations
type BookingWithRelations struct {
	models.Booking
	CustomerName  *string `db:"customer_user_name"`
	CustomerEmail *string `db:"customer_user_email"`
	BarberName    *string `db:"barber_shop_name"`
	BarberCity    *string `db:"barber_city"`
	BarberPhone   *string `db:"barber_phone"`
}

// FindAllWithRelations retrieves bookings with optional relation loading
// Use this instead of FindAll when you need customer/barber info
func (r *BookingRepository) FindAllWithRelations(ctx context.Context, filters BookingFilters) ([]BookingWithRelations, error) {
	// Build SELECT columns
	selectCols := `bk.*`
	joins := ""

	if filters.IncludeCustomer {
		selectCols += `, u.name as customer_user_name, u.email as customer_user_email`
		joins += ` LEFT JOIN users u ON bk.customer_id = u.id`
	}

	if filters.IncludeBarber {
		selectCols += `, b.shop_name as barber_shop_name, b.city as barber_city, b.phone as barber_phone`
		joins += ` LEFT JOIN barbers b ON bk.barber_id = b.id`
	}

	// Base query
	query := fmt.Sprintf(`SELECT %s FROM bookings bk%s WHERE 1=1`, selectCols, joins)
	args := []interface{}{}
	argCount := 1

	// Apply filters (same as FindAll)
	if filters.CustomerID > 0 {
		query += fmt.Sprintf(" AND bk.customer_id = $%d", argCount)
		args = append(args, filters.CustomerID)
		argCount++
	}

	if filters.BarberID > 0 {
		query += fmt.Sprintf(" AND bk.barber_id = $%d", argCount)
		args = append(args, filters.BarberID)
		argCount++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND bk.status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if len(filters.Statuses) > 0 {
		placeholders := make([]string, len(filters.Statuses))
		for i, status := range filters.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, status)
			argCount++
		}
		query += fmt.Sprintf(" AND bk.status IN (%s)", strings.Join(placeholders, ", "))
	}

	if filters.PaymentStatus != "" {
		query += fmt.Sprintf(" AND bk.payment_status = $%d", argCount)
		args = append(args, filters.PaymentStatus)
		argCount++
	}

	if !filters.StartDateFrom.IsZero() {
		query += fmt.Sprintf(" AND bk.scheduled_start_time >= $%d", argCount)
		args = append(args, filters.StartDateFrom)
		argCount++
	}

	if !filters.StartDateTo.IsZero() {
		query += fmt.Sprintf(" AND bk.scheduled_start_time <= $%d", argCount)
		args = append(args, filters.StartDateTo)
		argCount++
	}

	if !filters.CreatedFrom.IsZero() {
		query += fmt.Sprintf(" AND bk.created_at >= $%d", argCount)
		args = append(args, filters.CreatedFrom)
		argCount++
	}

	if !filters.CreatedTo.IsZero() {
		query += fmt.Sprintf(" AND bk.created_at <= $%d", argCount)
		args = append(args, filters.CreatedTo)
		argCount++
	}

	if filters.Search != "" {
		query += fmt.Sprintf(" AND (bk.booking_number ILIKE $%d OR bk.customer_name ILIKE $%d OR bk.customer_email ILIKE $%d)",
			argCount, argCount+1, argCount+2)
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
		argCount += 3
	}

	if filters.BookingSource != "" {
		query += fmt.Sprintf(" AND bk.booking_source = $%d", argCount)
		args = append(args, filters.BookingSource)
		argCount++
	}

	// Sorting
	orderBy := "bk.created_at DESC"
	if filters.SortBy != "" {
		order := "DESC"
		if filters.Order == "ASC" || filters.Order == "asc" {
			order = "ASC"
		}
		switch filters.SortBy {
		case "scheduled_start_time":
			orderBy = fmt.Sprintf("bk.scheduled_start_time %s", order)
		case "total_price":
			orderBy = fmt.Sprintf("bk.total_price %s", order)
		case "created_at":
			orderBy = fmt.Sprintf("bk.created_at %s", order)
		case "status":
			orderBy = fmt.Sprintf("bk.status %s", order)
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
	var bookings []BookingWithRelations
	err := r.db.SelectContext(ctx, &bookings, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find bookings with relations: %w", err)
	}

	return bookings, nil
}

// ========================================================================
// READ OPERATIONS - Specific Queries
// ========================================================================

// FindByCustomerID retrieves all bookings for a customer
func (r *BookingRepository) FindByCustomerID(ctx context.Context, customerID int, filters BookingFilters) ([]models.Booking, error) {
	filters.CustomerID = customerID
	return r.FindAll(ctx, filters)
}

// FindByBarberID retrieves all bookings for a barber
func (r *BookingRepository) FindByBarberID(ctx context.Context, barberID int, filters BookingFilters) ([]models.Booking, error) {
	filters.BarberID = barberID
	return r.FindAll(ctx, filters)
}

// GetUpcomingBookings retrieves upcoming bookings for a barber or customer
func (r *BookingRepository) GetUpcomingBookings(ctx context.Context, filters BookingFilters) ([]models.Booking, error) {
	filters.StartDateFrom = time.Now()
	filters.Statuses = []string{config.BookingStatusPending, config.BookingStatusConfirmed}
	filters.SortBy = "scheduled_start_time"
	filters.Order = "ASC"
	return r.FindAll(ctx, filters)
}

// GetTodayBookings retrieves today's bookings for a barber
func (r *BookingRepository) GetTodayBookings(ctx context.Context, barberID int) ([]models.Booking, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	filters := BookingFilters{
		BarberID:      barberID,
		StartDateFrom: startOfDay,
		StartDateTo:   endOfDay,
		Statuses:      []string{config.BookingStatusPending, config.BookingStatusConfirmed, config.BookingStatusInProgress},
		SortBy:        "scheduled_start_time",
		Order:         "ASC",
	}

	return r.FindAll(ctx, filters)
}

// ========================================================================
// UPDATE OPERATIONS
// ========================================================================

// Update updates a booking's basic information
func (r *BookingRepository) Update(ctx context.Context, booking *models.Booking) error {
	SetUpdateTimestamp(&booking.UpdatedAt)

	query := `
		UPDATE bookings SET
			service_name = :service_name,
			service_category = :service_category,
			estimated_duration_minutes = :estimated_duration_minutes,
			customer_name = :customer_name,
			customer_email = :customer_email,
			customer_phone = :customer_phone,
			notes = :notes,
			special_requests = :special_requests,
			internal_notes = :internal_notes,
			scheduled_start_time = :scheduled_start_time,
			scheduled_end_time = :scheduled_end_time,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, booking)
	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	return CheckRowsAffected(result, ErrBookingNotFound)
}

// UpdateStatus updates only the status of a booking
func (r *BookingRepository) UpdateStatus(ctx context.Context, id int, newStatus string) error {
	// ─────────────────────────────────────────────────────────────────
	// TODO: YOUR TASK #2 - Implement status update
	// ─────────────────────────────────────────────────────────────────
	// 1. First, get the current booking to check current status
	// 2. Validate the status transition using IsValidStatusTransition()
	// 3. Update the status and updated_at timestamp
	// 4. If status is "in_progress", also set actual_start_time
	// 5. If status is "completed", also set actual_end_time
	// ─────────────────────────────────────────────────────────────────

	// Get current booking
	booking, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate transition
	if !IsValidStatusTransition(booking.Status, newStatus) {
		return fmt.Errorf("%w: cannot change from '%s' to '%s'", ErrInvalidStatusChange, booking.Status, newStatus)
	}

	// Build update query
	now := time.Now()
	query := `UPDATE bookings SET status = $1, updated_at = $2`
	args := []interface{}{newStatus, now}
	argCount := 3

	// Handle special status updates
	switch {
	case newStatus == config.BookingStatusInProgress:
		query += fmt.Sprintf(", actual_start_time = $%d", argCount)
		args = append(args, now)
		argCount++
	case newStatus == config.BookingStatusCompleted:
		query += fmt.Sprintf(", actual_end_time = $%d", argCount)
		args = append(args, now)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}

	return CheckRowsAffected(result, ErrBookingNotFound)
}

// ========================================================================
// CANCEL OPERATIONS
// ========================================================================

// Cancel cancels a booking with a reason
func (r *BookingRepository) Cancel(ctx context.Context, id int, cancelledBy int, reason string, isByCustomer bool) error {
	// Get current booking
	booking, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if booking can be cancelled
	if !booking.CanBeCancelled() {
		return ErrCancellationNotAllowed
	}

	// Determine cancellation status
	status := "cancelled_by_barber"
	if isByCustomer {
		status = "cancelled_by_customer"
	}

	now := time.Now()
	query := `
		UPDATE bookings SET
			status = $1,
			cancelled_at = $2,
			cancelled_by = $3,
			cancellation_reason = $4,
			updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query, status, now, cancelledBy, reason, now, id)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	return CheckRowsAffected(result, ErrBookingNotFound)
}

// ========================================================================
// CONFLICT CHECKING
// ========================================================================

// CheckConflict checks if there's a conflicting booking for a barber at the given time
func (r *BookingRepository) CheckConflict(ctx context.Context, barberID int, startTime, endTime time.Time, excludeBookingID int) (bool, error) {
	query := `
		SELECT COUNT(*) FROM bookings
		WHERE barber_id = $1
		AND status NOT IN ('cancelled_by_customer', 'cancelled_by_barber', 'no_show', 'completed')
		AND id != $2
		AND scheduled_start_time < $3
		AND scheduled_end_time > $4
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, barberID, excludeBookingID, endTime, startTime)
	if err != nil {
		return false, fmt.Errorf("failed to check booking conflict: %w", err)
	}
	return count > 0, nil
}

// ========================================================================
// TRANSACTION SUPPORT
// ========================================================================

// BeginTx starts a new database transaction
func (r *BookingRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// CheckConflictForUpdate checks for booking conflicts with row locking (FOR UPDATE)
// This prevents race conditions by locking conflicting rows until transaction commits
func (r *BookingRepository) CheckConflictForUpdate(ctx context.Context, tx *sqlx.Tx, barberID int, startTime, endTime time.Time, excludeBookingID int) (bool, error) {
	query := `
		SELECT COUNT(*) FROM bookings
		WHERE barber_id = $1
		AND status NOT IN ('cancelled_by_customer', 'cancelled_by_barber', 'no_show', 'completed')
		AND id != $2
		AND scheduled_start_time < $3
		AND scheduled_end_time > $4
		FOR UPDATE
	`
	var count int
	err := tx.GetContext(ctx, &count, query, barberID, excludeBookingID, endTime, startTime)
	if err != nil {
		return false, fmt.Errorf("failed to check booking conflict: %w", err)
	}

	return count > 0, nil
}

// CreateTx inserts a new booking within a transaction
func (r *BookingRepository) CreateTx(ctx context.Context, tx *sqlx.Tx, booking *models.Booking) error {
	query := `
		INSERT INTO bookings (
			uuid, booking_number, customer_id, barber_id, time_slot_id,
			service_name, service_category, estimated_duration_minutes,
			customer_name, customer_email, customer_phone,
			status, service_price, total_price, discount_amount, tax_amount, tip_amount, currency,
			payment_status, payment_method, payment_reference,
			notes, special_requests, internal_notes,
			scheduled_start_time, scheduled_end_time,
			booking_source, referral_source, utm_campaign,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21,
			$22, $23, $24,
			$25, $26,
			$27, $28, $29,
			$30, $31
		) RETURNING id
	`

	// Set timestamps using helper
	SetCreateTimestamps(&booking.CreatedAt, &booking.UpdatedAt)

	// Set defaults using helpers
	SetDefaultString(&booking.Status, config.BookingStatusPending)
	SetDefaultString(&booking.PaymentStatus, config.PaymentStatusPending)
	SetDefaultString(&booking.Currency, config.DefaultCurrency)
	SetDefaultString(&booking.BookingSource, "web_app")

	err := tx.QueryRowContext(ctx, query,
		booking.UUID, booking.BookingNumber, booking.CustomerID, booking.BarberID, booking.TimeSlotID,
		booking.ServiceName, booking.ServiceCategory, booking.EstimatedDurationMinutes,
		booking.CustomerName, booking.CustomerEmail, booking.CustomerPhone,
		booking.Status, booking.ServicePrice, booking.TotalPrice, booking.DiscountAmount, booking.TaxAmount, booking.TipAmount, booking.Currency,
		booking.PaymentStatus, booking.PaymentMethod, booking.PaymentReference,
		booking.Notes, booking.SpecialRequests, booking.InternalNotes,
		booking.ScheduledStartTime, booking.ScheduledEndTime,
		booking.BookingSource, booking.ReferralSource, booking.UTMCampaign,
		booking.CreatedAt, booking.UpdatedAt,
	).Scan(&booking.ID)

	if err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}

	return nil
}

// CreateHistoryTx creates a booking history record within a transaction
func (r *BookingRepository) CreateHistoryTx(ctx context.Context, tx *sqlx.Tx, history *models.BookingHistory) error {
	query := `
		INSERT INTO booking_history (
			booking_id, changed_by, change_type,
			old_values, new_values, change_reason,
			ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now()
	}

	err := tx.QueryRowContext(ctx, query,
		history.BookingID, history.ChangedBy, history.ChangeType,
		history.OldValues, history.NewValues, history.ChangeReason,
		history.IPAddress, history.UserAgent, history.CreatedAt,
	).Scan(&history.ID)

	if err != nil {
		return fmt.Errorf("failed to create booking history: %w", err)
	}

	return nil
}

// ========================================================================
// PAYMENT OPERATIONS
// ========================================================================

// UpdatePaymentStatus updates the payment status of a booking
func (r *BookingRepository) UpdatePaymentStatus(ctx context.Context, id int, paymentStatus string, paymentMethod *string, paymentReference *string) error {
	now := time.Now()

	query := `
		UPDATE bookings SET
			payment_status = $1,
			payment_method = $2,
			payment_reference = $3,
			updated_at = $4
	`
	args := []interface{}{paymentStatus, paymentMethod, paymentReference, now}

	// If paid, set paid_at timestamp
	if paymentStatus == "paid" {
		query += ", paid_at = $5 WHERE id = $6"
		args = append(args, now, id)
	} else {
		query += " WHERE id = $5"
		args = append(args, id)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return CheckRowsAffected(result, ErrBookingNotFound)
}

// ========================================================================
// STATISTICS
// ========================================================================

// BookingStats represents booking statistics
type BookingStats struct {
	TotalBookings     int     `json:"total_bookings" db:"total_bookings"`
	CompletedBookings int     `json:"completed_bookings" db:"completed_bookings"`
	CancelledBookings int     `json:"cancelled_bookings" db:"cancelled_bookings"`
	NoShowBookings    int     `json:"no_show_bookings" db:"no_show_bookings"`
	TotalRevenue      float64 `json:"total_revenue" db:"total_revenue"`
	AveragePrice      float64 `json:"average_price" db:"average_price"`
}

// GetBarberStats retrieves booking statistics for a barber
func (r *BookingRepository) GetBarberStats(ctx context.Context, barberID int, from, to time.Time) (*BookingStats, error) {
	query := `
		SELECT
			COUNT(*) as total_bookings,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_bookings,
			COUNT(CASE WHEN status IN ('cancelled_by_customer', 'cancelled_by_barber') THEN 1 END) as cancelled_bookings,
			COUNT(CASE WHEN status = 'no_show' THEN 1 END) as no_show_bookings,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN total_price ELSE 0 END), 0) as total_revenue,
			COALESCE(AVG(CASE WHEN status = 'completed' THEN total_price END), 0) as average_price
		FROM bookings
		WHERE barber_id = $1
		AND created_at >= $2
		AND created_at <= $3
	`

	var stats BookingStats
	err := r.db.GetContext(ctx, &stats, query, barberID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get barber stats: %w", err)
	}

	return &stats, nil
}

// ========================================================================
// BOOKING HISTORY (Audit Trail)
// ========================================================================

// CreateHistory creates a booking history record
func (r *BookingRepository) CreateHistory(ctx context.Context, history *models.BookingHistory) error {
	query := `
		INSERT INTO booking_history (
			booking_id, changed_by, change_type,
			old_values, new_values, change_reason,
			ip_address, user_agent, created_at
		) VALUES (
			:booking_id, :changed_by, :change_type,
			:old_values, :new_values, :change_reason,
			:ip_address, :user_agent, :created_at
		) RETURNING id
	`

	history.CreatedAt = time.Now()

	rows, err := r.db.NamedQueryContext(ctx, query, history)
	if err != nil {
		return fmt.Errorf("failed to create booking history: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&history.ID); err != nil {
			return fmt.Errorf("failed to scan history id: %w", err)
		}
	}

	return nil
}

// GetHistory retrieves booking history for a booking
func (r *BookingRepository) GetHistory(ctx context.Context, bookingID int) ([]models.BookingHistory, error) {
	query := `
		SELECT * FROM booking_history
		WHERE booking_id = $1
		ORDER BY created_at DESC
	`

	var history []models.BookingHistory
	err := r.db.SelectContext(ctx, &history, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking history: %w", err)
	}

	return history, nil
}

// ========================================================================
// COUNT OPERATIONS
// ========================================================================

// Count returns the total number of bookings matching the filters
func (r *BookingRepository) Count(ctx context.Context, filters BookingFilters) (int, error) {
	query := `SELECT COUNT(*) FROM bookings WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Apply same filters as FindAll (simplified version)
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

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count bookings: %w", err)
	}

	return count, nil
}
