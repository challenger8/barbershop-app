// internal/handlers/service_handler.go
package handlers

import (
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/repository"
	"barber-booking-system/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServiceHandler handles HTTP requests for services
type ServiceHandler struct {
	serviceService *services.ServiceService
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(serviceService *services.ServiceService) *ServiceHandler {
	return &ServiceHandler{
		serviceService: serviceService,
	}
}

// ==================== Service Endpoints ====================

// GetAllServices godoc
// @Summary Get all services
// @Description Get list of all services from the catalog with optional filters
// @Tags services
// @Accept json
// @Produce json
// @Param category_id query int false "Filter by category ID"
// @Param service_type query string false "Filter by service type (haircut, styling, treatment, grooming)"
// @Param is_active query bool false "Filter by active status"
// @Param min_rating query number false "Minimum rating"
// @Param complexity query int false "Filter by complexity (1-5)"
// @Param target_gender query string false "Filter by target gender (male, female, all)"
// @Param search query string false "Search term"
// @Param sort_by query string false "Sort by field (name, popularity, rating, duration, complexity)"
// @Param limit query int false "Number of results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services [get]
func (h *ServiceHandler) GetAllServices(c *gin.Context) {
	filters := repository.ServiceFilters{
		CategoryID:   ParseIntQuery(c, "category_id", 0), // Shared function
		ServiceType:  c.Query("service_type"),
		MinRating:    ParseFloatQuery(c, "min_rating", 0), // Shared function
		Complexity:   ParseIntQuery(c, "complexity", 0),   // Shared function
		TargetGender: c.Query("target_gender"),
		Search:       c.Query("search"),
		SortBy:       c.Query("sort_by"),
		Limit:        ParseIntQuery(c, "limit", 20), // Shared function
		Offset:       ParseIntQuery(c, "offset", 0), // Shared function
	}

	if activeStr := c.Query("is_active"); activeStr != "" {
		active := activeStr == "true"
		filters.IsActive = &active
	}

	if approvedStr := c.Query("is_approved"); approvedStr != "" {
		approved := approvedStr == "true"
		filters.IsApproved = &approved
	}

	servicesList, err := h.serviceService.GetAllServices(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch services",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    servicesList,
		Meta: map[string]interface{}{
			"count":  len(servicesList),
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
	})
}

// GetService godoc
// @Summary Get service by ID
// @Description Get detailed information about a specific service
// @Tags services
// @Accept json
// @Produce json
// @Param id path int true "Service ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/{id} [get]
func (h *ServiceHandler) GetService(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "service")
	if !ok {
		return
	}

	service, err := h.serviceService.GetServiceByID(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrServiceNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Service not found",
				Message: "No service found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch service",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    service,
	})
}

// GetServiceBySlug godoc
// @Summary Get service by slug
// @Description Get detailed information about a specific service by its URL slug
// @Tags services
// @Accept json
// @Produce json
// @Param slug path string true "Service slug"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/slug/{slug} [get]
func (h *ServiceHandler) GetServiceBySlug(c *gin.Context) {
	slug := c.Param("slug")

	service, err := h.serviceService.GetServiceBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == repository.ErrServiceNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Service not found",
				Message: "No service found with the given slug",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch service",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    service,
	})
}

// CreateService godoc
// @Summary Create new service
// @Description Create a new service in the catalog (admin only)
// @Tags services
// @Accept json
// @Produce json
// @Param service body services.CreateServiceRequest true "Service data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services [post]
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req services.CreateServiceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Get created_by from auth context
	if userID, exists := middleware.GetUserID(c); exists {
		req.CreatedBy = &userID
	}

	service, err := h.serviceService.CreateService(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to create service",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    service,
		Message: "Service created successfully",
	})
}

