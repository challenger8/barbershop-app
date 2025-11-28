// internal/services/booking_service.go
package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/models"
	"barber-booking-system/internal/repository"

	"github.com/google/uuid"
)

// ========================================================================
// BOOKING SERVICE - Business Logic Layer
// ========================================================================

// BookingService handles booking business logic
type BookingService struct {
	repo        *repository.BookingRepository
	barberRepo  *repository.BarberRepository
	serviceRepo *repository.ServiceRepository
	cache       *cache.CacheService
}

// NewBookingService creates a new booking service
func NewBookingService(
	repo *repository.BookingRepository,
	barberRepo *repository.BarberRepository,
	serviceRepo *repository.ServiceRepository,
	cache *cache.CacheService,
) *BookingService {
	return &BookingService{
		repo:        repo,
		barberRepo:  barberRepo,
		serviceRepo: serviceRepo,
		cache:       cache,
	}
}

// ========================================================================
// REQUEST/RESPONSE STRUCTS
// ========================================================================

// CreateBookingRequest represents a request to create a booking
type CreateBookingRequest struct {
	// Required fields
	BarberID        int       `json:"barber_id" binding:"required"`
	ServiceID       int       `json:"service_id" binding:"required"`
	StartTime       time.Time `json:"start_time" binding:"required"`
	DurationMinutes int       `json:"duration_minutes" binding:"required,min=15,max=480"`

	// Customer info (either customer_id OR guest info)
	CustomerID    *int    `json:"customer_id"`
	CustomerName  *string `json:"customer_name"`
	CustomerEmail *string `json:"customer_email"`
	CustomerPhone *string `json:"customer_phone"`

	// Optional fields
	Notes           *string `json:"notes"`
	SpecialRequests *string `json:"special_requests"`
	BookingSource   string  `json:"booking_source"` // mobile_app, web_app, phone, walk_in

	// Pricing (optional - will be calculated if not provided)
	ServicePrice   *float64 `json:"service_price"`
	DiscountAmount *float64 `json:"discount_amount"`
}

// UpdateBookingRequest represents a request to update a booking
type UpdateBookingRequest struct {
	CustomerName    *string `json:"customer_name"`
	CustomerEmail   *string `json:"customer_email"`
	CustomerPhone   *string `json:"customer_phone"`
	Notes           *string `json:"notes"`
	SpecialRequests *string `json:"special_requests"`
	InternalNotes   *string `json:"internal_notes"`
}

// RescheduleBookingRequest represents a request to reschedule
type RescheduleBookingRequest struct {
	NewStartTime    time.Time `json:"new_start_time" binding:"required"`
	DurationMinutes int       `json:"duration_minutes"`
	Reason          *string   `json:"reason"`
}

// CancelBookingRequest represents a request to cancel
type CancelBookingRequest struct {
	Reason       string `json:"reason"`
	IsByCustomer bool   `json:"is_by_customer"`
}

// UpdateStatusRequest represents a status update request
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// BookingResponse wraps booking with additional computed fields
type BookingResponse struct {
	*models.Booking
	CanCancel     bool   `json:"can_cancel"`
	CanReschedule bool   `json:"can_reschedule"`
	TimeUntil     string `json:"time_until,omitempty"`
}

// ========================================================================
// HELPER FUNCTIONS
// ========================================================================

// generateBookingNumber creates a unique human-readable booking number
// Format: BK + YYYYMMDD + 4 random digits (e.g., BK202411281234)
func (s *BookingService) generateBookingNumber() string {
	now := time.Now()
	dateStr := now.Format("20060102")
	randomNum := rand.Intn(10000)
	return fmt.Sprintf("BK%s%04d", dateStr, randomNum)
}

// calculateEndTime calculates the end time based on start time and duration
func (s *BookingService) calculateEndTime(startTime time.Time, durationMinutes int) time.Time {
	return startTime.Add(time.Duration(durationMinutes) * time.Minute)
}

