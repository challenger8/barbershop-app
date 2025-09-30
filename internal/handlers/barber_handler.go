// internal/handlers/barber_handler.go
package handlers

import (
	"barber-booking-system/internal/models"
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
func (h *BarberHandler) GetAllBarbers(c *gin.Context) {
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

	barbers, err := h.barberService.GetAllBarbers(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
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
func (h *BarberHandler) GetBarber(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid barber ID",
			Message: "Barber ID must be a number",
		})
		return
	}

	// ✅ Fixed: Use GetBarber instead of GetBarberByID
	barber, err := h.barberService.GetBarber(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrBarberNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Barber not found",
				Message: "No barber found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
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
func (h *BarberHandler) GetBarberByUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	barber, err := h.barberService.GetBarberByUUID(c.Request.Context(), uuid)
	if err != nil {
		if err == repository.ErrBarberNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Barber not found",
				Message: "No barber found with the given UUID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
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
// CreateBarber godoc
// CreateBarber godoc
func (h *BarberHandler) CreateBarber(c *gin.Context) {
	var req services.CreateBarberRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// ✅ Convert request to barber model (only with existing fields)
	barber := &models.Barber{
		Address: req.Address,
		City:    req.City,
		State:   req.State,
		// Add any other fields that actually exist in your CreateBarberRequest
	}

	err := h.barberService.CreateBarber(c.Request.Context(), barber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
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
func (h *BarberHandler) UpdateBarber(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid barber ID",
			Message: "Barber ID must be a number",
		})
		return
	}

	var req services.UpdateBarberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	barber, err := h.barberService.UpdateBarber(c.Request.Context(), id, &req)
	if err != nil {
		if err == repository.ErrBarberNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Barber not found",
				Message: "No barber found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
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
func (h *BarberHandler) DeleteBarber(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid barber ID",
			Message: "Barber ID must be a number",
		})
		return
	}

	if err := h.barberService.DeleteBarber(c.Request.Context(), id); err != nil {
		if err == repository.ErrBarberNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Barber not found",
				Message: "No barber found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
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
func (h *BarberHandler) UpdateBarberStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid barber ID",
			Message: "Barber ID must be a number",
		})
		return
	}

	var req StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	if err := h.barberService.UpdateBarberStatus(c.Request.Context(), id, req.Status); err != nil {
		if err == repository.ErrBarberNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Barber not found",
				Message: "No barber found with the given ID",
			})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
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
func (h *BarberHandler) GetBarberStatistics(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid barber ID",
			Message: "Barber ID must be a number",
		})
		return
	}

	stats, err := h.barberService.GetBarberStatistics(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
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
func (h *BarberHandler) SearchBarbers(c *gin.Context) {
	query := c.Query("q")
	filters := repository.BarberFilters{
		City:   c.Query("city"),
		State:  c.Query("state"),
		Status: "active",
		Limit:  50,
	}

	barbers, err := h.barberService.SearchBarbers(c.Request.Context(), query, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
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

// Response types
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type StatusUpdateRequest struct {
	Status string `json:"status" binding:"required"`
}
