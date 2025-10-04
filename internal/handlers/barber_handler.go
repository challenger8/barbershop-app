// internal/handlers/barber_handler.go
package handlers

import (
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"
	"net/http"
	"strconv"

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
	// Parse filters from query parameters
	filters := repository.BarberFilters{
		Status:    c.Query("status"),
		City:      c.Query("city"),
		State:     c.Query("state"),
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		Limit:     parseIntQuery(c, "limit", 20),
		Offset:    parseIntQuery(c, "offset", 0),
		MinRating: parseFloatQuery(c, "min_rating", 0),
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid barber ID",
			Message: "Barber ID must be a number",
		})
		return
	}

	barber, err := h.barberService.GetBarberByID(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrBarberNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Barber not found",
				Message: "No barber found with the given ID",
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
func (h *BarberHandler) CreateBarber(c *gin.Context) {
	var req services.CreateBarberRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	barber, err := h.barberService.CreateBarber(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to create barber",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    barber,
		Message: "Barber created successfully",
	})
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid barber ID",
			Message: "Barber ID must be a number",
		})
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid barber ID",
			Message: "Barber ID must be a number",
		})
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid barber ID",
			Message: "Barber ID must be a number",
		})
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid barber ID",
			Message: "Barber ID must be a number",
		})
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

// Helper functions
func parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func parseFloatQuery(c *gin.Context, key string, defaultValue float64) float64 {
	if value := c.Query(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// Request types (handler-specific)
type StatusUpdateRequest struct {
	Status string `json:"status" binding:"required"`
}