// calculateTotalPrice calculates total price with tax
// ─────────────────────────────────────────────────────────────────────────
// TODO: YOUR TASK #1 - Implement this function
// ─────────────────────────────────────────────────────────────────────────
// Calculate: totalPrice = servicePrice - discountAmount + taxAmount
// Tax rate is 8% (0.08)
//
// Steps:
// 1. Calculate tax: taxAmount = (servicePrice - discountAmount) * 0.08
// 2. Calculate total: total = servicePrice - discountAmount + taxAmount
// 3. Return servicePrice, discountAmount, taxAmount, totalPrice
// ─────────────────────────────────────────────────────────────────────────
func (s *BookingService) calculateTotalPrice(servicePrice float64, discountAmount float64) (float64, float64, float64, float64) {
	// YOUR CODE HERE:
	// Hint: taxRate := 0.08

	// For now, returning placeholder - YOU IMPLEMENT THIS
	taxRate := 0.08
	taxableAmount := servicePrice - discountAmount
	taxAmount := taxableAmount * taxRate
	totalPrice := taxableAmount + taxAmount

	return servicePrice, discountAmount, taxAmount, totalPrice
}

// validateBookingTime checks if the booking time is valid
// ─────────────────────────────────────────────────────────────────────────
// TODO: YOUR TASK #2 - Implement this function
// ─────────────────────────────────────────────────────────────────────────
// Validation rules:
// 1. Booking must be in the future (startTime > now)
// 2. Booking must be at least 1 hour in advance
// 3. Booking must not be more than 30 days in advance
// 4. Duration must be between 15 and 480 minutes
//
// Return nil if valid, or error with descriptive message
// ─────────────────────────────────────────────────────────────────────────
func (s *BookingService) validateBookingTime(startTime time.Time, durationMinutes int) error {
	now := time.Now()

	// Rule 1: Must be in the future
	if startTime.Before(now) {
		return fmt.Errorf("booking time must be in the future")
	}

	// Rule 2: At least 1 hour in advance
	minAdvanceTime := now.Add(1 * time.Hour)
	if startTime.Before(minAdvanceTime) {
		return fmt.Errorf("booking must be at least 1 hour in advance")
	}

	// Rule 3: Not more than 30 days in advance
	maxAdvanceTime := now.Add(30 * 24 * time.Hour)
	if startTime.After(maxAdvanceTime) {
		return fmt.Errorf("booking cannot be more than 30 days in advance")
	}

	// Rule 4: Duration validation
	if durationMinutes < 15 {
		return fmt.Errorf("booking duration must be at least 15 minutes")
	}
	if durationMinutes > 480 {
		return fmt.Errorf("booking duration cannot exceed 8 hours (480 minutes)")
	}

	return nil
}

// toBookingResponse converts a booking to a response with computed fields
func (s *BookingService) toBookingResponse(booking *models.Booking) *BookingResponse {
	response := &BookingResponse{
		Booking:       booking,
		CanCancel:     booking.CanBeCancelled(),
		CanReschedule: booking.CanBeCancelled(), // Same rules as cancel
	}

	// Calculate time until booking
	if booking.IsUpcoming() {
		duration := time.Until(booking.ScheduledStartTime)
		if duration.Hours() >= 24 {
			days := int(duration.Hours() / 24)
			response.TimeUntil = fmt.Sprintf("%d days", days)
		} else if duration.Hours() >= 1 {
			response.TimeUntil = fmt.Sprintf("%.0f hours", duration.Hours())
		} else {
			response.TimeUntil = fmt.Sprintf("%.0f minutes", duration.Minutes())
		}
	}

	return response
}

// ========================================================================
// CREATE BOOKING
// ========================================================================

