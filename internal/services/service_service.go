// internal/services/service_service.go
package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/models"
	"barber-booking-system/internal/repository"

	"github.com/google/uuid"
)

// ServiceService handles service business logic
type ServiceService struct {
	repo  *repository.ServiceRepository
	cache *cache.CacheService
}

// NewServiceService creates a new service service
func NewServiceService(repo *repository.ServiceRepository, cache *cache.CacheService) *ServiceService {
	return &ServiceService{
		repo:  repo,
		cache: cache,
	}
}

// ==================== Service Operations ====================

// GetAllServices retrieves all services with filters
func (s *ServiceService) GetAllServices(ctx context.Context, filters repository.ServiceFilters) ([]models.Service, error) {
	return s.repo.FindAll(ctx, filters)
}

// GetServiceByID retrieves a service by ID with caching
func (s *ServiceService) GetServiceByID(ctx context.Context, id int) (*models.Service, error) {
	// Try cache first
	if s.cache != nil {
		var cachedService models.Service
		cacheKey := fmt.Sprintf("service:%d", id)
		err := s.cache.Get(ctx, cacheKey, &cachedService)
		if err == nil {
			return &cachedService, nil
		}
	}

	// Fetch from database
	service, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if s.cache != nil {
		cacheKey := fmt.Sprintf("service:%d", id)
		_ = s.cache.Set(ctx, cacheKey, service)
	}

	return service, nil
}

// GetServiceByUUID retrieves a service by UUID
func (s *ServiceService) GetServiceByUUID(ctx context.Context, uuid string) (*models.Service, error) {
	return s.repo.FindByUUID(ctx, uuid)
}

// GetServiceBySlug retrieves a service by slug
func (s *ServiceService) GetServiceBySlug(ctx context.Context, slug string) (*models.Service, error) {
	return s.repo.FindBySlug(ctx, slug)
}

// CreateService creates a new service
func (s *ServiceService) CreateService(ctx context.Context, req CreateServiceRequest) (*models.Service, error) {
	// Validate request
	if err := s.validateCreateServiceRequest(req); err != nil {
		return nil, err
	}

	// Generate slug if not provided
	slug := req.Slug
	if slug == "" {
		slug = s.generateSlug(req.Name)
	}

	// Build service model
	service := &models.Service{
		UUID:                   uuid.New().String(),
		Name:                   req.Name,
		Slug:                   slug,
		ShortDescription:       req.ShortDescription,
		DetailedDescription:    req.DetailedDescription,
		CategoryID:             req.CategoryID,
		ServiceType:            req.ServiceType,
		Complexity:             req.Complexity,
		SkillLevelRequired:     req.SkillLevelRequired,
		DefaultDurationMin:     req.DefaultDurationMin,
		DefaultDurationMax:     req.DefaultDurationMax,
		SuggestedPriceMin:      req.SuggestedPriceMin,
		SuggestedPriceMax:      req.SuggestedPriceMax,
		Currency:               req.Currency,
		TargetGender:           req.TargetGender,
		TargetAgeMin:           req.TargetAgeMin,
		TargetAgeMax:           req.TargetAgeMax,
		HairTypes:              req.HairTypes,
		RequiresConsultation:   req.RequiresConsultation,
		RequiredTools:          req.RequiredTools,
		RequiredProducts:       req.RequiredProducts,
		RequiredCertifications: req.RequiredCertifications,
		AllergenWarnings:       req.AllergenWarnings,
		HealthPrecautions:      req.HealthPrecautions,
		RequiresHealthCheck:    req.RequiresHealthCheck,
		ImageURL:               req.ImageURL,
		GalleryImages:          req.GalleryImages,
		VideoURL:               req.VideoURL,
		Tags:                   req.Tags,
		SearchKeywords:         req.SearchKeywords,
		MetaDescription:        req.MetaDescription,
		HasVariations:          req.HasVariations,
		AllowsAddOns:           req.AllowsAddOns,
		IsActive:               true,
		IsApproved:             false, // Requires admin approval
		CreatedBy:              req.CreatedBy,
	}

	// Set defaults
	if service.Currency == "" {
		service.Currency = "USD"
	}
	if service.TargetGender == "" {
		service.TargetGender = "all"
	}
	if service.SkillLevelRequired == "" {
		service.SkillLevelRequired = "intermediate"
	}

	// Create in database
	if err := s.repo.Create(ctx, service); err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return service, nil
}

