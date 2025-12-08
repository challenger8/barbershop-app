// internal/services/booking_service.go
package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/config"
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
func (s *BookingService) calculateTotalPrice(servicePrice float64, discountAmount float64) *models.PricingBreakdown {
	taxRate := 0.08
	return models.CalculatePricing(servicePrice, discountAmount, taxRate)
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
// EXTRACTED HELPER FUNCTIONS FOR CreateBooking
// ========================================================================
// Add these functions to booking_service.go after toBookingResponse() (around line 238)

// validateAndFetchBarber validates barber exists and is active
func (s *BookingService) validateAndFetchBarber(ctx context.Context, barberID int) (*models.Barber, error) {
	barber, err := s.barberRepo.FindByID(ctx, barberID)
	if err != nil {
		return nil, fmt.Errorf("barber not found: %w", err)
	}

	if barber.Status != config.BarberStatusActive {
		return nil, fmt.Errorf("barber is not accepting bookings")
	}

	return barber, nil
}

// validateAndFetchBarberService validates service exists, is active, and belongs to barber
func (s *BookingService) validateAndFetchBarberService(ctx context.Context, serviceID int) (*models.BarberService, error) {
	barberService, err := s.serviceRepo.FindBarberServiceByID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}

	if !barberService.IsActive {
		return nil, fmt.Errorf("service is not available")
	}

	return barberService, nil
}
func (s *BookingService) checkTimeSlotAvailability(
	ctx context.Context,
	barberID int,
	startTime, endTime time.Time,
	excludeBookingID int,
) error {
	opts := models.NewTimeSlotCheckOptions(startTime, endTime,
		models.WithExcludeBooking(excludeBookingID))
	return s.checkTimeSlotAvailabilityWithOptions(ctx, barberID, opts)
}
func (s *BookingService) checkTimeSlotAvailabilityWithOptions(
	ctx context.Context,
	barberID int,
	opts *models.TimeSlotCheckOptions,
) error {
	// Get effective times (with buffer if specified)
	effectiveStart := opts.GetEffectiveStartTime()
	effectiveEnd := opts.GetEffectiveEndTime()

	// Check for conflicts
	hasConflict, err := s.repo.CheckConflict(
		ctx,
		barberID,
		effectiveStart,
		effectiveEnd,
		opts.ExcludeBookingID,
	)
	if err != nil {
		return fmt.Errorf("failed to check availability: %w", err)
	}

	if hasConflict {
		return fmt.Errorf("time slot is not available, please choose another time")
	}

	return nil
}

// validateCustomerInfo ensures either customer_id or guest contact info is provided
func (s *BookingService) validateCustomerInfo(req CreateBookingRequest) error {
	if req.CustomerID != nil {
		return nil // Customer ID provided, no additional validation needed
	}

	// Guest booking - require contact info
	if req.CustomerName == nil || *req.CustomerName == "" {
		return fmt.Errorf("customer name is required for guest bookings")
	}

	if (req.CustomerEmail == nil || *req.CustomerEmail == "") &&
		(req.CustomerPhone == nil || *req.CustomerPhone == "") {
		return fmt.Errorf("email or phone is required for guest bookings")
	}

	return nil
}

// PricingResult holds calculated pricing breakdown
type PricingResult struct {
	ServicePrice   float64
	DiscountAmount float64
	TaxAmount      float64
	TotalPrice     float64
}

// calculateBookingPricing calculates all pricing components
func (s *BookingService) calculateBookingPricing(barberService *models.BarberService, req CreateBookingRequest) PricingResult {
	// Use provided price or default to barber service price
	servicePrice := barberService.Price
	if req.ServicePrice != nil {
		servicePrice = *req.ServicePrice
	}

	// Apply discount if provided
	discountAmount := 0.0
	if req.DiscountAmount != nil {
		discountAmount = *req.DiscountAmount
	}

	// Calculate tax and total
	pricing := s.calculateTotalPrice(servicePrice, discountAmount)

	return PricingResult{
		ServicePrice:   pricing.ServicePrice,
		DiscountAmount: pricing.DiscountAmount,
		TaxAmount:      pricing.TaxAmount,
		TotalPrice:     pricing.TotalPrice,
	}
}

