// internal/handlers/barber_handler.go
package handlers

import (
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"
	"net/http"

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
		Limit:     ParseIntQuery(c, "limit", 20),       // Shared function (capitalized)
		Offset:    ParseIntQuery(c, "offset", 0),       // Shared function (capitalized)
		MinRating: ParseFloatQuery(c, "min_rating", 0), // Shared function (capitalized)
	}

	if verifiedStr := c.Query("is_verified"); verifiedStr != "" {
		verified := verifiedStr == "true"
		filters.IsVerified = &verified
	}

	// Get barbers
	barbers, err := h.barberService.GetAllBarbers(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch barbers",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    barbers,
		Meta: map[string]interface{}{
			"count":  len(barbers),
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
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
	// Single line replaces 7 lines
	id, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	barber, err := h.barberService.GetBarberByID(c.Request.Context(), id)
	if err != nil {
		// Single line handles common errors
		if !HandleRepositoryError(c, err, "Barber") {
			RespondInternalError(c, "fetch barber", err)
		}
		return
	}

	// Single line replaces 4 lines
	RespondSuccess(c, barber)
}

func (h *BarberHandler) CreateBarber(c *gin.Context) {
	var req services.CreateBarberRequest

	// Single line replaces 5 lines
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}

	barber, err := h.barberService.CreateBarber(c.Request.Context(), req)
	if err != nil {
		RespondInternalError(c, "create barber", err)
		return
	}

	// Single line replaces 5 lines
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
	if err != nil {
		if err == repository.ErrBarberNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Barber not found",
				Message: "No barber found with the given UUID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch barber",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    barber,
	})
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
// @Router /api/v1/barbers [post]


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

	var req services.UpdateBarberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	barber, err := h.barberService.UpdateBarber(c.Request.Context(), id, req)
	if err != nil {
		if err == repository.ErrBarberNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Barber not found",
				Message: "No barber found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to update barber",
			Message: err.Error(),
		})
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
		if err == repository.ErrBarberNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Barber not found",
				Message: "No barber found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to delete barber",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Barber deleted successfully",
	})
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
func (h *BarberHandler) UpdateBarberStatus(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "barber")
	if !ok {
		return
	}

	var req StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	if err := h.barberService.UpdateBarberStatus(c.Request.Context(), id, req.Status); err != nil {
		if err == repository.ErrBarberNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Barber not found",
				Message: "No barber found with the given ID",
			})
			return
		}
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Failed to update status",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Status updated successfully",
	})
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
	})
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
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Search failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    barbers,
		Meta: map[string]interface{}{
			"query": query,
			"count": len(barbers),
		},
	})
}

// Request types (handler-specific)
type StatusUpdateRequest struct {
	Status string `json:"status" binding:"required"`
}