// UpdateService updates a service
func (s *ServiceService) UpdateService(ctx context.Context, id int, req UpdateServiceRequest) (*models.Service, error) {
	// Fetch existing service
	service, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		service.Name = *req.Name
	}
	if req.Slug != nil {
		service.Slug = *req.Slug
	}
	if req.ShortDescription != nil {
		service.ShortDescription = *req.ShortDescription
	}
	if req.DetailedDescription != nil {
		service.DetailedDescription = req.DetailedDescription
	}
	if req.CategoryID != nil {
		service.CategoryID = *req.CategoryID
	}
	if req.ServiceType != nil {
		service.ServiceType = *req.ServiceType
	}
	if req.Complexity != nil {
		service.Complexity = *req.Complexity
	}
	if req.SkillLevelRequired != nil {
		service.SkillLevelRequired = *req.SkillLevelRequired
	}
	if req.DefaultDurationMin != nil {
		service.DefaultDurationMin = *req.DefaultDurationMin
	}
	if req.DefaultDurationMax != nil {
		service.DefaultDurationMax = req.DefaultDurationMax
	}
	if req.SuggestedPriceMin != nil {
		service.SuggestedPriceMin = req.SuggestedPriceMin
	}
	if req.SuggestedPriceMax != nil {
		service.SuggestedPriceMax = req.SuggestedPriceMax
	}
	if req.Currency != nil {
		service.Currency = *req.Currency
	}
	if req.TargetGender != nil {
		service.TargetGender = *req.TargetGender
	}
	if req.HairTypes != nil {
		service.HairTypes = req.HairTypes
	}
	if req.RequiresConsultation != nil {
		service.RequiresConsultation = *req.RequiresConsultation
	}
	if req.RequiredTools != nil {
		service.RequiredTools = req.RequiredTools
	}
	if req.RequiredProducts != nil {
		service.RequiredProducts = req.RequiredProducts
	}
	if req.ImageURL != nil {
		service.ImageURL = req.ImageURL
	}
	if req.GalleryImages != nil {
		service.GalleryImages = req.GalleryImages
	}
	if req.Tags != nil {
		service.Tags = req.Tags
	}
	if req.SearchKeywords != nil {
		service.SearchKeywords = req.SearchKeywords
	}
	if req.HasVariations != nil {
		service.HasVariations = *req.HasVariations
	}
	if req.AllowsAddOns != nil {
		service.AllowsAddOns = *req.AllowsAddOns
	}
	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}
	if req.LastModifiedBy != nil {
		service.LastModifiedBy = req.LastModifiedBy
	}

	// Update in database
	if err := s.repo.Update(ctx, service); err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("service:%d", id)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	return service, nil
}