// CreateBooking creates a new booking with full validation
func (s *BookingService) CreateBooking(ctx context.Context, req CreateBookingRequest, createdByUserID *int) (*BookingResponse, error) {
	// ─────────────────────────────────────────────────────────────────
	// Step 1: Validate booking time
	// ─────────────────────────────────────────────────────────────────
	if err := s.validateBookingTime(req.StartTime, req.DurationMinutes); err != nil {
		return nil, err
	}

	// ─────────────────────────────────────────────────────────────────
	// Step 2: Validate barber exists and is active
	// ─────────────────────────────────────────────────────────────────
	barber, err := s.barberRepo.FindByID(ctx, req.BarberID)
	if err != nil {
		return nil, fmt.Errorf("barber not found: %w", err)
	}
	if barber.Status != "active" {
		return nil, fmt.Errorf("barber is not accepting bookings")
	}

	// ─────────────────────────────────────────────────────────────────
	// Step 3: Validate service exists and get pricing
	// ─────────────────────────────────────────────────────────────────
	barberService, err := s.serviceRepo.FindBarberServiceByID(ctx, req.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}
	if !barberService.IsActive {
		return nil, fmt.Errorf("service is not available")
	}

	// ─────────────────────────────────────────────────────────────────
	// Step 4: Check for time slot conflicts
	// ─────────────────────────────────────────────────────────────────
	endTime := s.calculateEndTime(req.StartTime, req.DurationMinutes)
	hasConflict, err := s.repo.CheckConflict(ctx, req.BarberID, req.StartTime, endTime, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	if hasConflict {
		return nil, fmt.Errorf("time slot is not available, please choose another time")
	}

	// ─────────────────────────────────────────────────────────────────
	// Step 5: Validate customer info
	// ─────────────────────────────────────────────────────────────────
	if req.CustomerID == nil {
		// Guest booking - require contact info
		if req.CustomerName == nil || *req.CustomerName == "" {
			return nil, fmt.Errorf("customer name is required for guest bookings")
		}
		if (req.CustomerEmail == nil || *req.CustomerEmail == "") &&
			(req.CustomerPhone == nil || *req.CustomerPhone == "") {
			return nil, fmt.Errorf("email or phone is required for guest bookings")
		}
	}

	// ─────────────────────────────────────────────────────────────────
	// Step 6: Calculate pricing
	// ─────────────────────────────────────────────────────────────────
	servicePrice := barberService.Price
	if req.ServicePrice != nil {
		servicePrice = *req.ServicePrice
	}

	discountAmount := 0.0
	if req.DiscountAmount != nil {
		discountAmount = *req.DiscountAmount
	}

	_, _, taxAmount, totalPrice := s.calculateTotalPrice(servicePrice, discountAmount)

	// ─────────────────────────────────────────────────────────────────
	// Step 7: Build booking model
	// ─────────────────────────────────────────────────────────────────
	booking := &models.Booking{
		UUID:          uuid.New().String(),
		BookingNumber: s.generateBookingNumber(),

		CustomerID: req.CustomerID,
		BarberID:   req.BarberID,

		ServiceName:              *barberService.CustomName,
		EstimatedDurationMinutes: req.DurationMinutes,

		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,

		Status: "pending",

		ServicePrice:   servicePrice,
		DiscountAmount: discountAmount,
		TaxAmount:      taxAmount,
		TotalPrice:     totalPrice,
		Currency:       "USD",

		PaymentStatus: "pending",

		Notes:           req.Notes,
		SpecialRequests: req.SpecialRequests,

		ScheduledStartTime: req.StartTime,
		ScheduledEndTime:   endTime,

		BookingSource: req.BookingSource,
	}

	if booking.BookingSource == "" {
		booking.BookingSource = "web_app"
	}

	// Get service category if available
	if barberService.Service.Name != "" {
		booking.ServiceName = barberService.Service.Name
	}

	// ─────────────────────────────────────────────────────────────────
	// Step 8: Create booking in database
	// ─────────────────────────────────────────────────────────────────
	if err := s.repo.Create(ctx, booking); err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// ─────────────────────────────────────────────────────────────────
	// Step 9: Create audit history
	// ─────────────────────────────────────────────────────────────────
	history := &models.BookingHistory{
		BookingID:  booking.ID,
		ChangedBy:  createdByUserID,
		ChangeType: "created",
		NewValues: models.JSONMap{
			"status":      booking.Status,
			"barber_id":   booking.BarberID,
			"start_time":  booking.ScheduledStartTime,
			"total_price": booking.TotalPrice,
		},
	}
	_ = s.repo.CreateHistory(ctx, history) // Don't fail if history creation fails

	// ─────────────────────────────────────────────────────────────────
	// Step 10: Invalidate cache
	// ─────────────────────────────────────────────────────────────────
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, req.BarberID)
	}

	return s.toBookingResponse(booking), nil
}

// ========================================================================
// READ OPERATIONS
// ========================================================================

// GetBookingByID retrieves a booking by ID
func (s *BookingService) GetBookingByID(ctx context.Context, id int) (*BookingResponse, error) {
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toBookingResponse(booking), nil
}

// GetBookingByUUID retrieves a booking by UUID
func (s *BookingService) GetBookingByUUID(ctx context.Context, uuid string) (*BookingResponse, error) {
	booking, err := s.repo.FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return s.toBookingResponse(booking), nil
}

