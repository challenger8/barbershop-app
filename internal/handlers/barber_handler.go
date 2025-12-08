// internal/handlers/barber_handler.go
package handlers

import (
	"fmt"
	"net/http"

	"barber-booking-system/internal/config"
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"

	"github.com/gin-gonic/gin"
)

// BarberHandler handles HTTP requests for barbers
type BarberHandler struct {
	barberService *services.BarberService
}

// NewBarberHandler creates a new barber handler
func NewBarberHandler(barberService *services.BarberService) *BarberHandler {
	return &BarberHandler{
		barberService: barberService,
	}
}

// GetAllBarbers godoc
// @Summary Get all barbers
// @Description Get list of all barbers with optional filters
// @Tags barbers
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param city query string false "Filter by city"
// @Param state query string false "Filter by state"
// @Param min_rating query number false "Minimum rating"
// @Param search query string false "Search term"
// @Param sort_by query string false "Sort by field (rating, total_bookings, shop_name)"
// @Param limit query int false "Number of results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers [get]
func (h *BarberHandler) GetAllBarbers(c *gin.Context) {
	filters := repository.BarberFilters{
		Status:    c.Query("status"),
		City:      c.Query("city"),
		State:     c.Query("state"),
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		Limit:     ParseIntQuery(c, "limit", 20),
		Offset:    ParseIntQuery(c, "offset", 0),
		MinRating: ParseFloatQuery(c, "min_rating", 0),
	}

	if verifiedStr := c.Query("is_verified"); verifiedStr != "" {
		verified := verifiedStr == "true"
		filters.IsVerified = &verified
	}

	// Get barbers
	barbers, err := h.barberService.GetAllBarbers(c.Request.Context(), filters)
	if err != nil {
		RespondInternalError(c, "fetch barbers", err)
		return
	}

	RespondSuccessWithMeta(c, barbers, map[string]interface{}{
		"count":  len(barbers),
		"limit":  filters.Limit,
		"offset": filters.Offset,
	})
}

// GetBarber godoc
// @Summary Get barber by ID
// @Description Get detailed information about a specific barber
// @Tags barbers
// @Accept json
// @Produce json
// @Param id path int true "Barber ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{id} [get]
func (h *BarberHandler) GetBarber(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	barber, err := h.barberService.GetBarberByID(c.Request.Context(), id)
	if HandleServiceError(c, err, "Barber", "fetch barber") {
		return
	}

	RespondSuccess(c, barber)
}

// CreateBarber godoc
// @Summary Create new barber
// @Description Create a new barber profile
// @Tags barbers
// @Accept json
// @Produce json
// @Param barber body services.CreateBarberRequest true "Barber data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/barbers [post]
func (h *BarberHandler) CreateBarber(c *gin.Context) {
	req, ok := BindJSON[services.CreateBarberRequest](c)
	if !ok {
		return
	}

	barber, err := h.barberService.CreateBarber(c.Request.Context(), *req)
	if HandleServiceError(c, err, "Barber", "create barber") {
		return
	}

	RespondCreated(c, barber, "Barber created successfully")
}

// GetBarberByUUID godoc
// @Summary Get barber by UUID
// @Description Get detailed information about a specific barber by UUID
// @Tags barbers
// @Accept json
// @Produce json
// @Param uuid path string true "Barber UUID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/uuid/{uuid} [get]
func (h *BarberHandler) GetBarberByUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	barber, err := h.barberService.GetBarberByUUID(c.Request.Context(), uuid)
	if HandleServiceError(c, err, "Barber", "fetch barber") {
		return
	}

	RespondSuccess(c, barber)
}

// UpdateBarber godoc
// @Summary Update barber
// @Description Update barber information
// @Tags barbers
// @Accept json
// @Produce json
// @Param id path int true "Barber ID"
// @Param barber body services.UpdateBarberRequest true "Updated barber data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{id} [put]
func (h *BarberHandler) UpdateBarber(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	req, ok := BindJSON[services.UpdateBarberRequest](c)
	if !ok {
		return
	}

	barber, err := h.barberService.UpdateBarber(c.Request.Context(), id, *req)
	if HandleServiceError(c, err, "Barber", "update barber") {
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    barber,
		Message: "Barber updated successfully",
	})
}

