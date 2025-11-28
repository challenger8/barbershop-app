// internal/handlers/booking_handler.go
package handlers

import (
	"net/http"
	"time"

	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"

	"github.com/gin-gonic/gin"
)

// ========================================================================
// BOOKING HANDLER - HTTP Request Handlers for Bookings
// ========================================================================

// BookingHandler handles booking-related HTTP requests
type BookingHandler struct {
	bookingService *services.BookingService
}

// NewBookingHandler creates a new booking handler
func NewBookingHandler(bookingService *services.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// ========================================================================
// HELPER FUNCTIONS
// ========================================================================

// buildBookingFilters builds BookingFilters from query parameters
func buildBookingFilters(c *gin.Context) repository.BookingFilters {
	return repository.BookingFilters{
		Status:        c.Query("status"),
		PaymentStatus: c.Query("payment_status"),
		BookingSource: c.Query("booking_source"),
		Search:        c.Query("search"),
		StartDateFrom: ParseTimeQuery(c, "start_date_from"), // Shared function
		StartDateTo:   ParseTimeQuery(c, "start_date_to"),   // Shared function
		CreatedFrom:   ParseTimeQuery(c, "created_from"),    // Shared function
		CreatedTo:     ParseTimeQuery(c, "created_to"),      // Shared function
		SortBy:        c.Query("sort_by"),
		Order:         c.Query("order"),
		Limit:         ParseIntQuery(c, "limit", 50), // Shared function
		Offset:        ParseIntQuery(c, "offset", 0), // Shared function
	}
}

// ========================================================================
// CREATE BOOKING
// ========================================================================

// CreateBooking godoc
// @Summary Create a new booking
// @Description Create a new appointment booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking body services.CreateBookingRequest true "Booking data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 409 {object} middleware.ErrorResponse "Time slot conflict"
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req services.CreateBookingRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Get user ID from auth context (if authenticated)
	var createdByUserID *int
	if userID, exists := middleware.GetUserID(c); exists {
		createdByUserID = &userID
		// If customer_id not provided, use authenticated user
		if req.CustomerID == nil {
			req.CustomerID = &userID
		}
	}

	// Create booking
	booking, err := h.bookingService.CreateBooking(c.Request.Context(), req, createdByUserID)
	if err != nil {
		// Check for specific error types
		statusCode := http.StatusInternalServerError
		if err.Error() == "time slot is not available, please choose another time" {
			statusCode = http.StatusConflict
		} else if containsAny(err.Error(), []string{"not found", "required", "must be", "cannot"}) {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, middleware.ErrorResponse{
			Error:   "Failed to create booking",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    booking,
		Message: "Booking created successfully",
	})
}

// containsAny checks if string contains any of the substrings
func containsAny(s string, substrings []string) bool {
	for _, sub := range substrings {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// ========================================================================
// GET BOOKING BY ID
// ========================================================================

// GetBooking godoc
// @Summary Get booking by ID
// @Description Get detailed information about a specific booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/bookings/{id} [get]
func (h *BookingHandler) GetBooking(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "booking")
	if !ok {
		return
	}

	booking, err := h.bookingService.GetBookingByID(c.Request.Context(), id)

	// Get booking
	if err != nil {
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Booking not found",
				Message: "No booking found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch booking",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    booking,
	})
}

// GetBookingByUUID godoc
// @Summary Get booking by UUID
// @Description Get detailed information about a specific booking by UUID
// @Tags bookings
// @Accept json
// @Produce json
// @Param uuid path string true "Booking UUID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/bookings/uuid/{uuid} [get]
func (h *BookingHandler) GetBookingByUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	booking, err := h.bookingService.GetBookingByUUID(c.Request.Context(), uuid)
	if err != nil {
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Booking not found",
				Message: "No booking found with the given UUID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch booking",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    booking,
	})
}

// GetBookingByNumber godoc
// @Summary Get booking by booking number
// @Description Get detailed information about a specific booking by its human-readable booking number
// @Tags bookings
// @Accept json
// @Produce json
// @Param number path string true "Booking Number (e.g., BK202411281234)"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/bookings/number/{number} [get]
func (h *BookingHandler) GetBookingByNumber(c *gin.Context) {
	bookingNumber := c.Param("number")

	booking, err := h.bookingService.GetBookingByNumber(c.Request.Context(), bookingNumber)
	if err != nil {
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Booking not found",
				Message: "No booking found with the given booking number",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch booking",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    booking,
	})
}

// ========================================================================
// GET MY BOOKINGS (Customer)
// ========================================================================

// GetMyBookings godoc
// @Summary Get my bookings
// @Description Get all bookings for the authenticated customer
// @Tags bookings
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param payment_status query string false "Filter by payment status"
// @Param start_date_from query string false "Filter by start date from (RFC3339)"
// @Param start_date_to query string false "Filter by start date to (RFC3339)"
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param order query string false "Sort order (ASC/DESC)" default(DESC)
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/bookings/me [get]
func (h *BookingHandler) GetMyBookings(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.ErrorResponse{
			Error:   "Unauthorized",
			Message: "You must be logged in to view your bookings",
		})
		return
	}

	// Build filters from query params
	filters := buildBookingFilters(c)

	// Get bookings
	bookings, err := h.bookingService.GetCustomerBookings(c.Request.Context(), userID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch bookings",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    bookings,
		Meta: map[string]interface{}{
			"count":  len(bookings),
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
	})
}