// GetBookingByNumber retrieves a booking by booking number
func (s *BookingService) GetBookingByNumber(ctx context.Context, bookingNumber string) (*BookingResponse, error) {
	booking, err := s.repo.FindByBookingNumber(ctx, bookingNumber)
	if err != nil {
		return nil, err
	}
	return s.toBookingResponse(booking), nil
}

// GetCustomerBookings retrieves all bookings for a customer
func (s *BookingService) GetCustomerBookings(ctx context.Context, customerID int, filters repository.BookingFilters) ([]BookingResponse, error) {
	bookings, err := s.repo.FindByCustomerID(ctx, customerID, filters)
	if err != nil {
		return nil, err
	}

	responses := make([]BookingResponse, len(bookings))
	for i, booking := range bookings {
		responses[i] = *s.toBookingResponse(&booking)
	}
	return responses, nil
}

// GetBarberBookings retrieves all bookings for a barber
func (s *BookingService) GetBarberBookings(ctx context.Context, barberID int, filters repository.BookingFilters) ([]BookingResponse, error) {
	bookings, err := s.repo.FindByBarberID(ctx, barberID, filters)
	if err != nil {
		return nil, err
	}

	responses := make([]BookingResponse, len(bookings))
	for i, booking := range bookings {
		responses[i] = *s.toBookingResponse(&booking)
	}
	return responses, nil
}

// GetUpcomingBookings retrieves upcoming bookings
func (s *BookingService) GetUpcomingBookings(ctx context.Context, filters repository.BookingFilters) ([]BookingResponse, error) {
	bookings, err := s.repo.GetUpcomingBookings(ctx, filters)
	if err != nil {
		return nil, err
	}

	responses := make([]BookingResponse, len(bookings))
	for i, booking := range bookings {
		responses[i] = *s.toBookingResponse(&booking)
	}
	return responses, nil
}

// GetTodayBookings retrieves today's bookings for a barber
func (s *BookingService) GetTodayBookings(ctx context.Context, barberID int) ([]BookingResponse, error) {
	bookings, err := s.repo.GetTodayBookings(ctx, barberID)
	if err != nil {
		return nil, err
	}

	responses := make([]BookingResponse, len(bookings))
	for i, booking := range bookings {
		responses[i] = *s.toBookingResponse(&booking)
	}
	return responses, nil
}

// ========================================================================
// UPDATE OPERATIONS
// ========================================================================

// UpdateBooking updates booking details (not status)
func (s *BookingService) UpdateBooking(ctx context.Context, id int, req UpdateBookingRequest, updatedByUserID *int) (*BookingResponse, error) {
	// Get existing booking
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store old values for history
	oldValues := models.JSONMap{
		"customer_name":  booking.CustomerName,
		"customer_email": booking.CustomerEmail,
		"notes":          booking.Notes,
	}

	// Update fields if provided
	if req.CustomerName != nil {
		booking.CustomerName = req.CustomerName
	}
	if req.CustomerEmail != nil {
		booking.CustomerEmail = req.CustomerEmail
	}
	if req.CustomerPhone != nil {
		booking.CustomerPhone = req.CustomerPhone
	}
	if req.Notes != nil {
		booking.Notes = req.Notes
	}
	if req.SpecialRequests != nil {
		booking.SpecialRequests = req.SpecialRequests
	}
	if req.InternalNotes != nil {
		booking.InternalNotes = req.InternalNotes
	}

	// Save changes
	if err := s.repo.Update(ctx, booking); err != nil {
		return nil, err
	}

	// Create history
	history := &models.BookingHistory{
		BookingID:  booking.ID,
		ChangedBy:  updatedByUserID,
		ChangeType: "updated",
		OldValues:  oldValues,
		NewValues: models.JSONMap{
			"customer_name":  booking.CustomerName,
			"customer_email": booking.CustomerEmail,
			"notes":          booking.Notes,
		},
	}
	_ = s.repo.CreateHistory(ctx, history)

	return s.toBookingResponse(booking), nil
}