// UpdateService godoc
// @Summary Update service
// @Description Update service information (admin only)
// @Tags services
// @Accept json
// @Produce json
// @Param id path int true "Service ID"
// @Param service body services.UpdateServiceRequest true "Updated service data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/{id} [put]
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "service")
	if !ok {
		return
	}

	var req services.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Get last_modified_by from auth context
	if userID, exists := middleware.GetUserID(c); exists {
		req.LastModifiedBy = &userID
	}

	service, err := h.serviceService.UpdateService(c.Request.Context(), id, req)
	if err != nil {
		if err == repository.ErrServiceNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Service not found",
				Message: "No service found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to update service",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    service,
		Message: "Service updated successfully",
	})
}

// DeleteService godoc
// @Summary Delete service
// @Description Soft delete a service (admin only)
// @Tags services
// @Accept json
// @Produce json
// @Param id path int true "Service ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/{id} [delete]
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "service")
	if !ok {
		return
	}

	if err := h.serviceService.DeleteService(c.Request.Context(), id); err != nil {
		if err == repository.ErrServiceNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Service not found",
				Message: "No service found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to delete service",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Service deleted successfully",
	})
}

// SearchServices godoc
// @Summary Search services
// @Description Search services by query string
// @Tags services
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param category_id query int false "Filter by category"
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/search [get]
func (h *ServiceHandler) SearchServices(c *gin.Context) {
	query := c.Query("q")
	filters := repository.ServiceFilters{
		CategoryID: ParseIntQuery(c, "category_id", 0), // Shared function
		Limit:      50,
	}
	isActive := true
	filters.IsActive = &isActive

	servicesList, err := h.serviceService.SearchServices(c.Request.Context(), query, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Search failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    servicesList,
		Meta: map[string]interface{}{
			"query": query,
			"count": len(servicesList),
		},
	})
}

// ==================== Category Endpoints ====================

// GetAllCategories godoc
// @Summary Get all service categories
// @Description Get list of all service categories
// @Tags services
// @Accept json
// @Produce json
// @Param active_only query bool false "Only return active categories" default(true)
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/categories [get]
func (h *ServiceHandler) GetAllCategories(c *gin.Context) {
	activeOnly := c.Query("active_only") != "false"

	categories, err := h.serviceService.GetAllCategories(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch categories",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    categories,
		Meta: map[string]interface{}{
			"count": len(categories),
		},
	})
}

// GetCategory godoc
// @Summary Get category by ID
// @Description Get detailed information about a specific category
// @Tags services
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/categories/{id} [get]
func (h *ServiceHandler) GetCategory(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "category")
	if !ok {
		return
	}

	category, err := h.serviceService.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrCategoryNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Category not found",
				Message: "No category found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch category",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    category,
	})
}

// CreateCategory godoc
// @Summary Create new category
// @Description Create a new service category (admin only)
// @Tags services
// @Accept json
// @Produce json
// @Param category body services.CreateCategoryRequest true "Category data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/categories [post]
func (h *ServiceHandler) CreateCategory(c *gin.Context) {
	var req services.CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	category, err := h.serviceService.CreateCategory(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to create category",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    category,
		Message: "Category created successfully",
	})
}

// UpdateCategory godoc
// @Summary Update category
// @Description Update category information (admin only)
// @Tags services
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body services.UpdateCategoryRequest true "Updated category data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/categories/{id} [put]
func (h *ServiceHandler) UpdateCategory(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "category")
	if !ok {
		return
	}

	var req services.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	category, err := h.serviceService.UpdateCategory(c.Request.Context(), id, req)
	if err != nil {
		if err == repository.ErrCategoryNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Category not found",
				Message: "No category found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to update category",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    category,
		Message: "Category updated successfully",
	})
}

// DeleteCategory godoc
// @Summary Delete category
// @Description Soft delete a category (admin only)
// @Tags services
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/categories/{id} [delete]
func (h *ServiceHandler) DeleteCategory(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "category")
	if !ok {
		return
	}

	if err := h.serviceService.DeleteCategory(c.Request.Context(), id); err != nil {
		if err == repository.ErrCategoryNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Category not found",
				Message: "No category found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to delete category",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Category deleted successfully",
	})
}