// buildBookingFromRequest constructs a booking model from request data
func (s *BookingService) buildBookingFromRequest(
	req CreateBookingRequest,
	barberService *models.BarberService,
	pricing PricingResult,
	endTime time.Time,
) *models.Booking {
	booking := &models.Booking{
		UUID:          uuid.New().String(),
		BookingNumber: s.generateBookingNumber(),

		CustomerID: req.CustomerID,
		BarberID:   req.BarberID,

		ServiceName:              getServiceName(barberService),
		EstimatedDurationMinutes: req.DurationMinutes,

		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,

		Status: config.BookingStatusPending,

		ServicePrice:   pricing.ServicePrice,
		DiscountAmount: pricing.DiscountAmount,
		TaxAmount:      pricing.TaxAmount,
		TotalPrice:     pricing.TotalPrice,
		Currency:       config.DefaultCurrency,

		PaymentStatus: config.PaymentStatusPending,

		Notes:           req.Notes,
		SpecialRequests: req.SpecialRequests,

		ScheduledStartTime: req.StartTime,
		ScheduledEndTime:   endTime,

		BookingSource: getBookingSource(req.BookingSource),
	}

	return booking
}

// getServiceName extracts the appropriate service name
func getServiceName(barberService *models.BarberService) string {
	if barberService.CustomName != nil && *barberService.CustomName != "" {
		return *barberService.CustomName
	}
	if barberService.Service.Name != "" {
		return barberService.Service.Name
	}
	return "Unknown Service"
}

// getBookingSource returns booking source or default
func getBookingSource(source string) string {
	if source != "" {
		return source
	}
	return "web_app"
}

// saveBookingWithHistory saves booking and creates audit trail
func (s *BookingService) saveBookingWithHistory(ctx context.Context, booking *models.Booking, createdByUserID *int) error {
	// Create booking in database
	if err := s.repo.Create(ctx, booking); err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}

	// Create audit history
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

	return nil
}

// ========================================================================
// CREATE BOOKING
// ========================================================================

// CreateBooking creates a new booking with full validation
// ========================================================================
// REFACTORED CreateBooking - REPLACE YOUR EXISTING ONE (lines 239-368)
// ========================================================================

// CreateBooking creates a new booking with full validation
func (s *BookingService) CreateBooking(ctx context.Context, req CreateBookingRequest, createdByUserID *int) (*BookingResponse, error) {
	// Step 1: Validate booking time
	if err := s.validateBookingTime(req.StartTime, req.DurationMinutes); err != nil {
		return nil, err
	}

	// Step 2: Validate barber exists and is active
	barber, err := s.validateAndFetchBarber(ctx, req.BarberID)
	if err != nil {
		return nil, err
	}

	// Step 3: Validate service exists and get pricing
	barberService, err := s.validateAndFetchBarberService(ctx, req.ServiceID)
	if err != nil {
		return nil, err
	}

	// Step 4: Check for time slot conflicts
	endTime := s.calculateEndTime(req.StartTime, req.DurationMinutes)
	if err := s.checkTimeSlotAvailability(ctx, req.BarberID, req.StartTime, endTime, 0); err != nil {
		return nil, err
	}

	// Step 5: Validate customer info
	if err := s.validateCustomerInfo(req); err != nil {
		return nil, err
	}

	// Step 6: Calculate pricing
	pricing := s.calculateBookingPricing(barberService, req)

	// Step 7: Build booking model
	booking := s.buildBookingFromRequest(req, barberService, pricing, endTime)

	// Step 8: Save booking with audit trail
	if err := s.saveBookingWithHistory(ctx, booking, createdByUserID); err != nil {
		return nil, err
	}

	// Step 9: Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, req.BarberID)
	}

	// Suppress unused variable warning (can be removed if barber is used elsewhere)
	_ = barber

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

// ========================================================================
// UPDATED: internal/services/booking_service.go
// ========================================================================
//
// Replace the UpdateStatus method (around line 613-645) with this version:
// ========================================================================

// UpdateStatus updates the booking status with state machine validation
func (s *BookingService) UpdateStatus(ctx context.Context, id int, newStatus string, updatedByUserID *int) (*BookingResponse, error) {
	// Get current booking
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := booking.Status

	// ⭐ VALIDATE STATE TRANSITION USING STATE MACHINE ⭐
	if err := booking.ValidateStatusTransition(newStatus); err != nil {
		return nil, fmt.Errorf("invalid status transition: %w", err)
	}

	// Update status in database
	if err := s.repo.UpdateStatus(ctx, id, newStatus); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	// Create audit history
	history := &models.BookingHistory{
		BookingID:  booking.ID,
		ChangedBy:  updatedByUserID,
		ChangeType: "status_changed",
		OldValues:  models.JSONMap{"status": oldStatus},
		NewValues:  models.JSONMap{"status": newStatus},
	}

	// Log history (don't fail if history creation fails)
	if err := s.repo.CreateHistory(ctx, history); err != nil {
		// Log error but don't fail the status update
		fmt.Printf("Warning: failed to create booking history: %v\n", err)
	}

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, booking.BarberID)
	}

	// Return updated booking
	return s.GetBookingByID(ctx, id)
}