// UpdateStatus updates the booking status
func (s *BookingService) UpdateStatus(ctx context.Context, id int, newStatus string, updatedByUserID *int) (*BookingResponse, error) {
	// Get current booking for history
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := booking.Status

	// Update status (repository handles validation)
	if err := s.repo.UpdateStatus(ctx, id, newStatus); err != nil {
		return nil, err
	}

	// Create history
	history := &models.BookingHistory{
		BookingID:  booking.ID,
		ChangedBy:  updatedByUserID,
		ChangeType: "status_changed",
		OldValues:  models.JSONMap{"status": oldStatus},
		NewValues:  models.JSONMap{"status": newStatus},
	}
	_ = s.repo.CreateHistory(ctx, history)

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, booking.BarberID)
	}

	// Return updated booking
	return s.GetBookingByID(ctx, id)
}

// ========================================================================
// RESCHEDULE OPERATION
// ========================================================================

// RescheduleBooking reschedules a booking to a new time
func (s *BookingService) RescheduleBooking(ctx context.Context, id int, req RescheduleBookingRequest, rescheduledByUserID *int) (*BookingResponse, error) {
	// Get existing booking
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if booking can be rescheduled
	if !booking.CanBeCancelled() {
		return nil, fmt.Errorf("booking cannot be rescheduled in current status: %s", booking.Status)
	}

	// Determine duration
	durationMinutes := booking.EstimatedDurationMinutes
	if req.DurationMinutes > 0 {
		durationMinutes = req.DurationMinutes
	}

	// Validate new time
	if err := s.validateBookingTime(req.NewStartTime, durationMinutes); err != nil {
		return nil, err
	}

	// Check for conflicts (exclude current booking)
	newEndTime := s.calculateEndTime(req.NewStartTime, durationMinutes)
	hasConflict, err := s.repo.CheckConflict(ctx, booking.BarberID, req.NewStartTime, newEndTime, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	if hasConflict {
		return nil, fmt.Errorf("new time slot is not available")
	}

	// Store old values
	oldValues := models.JSONMap{
		"scheduled_start_time": booking.ScheduledStartTime,
		"scheduled_end_time":   booking.ScheduledEndTime,
	}

	// Update booking
	booking.ScheduledStartTime = req.NewStartTime
	booking.ScheduledEndTime = newEndTime
	booking.EstimatedDurationMinutes = durationMinutes

	if err := s.repo.Update(ctx, booking); err != nil {
		return nil, err
	}

	// Create history
	changeReason := req.Reason
	history := &models.BookingHistory{
		BookingID:  booking.ID,
		ChangedBy:  rescheduledByUserID,
		ChangeType: "rescheduled",
		OldValues:  oldValues,
		NewValues: models.JSONMap{
			"scheduled_start_time": booking.ScheduledStartTime,
			"scheduled_end_time":   booking.ScheduledEndTime,
		},
		ChangeReason: changeReason,
	}
	_ = s.repo.CreateHistory(ctx, history)

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, booking.BarberID)
	}

	return s.toBookingResponse(booking), nil
}

// ========================================================================
// CANCEL OPERATION
// ========================================================================

// CancelBooking cancels a booking
func (s *BookingService) CancelBooking(ctx context.Context, id int, req CancelBookingRequest, cancelledByUserID int) error {
	// Get booking to validate and for cache invalidation
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Cancel the booking
	if err := s.repo.Cancel(ctx, id, cancelledByUserID, req.Reason, req.IsByCustomer); err != nil {
		return err
	}

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, booking.BarberID)
	}

	return nil
}

// ========================================================================
// STATISTICS
// ========================================================================

// GetBarberStats retrieves booking statistics for a barber
func (s *BookingService) GetBarberStats(ctx context.Context, barberID int, from, to time.Time) (*repository.BookingStats, error) {
	return s.repo.GetBarberStats(ctx, barberID, from, to)
}

// GetBookingHistory retrieves the audit history for a booking
func (s *BookingService) GetBookingHistory(ctx context.Context, bookingID int) ([]models.BookingHistory, error) {
	return s.repo.GetHistory(ctx, bookingID)
}

// ========================================================================
// AVAILABILITY CHECK
// ========================================================================

// CheckAvailability checks if a time slot is available for a barber
func (s *BookingService) CheckAvailability(ctx context.Context, barberID int, startTime time.Time, durationMinutes int) (bool, error) {
	endTime := s.calculateEndTime(startTime, durationMinutes)
	hasConflict, err := s.repo.CheckConflict(ctx, barberID, startTime, endTime, 0)
	if err != nil {
		return false, err
	}
	return !hasConflict, nil
}