// ========================================================================
// GET BARBER'S BOOKINGS
// ========================================================================

// GetBarberBookings godoc
// @Summary Get barber's bookings
// @Description Get all bookings for a specific barber
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Barber ID"
// @Param status query string false "Filter by status"
// @Param payment_status query string false "Filter by payment status"
// @Param start_date_from query string false "Filter by start date from (RFC3339)"
// @Param start_date_to query string false "Filter by start date to (RFC3339)"
// @Param sort_by query string false "Sort by field" default(scheduled_start_time)
// @Param order query string false "Sort order (ASC/DESC)" default(ASC)
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{id}/bookings [get]
func (h *BookingHandler) GetBarberBookings(c *gin.Context) {
	barberID, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	// Build filters from query params
	filters := buildBookingFilters(c)
	// Default sort for barber view is by scheduled time
	if c.Query("sort_by") == "" {
		filters.SortBy = "scheduled_start_time"
		filters.Order = "ASC"
	}

	// Get bookings
	bookings, err := h.bookingService.GetBarberBookings(c.Request.Context(), barberID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch bookings",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    bookings,
		Meta: map[string]interface{}{
			"barber_id": barberID,
			"count":     len(bookings),
			"limit":     filters.Limit,
			"offset":    filters.Offset,
		},
	})
}

// GetTodayBookings godoc
// @Summary Get today's bookings for a barber
// @Description Get all bookings scheduled for today for a specific barber
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Barber ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{id}/bookings/today [get]
func (h *BookingHandler) GetTodayBookings(c *gin.Context) {
	barberID, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	// Get today's bookings
	bookings, err := h.bookingService.GetTodayBookings(c.Request.Context(), barberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch today's bookings",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    bookings,
		Meta: map[string]interface{}{
			"barber_id": barberID,
			"date":      time.Now().Format("2006-01-02"),
			"count":     len(bookings),
		},
	})
}

// ========================================================================
// UPDATE BOOKING
// ========================================================================

// UpdateBooking godoc
// @Summary Update booking details
// @Description Update booking information (not status)
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param booking body services.UpdateBookingRequest true "Updated booking data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/bookings/{id} [put]
func (h *BookingHandler) UpdateBooking(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "booking")
	if !ok {
		return
	}

	// Parse request body
	var req services.UpdateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Get user ID from auth context
	var updatedByUserID *int
	if userID, exists := middleware.GetUserID(c); exists {
		updatedByUserID = &userID
	}

	// Update booking
	booking, err := h.bookingService.UpdateBooking(c.Request.Context(), id, req, updatedByUserID)
	if err != nil {
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Booking not found",
				Message: "No booking found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to update booking",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    booking,
		Message: "Booking updated successfully",
	})
}

// ========================================================================
// UPDATE STATUS
// ========================================================================

// UpdateBookingStatus godoc
// @Summary Update booking status
// @Description Update the status of a booking (pending → confirmed → in_progress → completed)
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param status body services.UpdateStatusRequest true "New status"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 422 {object} middleware.ErrorResponse "Invalid status transition"
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/bookings/{id}/status [patch]
func (h *BookingHandler) UpdateBookingStatus(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "booking")
	if !ok {
		return
	}

	// Step 2: Parse request body
	var req services.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Step 3: Get user ID from auth context
	var updatedByUserID *int
	if userID, exists := middleware.GetUserID(c); exists {
		updatedByUserID = &userID
	}

	// Step 4: Update status
	booking, err := h.bookingService.UpdateStatus(c.Request.Context(), id, req.Status, updatedByUserID)
	if err != nil {
		// Step 5: Handle errors
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Booking not found",
				Message: "No booking found with the given ID",
			})
			return
		}
		// Check for invalid status transition
		if containsAny(err.Error(), []string{"invalid status", "cannot change"}) {
			c.JSON(http.StatusUnprocessableEntity, middleware.ErrorResponse{
				Error:   "Invalid status transition",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to update booking status",
			Message: err.Error(),
		})
		return
	}

	// Step 6: Return success
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    booking,
		Message: "Booking status updated successfully",
	})
}

// ========================================================================
// RESCHEDULE BOOKING
// ========================================================================