// DeleteService soft deletes a service
func (s *ServiceService) DeleteService(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("service:%d", id)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// ApproveService approves a service for public listing
func (s *ServiceService) ApproveService(ctx context.Context, id int, approvedBy *int, notes *string) error {
	service, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	service.IsApproved = true
	service.ApprovalNotes = notes
	service.LastModifiedBy = approvedBy

	return s.repo.Update(ctx, service)
}

// SearchServices searches services by query
func (s *ServiceService) SearchServices(ctx context.Context, query string, filters repository.ServiceFilters) ([]models.Service, error) {
	filters.Search = query
	return s.repo.FindAll(ctx, filters)
}

// ==================== Category Operations ====================

// GetAllCategories retrieves all service categories
func (s *ServiceService) GetAllCategories(ctx context.Context, activeOnly bool) ([]models.ServiceCategory, error) {
	return s.repo.FindAllCategories(ctx, activeOnly)
}

// GetCategoryByID retrieves a category by ID
func (s *ServiceService) GetCategoryByID(ctx context.Context, id int) (*models.ServiceCategory, error) {
	return s.repo.FindCategoryByID(ctx, id)
}

// CreateCategory creates a new service category
func (s *ServiceService) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*models.ServiceCategory, error) {
	// Generate slug if not provided
	slug := req.Slug
	if slug == "" {
		slug = s.generateSlug(req.Name)
	}

	category := &models.ServiceCategory{
		Name:             req.Name,
		Slug:             slug,
		Description:      req.Description,
		ParentCategoryID: req.ParentCategoryID,
		Level:            req.Level,
		CategoryPath:     req.CategoryPath,
		IconURL:          req.IconURL,
		ColorHex:         req.ColorHex,
		ImageURL:         req.ImageURL,
		SortOrder:        req.SortOrder,
		IsActive:         true,
		IsFeatured:       req.IsFeatured,
		MetaTitle:        req.MetaTitle,
		MetaDescription:  req.MetaDescription,
		Keywords:         req.Keywords,
	}

	// Set defaults
	if category.Level == 0 {
		if category.ParentCategoryID != nil {
			category.Level = 2
		} else {
			category.Level = 1
		}
	}
	if category.CategoryPath == "" {
		category.CategoryPath = slug
	}

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// UpdateCategory updates a service category
func (s *ServiceService) UpdateCategory(ctx context.Context, id int, req UpdateCategoryRequest) (*models.ServiceCategory, error) {
	category, err := s.repo.FindCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Slug != nil {
		category.Slug = *req.Slug
	}
	if req.Description != nil {
		category.Description = req.Description
	}
	if req.IconURL != nil {
		category.IconURL = req.IconURL
	}
	if req.ColorHex != nil {
		category.ColorHex = req.ColorHex
	}
	if req.ImageURL != nil {
		category.ImageURL = req.ImageURL
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if req.IsFeatured != nil {
		category.IsFeatured = *req.IsFeatured
	}

	if err := s.repo.UpdateCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

// DeleteCategory soft deletes a service category
func (s *ServiceService) DeleteCategory(ctx context.Context, id int) error {
	return s.repo.DeleteCategory(ctx, id)
}

// ==================== Barber Service Operations ====================

// GetBarberServices retrieves all services offered by a barber
func (s *ServiceService) GetBarberServices(ctx context.Context, barberID int) ([]models.BarberService, error) {
	return s.repo.GetServicesByBarberID(ctx, barberID)
}

// GetBarberServiceByID retrieves a barber service by ID
func (s *ServiceService) GetBarberServiceByID(ctx context.Context, id int) (*models.BarberService, error) {
	return s.repo.FindBarberServiceByID(ctx, id)
}

// GetBarbersOfferingService retrieves all barbers offering a specific service
func (s *ServiceService) GetBarbersOfferingService(ctx context.Context, serviceID int) ([]models.BarberService, error) {
	return s.repo.GetBarbersByServiceID(ctx, serviceID)
}

// AddServiceToBarber adds a service to a barber's offerings
func (s *ServiceService) AddServiceToBarber(ctx context.Context, req CreateBarberServiceRequest) (*models.BarberService, error) {
	// Validate request
	if err := s.validateBarberServiceRequest(req); err != nil {
		return nil, err
	}

	barberService := &models.BarberService{
		BarberID:               req.BarberID,
		ServiceID:              req.ServiceID,
		CustomName:             req.CustomName,
		CustomDescription:      req.CustomDescription,
		Price:                  req.Price,
		MaxPrice:               req.MaxPrice,
		Currency:               req.Currency,
		DiscountPrice:          req.DiscountPrice,
		DiscountValidUntil:     req.DiscountValidUntil,
		EstimatedDurationMin:   req.EstimatedDurationMin,
		EstimatedDurationMax:   req.EstimatedDurationMax,
		BufferTimeMinutes:      req.BufferTimeMinutes,
		AdvanceNoticeHours:     req.AdvanceNoticeHours,
		MaxAdvanceBookingDays:  req.MaxAdvanceBookingDays,
		AvailableDays:          req.AvailableDays,
		AvailableTimeSlots:     req.AvailableTimeSlots,
		RequiresConsultation:   req.RequiresConsultation,
		ConsultationDuration:   req.ConsultationDuration,
		PreServiceInstructions: req.PreServiceInstructions,
		PostServiceCare:        req.PostServiceCare,
		MinCustomerAge:         req.MinCustomerAge,
		MaxCustomerAge:         req.MaxCustomerAge,
		IsSeasonal:             req.IsSeasonal,
		SeasonalStartMonth:     req.SeasonalStartMonth,
		SeasonalEndMonth:       req.SeasonalEndMonth,
		PortfolioImages:        req.PortfolioImages,
		BeforeAfterImages:      req.BeforeAfterImages,
		IsPromotional:          req.IsPromotional,
		PromotionalText:        req.PromotionalText,
		PromotionStartDate:     req.PromotionStartDate,
		PromotionEndDate:       req.PromotionEndDate,
		IsFeatured:             req.IsFeatured,
		DisplayOrder:           req.DisplayOrder,
		ServiceNote:            req.ServiceNote,
		IsActive:               true,
	}

	// Set defaults
	if barberService.Currency == "" {
		barberService.Currency = "USD"
	}

	if err := s.repo.CreateBarberService(ctx, barberService); err != nil {
		return nil, fmt.Errorf("failed to add service to barber: %w", err)
	}

	return barberService, nil
}

// UpdateBarberService updates a barber's service offering
func (s *ServiceService) UpdateBarberService(ctx context.Context, id int, req UpdateBarberServiceRequest) (*models.BarberService, error) {
	barberService, err := s.repo.FindBarberServiceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.CustomName != nil {
		barberService.CustomName = req.CustomName
	}
	if req.CustomDescription != nil {
		barberService.CustomDescription = req.CustomDescription
	}
	if req.Price != nil {
		barberService.Price = *req.Price
	}
	if req.MaxPrice != nil {
		barberService.MaxPrice = req.MaxPrice
	}
	if req.Currency != nil {
		barberService.Currency = *req.Currency
	}
	if req.DiscountPrice != nil {
		barberService.DiscountPrice = req.DiscountPrice
	}
	if req.DiscountValidUntil != nil {
		barberService.DiscountValidUntil = req.DiscountValidUntil
	}
	if req.EstimatedDurationMin != nil {
		barberService.EstimatedDurationMin = *req.EstimatedDurationMin
	}
	if req.EstimatedDurationMax != nil {
		barberService.EstimatedDurationMax = req.EstimatedDurationMax
	}
	if req.BufferTimeMinutes != nil {
		barberService.BufferTimeMinutes = *req.BufferTimeMinutes
	}
	if req.AdvanceNoticeHours != nil {
		barberService.AdvanceNoticeHours = *req.AdvanceNoticeHours
	}
	if req.MaxAdvanceBookingDays != nil {
		barberService.MaxAdvanceBookingDays = req.MaxAdvanceBookingDays
	}
	if req.AvailableDays != nil {
		barberService.AvailableDays = req.AvailableDays
	}
	if req.PortfolioImages != nil {
		barberService.PortfolioImages = req.PortfolioImages
	}
	if req.IsPromotional != nil {
		barberService.IsPromotional = *req.IsPromotional
	}
	if req.PromotionalText != nil {
		barberService.PromotionalText = req.PromotionalText
	}
	if req.IsFeatured != nil {
		barberService.IsFeatured = *req.IsFeatured
	}
	if req.DisplayOrder != nil {
		barberService.DisplayOrder = *req.DisplayOrder
	}
	if req.ServiceNote != nil {
		barberService.ServiceNote = req.ServiceNote
	}
	if req.IsActive != nil {
		barberService.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateBarberService(ctx, barberService); err != nil {
		return nil, fmt.Errorf("failed to update barber service: %w", err)
	}

	return barberService, nil
}

// RemoveServiceFromBarber removes a service from a barber's offerings
func (s *ServiceService) RemoveServiceFromBarber(ctx context.Context, id int) error {
	return s.repo.DeleteBarberService(ctx, id)
}

// ==================== Helper Methods ====================

func (s *ServiceService) validateCreateServiceRequest(req CreateServiceRequest) error {
	var errors []string

	if strings.TrimSpace(req.Name) == "" {
		errors = append(errors, "name is required")
	}
	if strings.TrimSpace(req.ShortDescription) == "" {
		errors = append(errors, "short_description is required")
	}
	if req.CategoryID <= 0 {
		errors = append(errors, "valid category_id is required")
	}
	if req.DefaultDurationMin <= 0 {
		errors = append(errors, "default_duration_min must be positive")
	}
	if req.Complexity < 1 || req.Complexity > 5 {
		errors = append(errors, "complexity must be between 1 and 5")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

func (s *ServiceService) validateBarberServiceRequest(req CreateBarberServiceRequest) error {
	var errors []string

	if req.BarberID <= 0 {
		errors = append(errors, "valid barber_id is required")
	}
	if req.ServiceID <= 0 {
		errors = append(errors, "valid service_id is required")
	}
	if req.Price <= 0 {
		errors = append(errors, "price must be positive")
	}
	if req.EstimatedDurationMin <= 0 {
		errors = append(errors, "estimated_duration_min must be positive")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

func (s *ServiceService) generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")
	return slug
}

// ==================== Request DTOs ====================

// CreateServiceRequest represents the create service request
type CreateServiceRequest struct {
	Name                   string             `json:"name" binding:"required"`
	Slug                   string             `json:"slug"`
	ShortDescription       string             `json:"short_description" binding:"required"`
	DetailedDescription    *string            `json:"detailed_description"`
	CategoryID             int                `json:"category_id" binding:"required"`
	ServiceType            string             `json:"service_type" binding:"required"`
	Complexity             int                `json:"complexity" binding:"required,min=1,max=5"`
	SkillLevelRequired     string             `json:"skill_level_required"`
	DefaultDurationMin     int                `json:"default_duration_min" binding:"required"`
	DefaultDurationMax     *int               `json:"default_duration_max"`
	SuggestedPriceMin      *float64           `json:"suggested_price_min"`
	SuggestedPriceMax      *float64           `json:"suggested_price_max"`
	Currency               string             `json:"currency"`
	TargetGender           string             `json:"target_gender"`
	TargetAgeMin           *int               `json:"target_age_min"`
	TargetAgeMax           *int               `json:"target_age_max"`
	HairTypes              models.StringArray `json:"hair_types"`
	RequiresConsultation   bool               `json:"requires_consultation"`
	RequiredTools          models.StringArray `json:"required_tools"`
	RequiredProducts       models.StringArray `json:"required_products"`
	RequiredCertifications models.StringArray `json:"required_certifications"`
	AllergenWarnings       models.StringArray `json:"allergen_warnings"`
	HealthPrecautions      models.StringArray `json:"health_precautions"`
	RequiresHealthCheck    bool               `json:"requires_health_check"`
	ImageURL               *string            `json:"image_url"`
	GalleryImages          models.StringArray `json:"gallery_images"`
	VideoURL               *string            `json:"video_url"`
	Tags                   models.StringArray `json:"tags"`
	SearchKeywords         models.StringArray `json:"search_keywords"`
	MetaDescription        *string            `json:"meta_description"`
	HasVariations          bool               `json:"has_variations"`
	AllowsAddOns           bool               `json:"allows_add_ons"`
	CreatedBy              *int               `json:"created_by"`
}

// UpdateServiceRequest represents the update service request
type UpdateServiceRequest struct {
	Name                 *string            `json:"name,omitempty"`
	Slug                 *string            `json:"slug,omitempty"`
	ShortDescription     *string            `json:"short_description,omitempty"`
	DetailedDescription  *string            `json:"detailed_description,omitempty"`
	CategoryID           *int               `json:"category_id,omitempty"`
	ServiceType          *string            `json:"service_type,omitempty"`
	Complexity           *int               `json:"complexity,omitempty"`
	SkillLevelRequired   *string            `json:"skill_level_required,omitempty"`
	DefaultDurationMin   *int               `json:"default_duration_min,omitempty"`
	DefaultDurationMax   *int               `json:"default_duration_max,omitempty"`
	SuggestedPriceMin    *float64           `json:"suggested_price_min,omitempty"`
	SuggestedPriceMax    *float64           `json:"suggested_price_max,omitempty"`
	Currency             *string            `json:"currency,omitempty"`
	TargetGender         *string            `json:"target_gender,omitempty"`
	HairTypes            models.StringArray `json:"hair_types,omitempty"`
	RequiresConsultation *bool              `json:"requires_consultation,omitempty"`
	RequiredTools        models.StringArray `json:"required_tools,omitempty"`
	RequiredProducts     models.StringArray `json:"required_products,omitempty"`
	ImageURL             *string            `json:"image_url,omitempty"`
	GalleryImages        models.StringArray `json:"gallery_images,omitempty"`
	Tags                 models.StringArray `json:"tags,omitempty"`
	SearchKeywords       models.StringArray `json:"search_keywords,omitempty"`
	HasVariations        *bool              `json:"has_variations,omitempty"`
	AllowsAddOns         *bool              `json:"allows_add_ons,omitempty"`
	IsActive             *bool              `json:"is_active,omitempty"`
	LastModifiedBy       *int               `json:"last_modified_by,omitempty"`
}

// CreateCategoryRequest represents the create category request
type CreateCategoryRequest struct {
	Name             string             `json:"name" binding:"required"`
	Slug             string             `json:"slug"`
	Description      *string            `json:"description"`
	ParentCategoryID *int               `json:"parent_category_id"`
	Level            int                `json:"level"`
	CategoryPath     string             `json:"category_path"`
	IconURL          *string            `json:"icon_url"`
	ColorHex         *string            `json:"color_hex"`
	ImageURL         *string            `json:"image_url"`
	SortOrder        int                `json:"sort_order"`
	IsFeatured       bool               `json:"is_featured"`
	MetaTitle        *string            `json:"meta_title"`
	MetaDescription  *string            `json:"meta_description"`
	Keywords         models.StringArray `json:"keywords"`
}

// UpdateCategoryRequest represents the update category request
type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	ColorHex    *string `json:"color_hex,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	IsFeatured  *bool   `json:"is_featured,omitempty"`
}

// CreateBarberServiceRequest represents the request to add a service to a barber
type CreateBarberServiceRequest struct {
	BarberID               int                `json:"barber_id" binding:"required"`
	ServiceID              int                `json:"service_id" binding:"required"`
	CustomName             *string            `json:"custom_name"`
	CustomDescription      *string            `json:"custom_description"`
	Price                  float64            `json:"price" binding:"required"`
	MaxPrice               *float64           `json:"max_price"`
	Currency               string             `json:"currency"`
	DiscountPrice          *float64           `json:"discount_price"`
	DiscountValidUntil     *time.Time         `json:"discount_valid_until"`
	EstimatedDurationMin   int                `json:"estimated_duration_min" binding:"required"`
	EstimatedDurationMax   *int               `json:"estimated_duration_max"`
	BufferTimeMinutes      int                `json:"buffer_time_minutes"`
	AdvanceNoticeHours     int                `json:"advance_notice_hours"`
	MaxAdvanceBookingDays  *int               `json:"max_advance_booking_days"`
	AvailableDays          models.StringArray `json:"available_days"`
	AvailableTimeSlots     models.JSONMap     `json:"available_time_slots"`
	RequiresConsultation   *bool              `json:"requires_consultation"`
	ConsultationDuration   *int               `json:"consultation_duration"`
	PreServiceInstructions *string            `json:"pre_service_instructions"`
	PostServiceCare        *string            `json:"post_service_care"`
	MinCustomerAge         *int               `json:"min_customer_age"`
	MaxCustomerAge         *int               `json:"max_customer_age"`
	IsSeasonal             bool               `json:"is_seasonal"`
	SeasonalStartMonth     *int               `json:"seasonal_start_month"`
	SeasonalEndMonth       *int               `json:"seasonal_end_month"`
	PortfolioImages        models.StringArray `json:"portfolio_images"`
	BeforeAfterImages      models.StringArray `json:"before_after_images"`
	IsPromotional          bool               `json:"is_promotional"`
	PromotionalText        *string            `json:"promotional_text"`
	PromotionStartDate     *time.Time         `json:"promotion_start_date"`
	PromotionEndDate       *time.Time         `json:"promotion_end_date"`
	IsFeatured             bool               `json:"is_featured"`
	DisplayOrder           int                `json:"display_order"`
	ServiceNote            *string            `json:"service_note"`
}

// UpdateBarberServiceRequest represents the request to update a barber's service
type UpdateBarberServiceRequest struct {
	CustomName            *string            `json:"custom_name,omitempty"`
	CustomDescription     *string            `json:"custom_description,omitempty"`
	Price                 *float64           `json:"price,omitempty"`
	MaxPrice              *float64           `json:"max_price,omitempty"`
	Currency              *string            `json:"currency,omitempty"`
	DiscountPrice         *float64           `json:"discount_price,omitempty"`
	DiscountValidUntil    *time.Time         `json:"discount_valid_until,omitempty"`
	EstimatedDurationMin  *int               `json:"estimated_duration_min,omitempty"`
	EstimatedDurationMax  *int               `json:"estimated_duration_max,omitempty"`
	BufferTimeMinutes     *int               `json:"buffer_time_minutes,omitempty"`
	AdvanceNoticeHours    *int               `json:"advance_notice_hours,omitempty"`
	MaxAdvanceBookingDays *int               `json:"max_advance_booking_days,omitempty"`
	AvailableDays         models.StringArray `json:"available_days,omitempty"`
	PortfolioImages       models.StringArray `json:"portfolio_images,omitempty"`
	IsPromotional         *bool              `json:"is_promotional,omitempty"`
	PromotionalText       *string            `json:"promotional_text,omitempty"`
	IsFeatured            *bool              `json:"is_featured,omitempty"`
	DisplayOrder          *int               `json:"display_order,omitempty"`
	ServiceNote           *string            `json:"service_note,omitempty"`
	IsActive              *bool              `json:"is_active,omitempty"`
}
