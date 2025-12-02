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
		CategoryID:   ParseIntQuery(c, "category_id", 0),
		ServiceType:  c.Query("service_type"),
		MinRating:    ParseFloatQuery(c, "min_rating", 0),
		Complexity:   ParseIntQuery(c, "complexity", 0),
		TargetGender: c.Query("target_gender"),
		Search:       c.Query("search"),
		SortBy:       c.Query("sort_by"),
		Limit:        ParseIntQuery(c, "limit", 20),
		Offset:       ParseIntQuery(c, "offset", 0),
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
		RespondInternalError(c, "fetch services", err)
		return
	}

	RespondSuccessWithMeta(c, servicesList, map[string]interface{}{
		"count":  len(servicesList),
		"limit":  filters.Limit,
		"offset": filters.Offset,
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
			RespondNotFound(c, "Service")
			return
		}
		RespondInternalError(c, "fetch service", err)
		return
	}

	RespondSuccess(c, service)
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
			RespondNotFound(c, "Service")
			return
		}
		RespondInternalError(c, "fetch service", err)
		return
	}

	RespondSuccess(c, service)
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
	req, ok := BindJSON[services.CreateServiceRequest](c)
	if !ok {
		return
	}

	// Get created_by from auth context
	if userID, exists := middleware.GetUserID(c); exists {
		req.CreatedBy = &userID
	}

	service, err := h.serviceService.CreateService(c.Request.Context(), *req)
	if err != nil {
		RespondInternalError(c, "create service", err)
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

	req, ok := BindJSON[services.UpdateServiceRequest](c)
	if !ok {
		return
	}

	// Get last_modified_by from auth context
	if userID, exists := middleware.GetUserID(c); exists {
		req.LastModifiedBy = &userID
	}

	service, err := h.serviceService.UpdateService(c.Request.Context(), id, *req)
	if err != nil {
		if err == repository.ErrServiceNotFound {
			RespondNotFound(c, "Service")
			return
		}
		RespondInternalError(c, "update service", err)
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
			RespondNotFound(c, "Service")
			return
		}
		RespondInternalError(c, "delete service", err)
		return
	}

	RespondSuccessWithMessage(c, "Service deleted successfully")
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
		CategoryID: ParseIntQuery(c, "category_id", 0),
		Limit:      50,
	}
	isActive := true
	filters.IsActive = &isActive

	servicesList, err := h.serviceService.SearchServices(c.Request.Context(), query, filters)
	if err != nil {
		RespondInternalError(c, "search services", err)
		return
	}

	RespondSuccessWithMeta(c, servicesList, map[string]interface{}{
		"query": query,
		"count": len(servicesList),
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
		RespondInternalError(c, "fetch categories", err)
		return
	}

	RespondSuccessWithMeta(c, categories, map[string]interface{}{
		"count": len(categories),
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
			RespondNotFound(c, "Category")
			return
		}
		RespondInternalError(c, "fetch category", err)
		return
	}

	RespondSuccess(c, category)
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
	req, ok := BindJSON[services.CreateCategoryRequest](c)
	if !ok {
		return
	}

	category, err := h.serviceService.CreateCategory(c.Request.Context(), *req)
	if err != nil {
		RespondInternalError(c, "create category", err)
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

	req, ok := BindJSON[services.UpdateCategoryRequest](c)
	if !ok {
		return
	}

	category, err := h.serviceService.UpdateCategory(c.Request.Context(), id, *req)
	if err != nil {
		if err == repository.ErrCategoryNotFound {
			RespondNotFound(c, "Category")
			return
		}
		RespondInternalError(c, "update category", err)
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
			RespondNotFound(c, "Category")
			return
		}
		RespondInternalError(c, "delete category", err)
		return
	}

	RespondSuccessWithMessage(c, "Category deleted successfully")
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
		RespondInternalError(c, "fetch barber services", err)
		return
	}

	RespondSuccessWithMeta(c, barberServices, map[string]interface{}{
		"barber_id": barberID,
		"count":     len(barberServices),
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
		RespondInternalError(c, "fetch barbers offering service", err)
		return
	}

	RespondSuccessWithMeta(c, barberServices, map[string]interface{}{
		"service_id": serviceID,
		"count":      len(barberServices),
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
	req, ok := BindJSON[services.CreateBarberServiceRequest](c)
	if !ok {
		return
	}

	barberService, err := h.serviceService.AddServiceToBarber(c.Request.Context(), *req)
	if err != nil {
		RespondInternalError(c, "add service to barber", err)
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

	req, ok := BindJSON[services.UpdateBarberServiceRequest](c)
	if !ok {
		return
	}

	barberService, err := h.serviceService.UpdateBarberService(c.Request.Context(), id, *req)
	if err != nil {
		if err == repository.ErrBarberServiceNotFound {
			RespondNotFound(c, "Barber service")
			return
		}
		RespondInternalError(c, "update barber service", err)
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
			RespondNotFound(c, "Barber service")
			return
		}
		RespondInternalError(c, "remove service from barber", err)
		return
	}

	RespondSuccessWithMessage(c, "Service removed from barber successfully")
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
			RespondNotFound(c, "Barber service")
			return
		}
		RespondInternalError(c, "fetch barber service", err)
		return
	}

	RespondSuccess(c, barberService)
}