// RescheduleBooking godoc
// @Summary Reschedule a booking
// @Description Change the scheduled time of a booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param reschedule body services.RescheduleBookingRequest true "New schedule"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 409 {object} middleware.ErrorResponse "Time slot conflict"
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/bookings/{id}/reschedule [put]
func (h *BookingHandler) RescheduleBooking(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "booking")
	if !ok {
		return
	}

	// Parse request body
	var req services.RescheduleBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Get user ID from auth context
	var rescheduledByUserID *int
	if userID, exists := middleware.GetUserID(c); exists {
		rescheduledByUserID = &userID
	}

	// Reschedule booking
	booking, err := h.bookingService.RescheduleBooking(c.Request.Context(), id, req, rescheduledByUserID)
	if err != nil {
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Booking not found",
				Message: "No booking found with the given ID",
			})
			return
		}
		if containsAny(err.Error(), []string{"not available", "conflict"}) {
			c.JSON(http.StatusConflict, middleware.ErrorResponse{
				Error:   "Time slot not available",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to reschedule booking",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    booking,
		Message: "Booking rescheduled successfully",
	})
}

// ========================================================================
// CANCEL BOOKING
// ========================================================================

// CancelBooking godoc
// @Summary Cancel a booking
// @Description Cancel an existing booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param cancel body services.CancelBookingRequest false "Cancellation details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 422 {object} middleware.ErrorResponse "Cannot cancel booking"
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/bookings/{id} [delete]
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "booking")
	if !ok {
		return
	}
	// Step 2: Parse request body (optional)
	var req services.CancelBookingRequest
	// Ignore error if body is empty - cancellation reason is optional
	_ = c.ShouldBindJSON(&req)

	// Step 3: Get user ID from auth context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.ErrorResponse{
			Error:   "Unauthorized",
			Message: "You must be logged in to cancel a booking",
		})
		return
	}

	// Step 4: Cancel booking
	err := h.bookingService.CancelBooking(c.Request.Context(), id, req, userID)
	if err != nil {
		// Step 5: Handle errors
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Booking not found",
				Message: "No booking found with the given ID",
			})
			return
		}
		if err == repository.ErrCancellationNotAllowed {
			c.JSON(http.StatusUnprocessableEntity, middleware.ErrorResponse{
				Error:   "Cannot cancel booking",
				Message: "This booking cannot be cancelled in its current status",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to cancel booking",
			Message: err.Error(),
		})
		return
	}

	// Step 6: Return success
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Booking cancelled successfully",
	})
}

// ========================================================================
// CHECK AVAILABILITY
// ========================================================================

// CheckAvailability godoc
// @Summary Check time slot availability
// @Description Check if a specific time slot is available for a barber
// @Tags bookings
// @Accept json
// @Produce json
// @Param barber_id query int true "Barber ID"
// @Param start_time query string true "Start time (RFC3339)"
// @Param duration query int true "Duration in minutes"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/bookings/availability [get]
func (h *BookingHandler) CheckAvailability(c *gin.Context) {
	barberID := ParseIntQuery(c, "barber_id", 0) // Shared function
	if barberID == 0 {
		RespondBadRequest(c, "Missing barber_id", "barber_id query parameter is required")
		return
	}

	startTime := ParseTimeQuery(c, "start_time")
	if startTime.IsZero() {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid start_time",
			Message: "start_time query parameter is required (RFC3339 format)",
		})
		return
	}

	duration := ParseIntQuery(c, "duration", 0)
	if duration == 0 {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Missing duration",
			Message: "duration query parameter is required (in minutes)",
		})
		return
	}

	// Check availability
	available, err := h.bookingService.CheckAvailability(c.Request.Context(), barberID, startTime, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to check availability",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: map[string]interface{}{
			"barber_id":  barberID,
			"start_time": startTime,
			"duration":   duration,
			"end_time":   startTime.Add(time.Duration(duration) * time.Minute),
			"available":  available,
		},
	})
}

// ========================================================================
// GET BOOKING STATISTICS
// ========================================================================

// GetBarberBookingStats godoc
// @Summary Get booking statistics for a barber
// @Description Get booking statistics (total, completed, cancelled, revenue) for a barber
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Barber ID"
// @Param from query string false "From date (RFC3339)" default(30 days ago)
// @Param to query string false "To date (RFC3339)" default(now)
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{id}/bookings/stats [get]
func (h *BookingHandler) GetBarberBookingStats(c *gin.Context) {
	barberID, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	// Parse date range (default to last 30 days)
	to := ParseTimeQuery(c, "to")

	
	if to.IsZero() {
		to = time.Now()
	}

	from := ParseTimeQuery(c, "from")
	if from.IsZero() {
		from = to.AddDate(0, 0, -30) // 30 days ago
	}

	// Get statistics
	stats, err := h.bookingService.GetBarberStats(c.Request.Context(), barberID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch statistics",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    stats,
		Meta: map[string]interface{}{
			"barber_id": barberID,
			"from":      from,
			"to":        to,
		},
	})
}

// ========================================================================
// GET BOOKING HISTORY (Audit Trail)
// ========================================================================

// GetBookingHistory godoc
// @Summary Get booking history
// @Description Get the audit trail of changes for a booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/bookings/{id}/history [get]
func (h *BookingHandler) GetBookingHistory(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "booking")
	if !ok {
		return
	}
	// Get history
	history, err := h.bookingService.GetBookingHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch booking history",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    history,
		Meta: map[string]interface{}{
			"booking_id": id,
			"count":      len(history),
		},
	})
}