// DeleteBarber godoc
// @Summary Delete barber
// @Description Soft delete a barber
// @Tags barbers
// @Accept json
// @Produce json
// @Param id path int true "Barber ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{id} [delete]
func (h *BarberHandler) DeleteBarber(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	if err := h.barberService.DeleteBarber(c.Request.Context(), id); err != nil {
		if HandleServiceError(c, err, "Barber", "delete barber") {
			return
		}
	}

	RespondSuccessWithMessage(c, "Barber deleted successfully")
}

// UpdateBarberStatus godoc
// @Summary Update barber status
// @Description Update the status of a barber (pending, active, inactive, suspended, rejected)
// @Tags barbers
// @Accept json
// @Produce json
// @Param id path int true "Barber ID"
// @Param status body StatusUpdateRequest true "Status data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{id}/status [patch]
// ========================================================================
// EXACT CODE TO ADD TO: internal/handlers/barber_handler.go
// ========================================================================
//
// Find the UpdateBarberStatus function (around line 223)
// Replace the ENTIRE function with this version:
// ========================================================================

func (h *BarberHandler) UpdateBarberStatus(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	req, ok := BindJSON[StatusUpdateRequest](c)
	if !ok {
		return
	}

	// ⭐ VALIDATE STATUS BEFORE CALLING SERVICE ⭐
	validStatuses := []string{
		config.BarberStatusPending,   // "pending"
		config.BarberStatusActive,    // "active"
		config.BarberStatusInactive,  // "inactive"
		config.BarberStatusSuspended, // "suspended"
		config.BarberStatusRejected,  // "rejected"
	}

	isValid := false
	for _, s := range validStatuses {
		if req.Status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(400, middleware.ErrorResponse{
			Error:   "Invalid status",
			Message: fmt.Sprintf("Status must be one of: %v. Got: %s", validStatuses, req.Status),
		})
		return
	}

	// Now call service with validated status
	if err := h.barberService.UpdateBarberStatus(c.Request.Context(), id, req.Status); err != nil {
		if HandleServiceError(c, err, "Barber", "update barber status") {
			return
		}
	}

	RespondSuccessWithMessage(c, "Status updated successfully")
}

// GetBarberStatistics godoc
// @Summary Get barber statistics
// @Description Get comprehensive statistics for a barber
// @Tags barbers
// @Accept json
// @Produce json
// @Param id path int true "Barber ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{id}/statistics [get]
func (h *BarberHandler) GetBarberStatistics(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	stats, err := h.barberService.GetBarberStatistics(c.Request.Context(), id)
	if HandleServiceError(c, err, "Barber", "fetch barber statistics") {
		return
	}

	RespondSuccess(c, stats)
}

// SearchBarbers godoc
// @Summary Search barbers
// @Description Search barbers by query string
// @Tags barbers
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param city query string false "Filter by city"
// @Param state query string false "Filter by state"
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/search [get]
func (h *BarberHandler) SearchBarbers(c *gin.Context) {
	query := c.Query("q")
	filters := repository.BarberFilters{
		City:   c.Query("city"),
		Name:   c.Query("user_name"),
		State:  c.Query("state"),
		Status: "active",
		Limit:  50,
	}

	barbers, err := h.barberService.SearchBarbers(c.Request.Context(), query, filters)
	if err != nil {
		RespondInternalError(c, "search barbers", err)
		return
	}

	RespondSuccessWithMeta(c, barbers, map[string]interface{}{
		"query": query,
		"count": len(barbers),
	})
}

// Request types (handler-specific)
type StatusUpdateRequest struct {
	Status string `json:"status" binding:"required"`
}