// ========================================================================
// NEW HELPER METHOD: Get Allowed Transitions
// ========================================================================

// GetAllowedStatusTransitions returns the valid next states for a booking
// This is useful for API clients to know what actions are available
func (s *BookingService) GetAllowedStatusTransitions(ctx context.Context, id int) ([]string, error) {
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return booking.GetAllowedStatusTransitions(), nil
}

// ========================================================================
// UPDATED: Cancel Booking Method
// ========================================================================

// CancelBooking cancels a booking with state machine validation
func (s *BookingService) CancelBooking(ctx context.Context, id int, req CancelBookingRequest, cancelledByUserID *int) (*BookingResponse, error) {
	// Get existing booking
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// ⭐ CHECK IF CANCELLATION IS ALLOWED USING STATE MACHINE ⭐
	if !booking.CanTransitionTo(config.BookingStatusCancelled) {
		return nil, fmt.Errorf(
			"booking cannot be cancelled from status '%s'. Current status must be one of: %v",
			booking.Status,
			[]string{config.BookingStatusPending, config.BookingStatusConfirmed, config.BookingStatusInProgress},
		)
	}

	// Check if already in a terminal state
	if booking.IsInTerminalState() {
		return nil, fmt.Errorf("booking is already in a terminal state: %s", booking.Status)
	}

	// Update status to cancelled
	return s.UpdateStatus(ctx, id, config.BookingStatusCancelled, cancelledByUserID)
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

// ========================================================================
// STATISTICS
// ========================================================================
func (s *BookingService) GetBarberStats(
	ctx context.Context,
	barberID int,
	from, to time.Time,
) (*repository.BookingStats, error) {
	opts := models.NewStatsQueryOptions(from, to)
	return s.GetBarberStatsEnhanced(ctx, barberID, opts)
}

// GetBarberStats retrieves booking statistics for a barber
func (s *BookingService) GetBarberStatsEnhanced(
	ctx context.Context,
	barberID int,
	opts *models.StatsQueryOptions,
) (*repository.BookingStats, error) {
	// Get base stats
	stats, err := s.repo.GetBarberStats(ctx, barberID, opts.FromDate, opts.ToDate)
	if err != nil {
		return nil, err
	}

	// Add trends if requested
	if opts.IncludeTrends {
		// TODO: Calculate trends
		// This would compare current period with previous period
	}

	// Add ratings if requested
	if opts.IncludeRatings {
		// TODO: Add rating metrics
		// This would aggregate review data
	}

	return stats, nil
}

// GetBookingHistory retrieves the audit history for a booking
func (s *BookingService) GetBookingHistory(ctx context.Context, bookingID int) ([]models.BookingHistory, error) {
	return s.repo.GetHistory(ctx, bookingID)
}

// ========================================================================
// AVAILABILITY CHECK
// ========================================================================

// CheckAvailability checks if a time slot is available for a barber
func (s *BookingService) CheckAvailabilityEnhanced(
	ctx context.Context,
	barberID int,
	startTime time.Time,
	durationMinutes int,
	opts ...models.TimeSlotCheckOption,
) (bool, error) {
	endTime := s.calculateEndTime(startTime, durationMinutes)

	// Create options with any additional settings
	checkOpts := models.NewTimeSlotCheckOptions(startTime, endTime, opts...)

	// Use the enhanced check
	err := s.checkTimeSlotAvailabilityWithOptions(ctx, barberID, checkOpts)
	if err != nil {
		return false, nil // Not available (has conflict)
	}

	return true, nil // Available
}

// KEEP OLD VERSION FOR BACKWARD COMPATIBILITY
func (s *BookingService) CheckAvailability(
	ctx context.Context,
	barberID int,
	startTime time.Time,
	durationMinutes int,
) (bool, error) {
	return s.CheckAvailabilityEnhanced(ctx, barberID, startTime, durationMinutes)
}