// ==================== Barber Service Endpoints ====================

// GetBarberServices godoc
// @Summary Get barber's services
// @Description Get all services offered by a specific barber
// @Tags services
// @Accept json
// @Produce json
// @Param barber_id path int true "Barber ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barbers/{barber_id}/services [get]
func (h *ServiceHandler) GetBarberServices(c *gin.Context) {
	barberID, ok := RequireIntParam(c, "barber_id", "barber")
	if !ok {
		return
	}

	barberServices, err := h.serviceService.GetBarberServices(c.Request.Context(), barberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch barber services",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    barberServices,
		Meta: map[string]interface{}{
			"barber_id": barberID,
			"count":     len(barberServices),
		},
	})
}

// GetBarbersOfferingService godoc
// @Summary Get barbers offering a service
// @Description Get all barbers that offer a specific service
// @Tags services
// @Accept json
// @Produce json
// @Param service_id path int true "Service ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/services/{service_id}/barbers [get]
func (h *ServiceHandler) GetBarbersOfferingService(c *gin.Context) {
	serviceID, ok := RequireIntParam(c, "service_id", "service")
	if !ok {
		return
	}

	barberServices, err := h.serviceService.GetBarbersOfferingService(c.Request.Context(), serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch barbers",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    barberServices,
		Meta: map[string]interface{}{
			"service_id": serviceID,
			"count":      len(barberServices),
		},
	})
}

// AddServiceToBarber godoc
// @Summary Add service to barber
// @Description Add a service to a barber's offerings (protected)
// @Tags services
// @Accept json
// @Produce json
// @Param barber_service body services.CreateBarberServiceRequest true "Barber service data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barber-services [post]
func (h *ServiceHandler) AddServiceToBarber(c *gin.Context) {
	var req services.CreateBarberServiceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	barberService, err := h.serviceService.AddServiceToBarber(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to add service to barber",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    barberService,
		Message: "Service added to barber successfully",
	})
}

// UpdateBarberService godoc
// @Summary Update barber service
// @Description Update a barber's service offering (protected)
// @Tags services
// @Accept json
// @Produce json
// @Param id path int true "Barber service ID"
// @Param barber_service body services.UpdateBarberServiceRequest true "Updated barber service data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barber-services/{id} [put]
func (h *ServiceHandler) UpdateBarberService(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "barber service")
	if !ok {
		return
	}

	var req services.UpdateBarberServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	barberService, err := h.serviceService.UpdateBarberService(c.Request.Context(), id, req)
	if err != nil {
		if err == repository.ErrBarberServiceNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Barber service not found",
				Message: "No barber service found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to update barber service",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    barberService,
		Message: "Barber service updated successfully",
	})
}

// RemoveServiceFromBarber godoc
// @Summary Remove service from barber
// @Description Remove a service from a barber's offerings (protected)
// @Tags services
// @Accept json
// @Produce json
// @Param id path int true "Barber service ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barber-services/{id} [delete]
func (h *ServiceHandler) RemoveServiceFromBarber(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "barber service")
	if !ok {
		return
	}

	if err := h.serviceService.RemoveServiceFromBarber(c.Request.Context(), id); err != nil {
		if err == repository.ErrBarberServiceNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Barber service not found",
				Message: "No barber service found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to remove service from barber",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Service removed from barber successfully",
	})
}

// GetBarberServiceByID godoc
// @Summary Get barber service by ID
// @Description Get detailed information about a specific barber service
// @Tags services
// @Accept json
// @Produce json
// @Param id path int true "Barber service ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/barber-services/{id} [get]
func (h *ServiceHandler) GetBarberServiceByID(c *gin.Context) {
	id, ok := RequireIntParam(c, "id", "barber service")
	if !ok {
		return
	}

	barberService, err := h.serviceService.GetBarberServiceByID(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrBarberServiceNotFound {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "Barber service not found",
				Message: "No barber service found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch barber service",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    barberService,
	})
}
