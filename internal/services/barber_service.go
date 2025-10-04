// internal/services/barber_service.go
package services

import (
	"context"
	"fmt"
	"strings"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/models"
	"barber-booking-system/internal/repository"

	"github.com/google/uuid"
)

type BarberService struct {
	repo  *repository.BarberRepository
	cache *cache.CacheService
}

// NewBarberService creates a new barber service with optional cache
func NewBarberService(repo *repository.BarberRepository, cache *cache.CacheService) *BarberService {
	return &BarberService{
		repo:  repo,
		cache: cache,
	}
}

// GetByID retrieves a barber by ID with caching
func (s *BarberService) GetByID(ctx context.Context, id int) (*models.Barber, error) {
	var barber *models.Barber
	var err error

	// Try cache first if available
	if s.cache != nil {
		var cachedBarber models.Barber
		err = s.cache.GetBarber(ctx, id, &cachedBarber)
		if err == nil {
			return &cachedBarber, nil // Cache hit
		}
	}

	// Cache miss or no cache - fetch from database
	barber, err = s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache if available
	if s.cache != nil {
		_ = s.cache.CacheBarber(ctx, id, barber)
	}

	return barber, nil
}

// GetBarberByID is an alias for GetByID (for compatibility with handlers)
func (s *BarberService) GetBarberByID(ctx context.Context, id int) (*models.Barber, error) {
	return s.GetByID(ctx, id)
}

// Update updates a barber with caching
func (s *BarberService) Update(ctx context.Context, barber *models.Barber) error {
	// Update in database
	err := s.repo.Update(ctx, barber)
	if err != nil {
		return err
	}

	// Invalidate cache if available
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, barber.ID)
	}

	return nil
}

// UpdateBarber is a wrapper that fetches, updates, and saves
func (s *BarberService) UpdateBarber(ctx context.Context, id int, req UpdateBarberRequest) (*models.Barber, error) {
	// Fetch existing barber
	barber, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch barber: %w", err)
	}

	// Apply updates from request - only update if the field is provided
	if req.ShopName != nil {
		barber.ShopName = *req.ShopName
	}
	if req.BusinessName != nil {
		barber.BusinessName = req.BusinessName
	}
	if req.Address != nil {
		barber.Address = *req.Address
	}
	if req.AddressLine2 != nil {
		barber.AddressLine2 = req.AddressLine2
	}
	if req.City != nil {
		barber.City = *req.City
	}
	if req.State != nil {
		barber.State = *req.State
	}
	if req.Country != nil {
		barber.Country = *req.Country
	}
	if req.PostalCode != nil {
		barber.PostalCode = *req.PostalCode
	}
	if req.Phone != nil {
		barber.Phone = req.Phone
	}
	if req.BusinessEmail != nil {
		barber.BusinessEmail = req.BusinessEmail
	}
	if req.WebsiteURL != nil {
		barber.WebsiteURL = req.WebsiteURL
	}
	if req.Description != nil {
		barber.Description = req.Description
	}
	if req.YearsExperience != nil {
		barber.YearsExperience = req.YearsExperience
	}
	if req.Specialties != nil {
		barber.Specialties = req.Specialties
	}
	if req.Certifications != nil {
		barber.Certifications = req.Certifications
	}
	if req.LanguagesSpoken != nil {
		barber.LanguagesSpoken = req.LanguagesSpoken
	}
	if req.ProfileImageURL != nil {
		barber.ProfileImageURL = req.ProfileImageURL
	}
	if req.CoverImageURL != nil {
		barber.CoverImageURL = req.CoverImageURL
	}
	if req.GalleryImages != nil {
		barber.GalleryImages = req.GalleryImages
	}
	if req.WorkingHours != nil {
		barber.WorkingHours = req.WorkingHours
	}

	// Update barber using the Update method
	if err := s.Update(ctx, barber); err != nil {
		return nil, fmt.Errorf("failed to update barber: %w", err)
	}

	return barber, nil
}

// Delete deletes a barber and invalidates cache
func (s *BarberService) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate cache if available
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, id)
	}

	return nil
}

// DeleteBarber is an alias for Delete
func (s *BarberService) DeleteBarber(ctx context.Context, id int) error {
	return s.Delete(ctx, id)
}

// UpdateStatus updates barber status with cache invalidation
func (s *BarberService) UpdateStatus(ctx context.Context, id int, status string) error {
	err := s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	// Invalidate cache if available
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, id)
	}

	return nil
}

// UpdateBarberStatus is an alias for UpdateStatus with validation
func (s *BarberService) UpdateBarberStatus(ctx context.Context, id int, status string) error {
	// Validate status
	validStatuses := []string{"pending", "active", "inactive", "suspended", "rejected"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid status: %s", status)
	}

	return s.UpdateStatus(ctx, id, status)
}

// GetBarberStatistics retrieves barber statistics with caching
func (s *BarberService) GetBarberStatistics(ctx context.Context, id int) (*repository.BarberStatistics, error) {
	// Try cache first if available
	if s.cache != nil {
		var cachedStats repository.BarberStatistics
		statsKey := fmt.Sprintf("stats:%d", id)
		err := s.cache.GetStats(ctx, statsKey, &cachedStats)
		if err == nil {
			return &cachedStats, nil
		}
	}

	// Fetch from database
	stats, err := s.repo.GetStatistics(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// Cache if available
	if s.cache != nil {
		statsKey := fmt.Sprintf("stats:%d", id)
		_ = s.cache.CacheStats(ctx, statsKey, stats)
	}

	return stats, nil
}

// SearchBarbers searches barbers by various criteria
func (s *BarberService) SearchBarbers(ctx context.Context, query string, filters repository.BarberFilters) ([]models.Barber, error) {
	filters.Search = query
	return s.repo.FindAllWithEnhancedSearch(ctx, filters)
}

// GetAllBarbers retrieves all barbers with filters
func (s *BarberService) GetAllBarbers(ctx context.Context, filters repository.BarberFilters) ([]models.Barber, error) {
	return s.repo.FindAllWithEnhancedSearch(ctx, filters)
}

// CreateBarber creates a new barber (UPDATED to use DTO pattern)
func (s *BarberService) CreateBarber(ctx context.Context, req CreateBarberRequest) (*models.Barber, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Build barber model from request
	barber := &models.Barber{
		UserID:                     req.UserID,
		UUID:                       uuid.New().String(),
		ShopName:                   req.ShopName,
		BusinessName:               req.BusinessName,
		BusinessRegistrationNumber: req.BusinessRegistrationNumber,
		TaxID:                      req.TaxID,
		Address:                    req.Address,
		AddressLine2:               req.AddressLine2,
		City:                       req.City,
		State:                      req.State,
		Country:                    req.Country,
		PostalCode:                 req.PostalCode,
		Latitude:                   req.Latitude,
		Longitude:                  req.Longitude,
		Phone:                      req.Phone,
		BusinessEmail:              req.BusinessEmail,
		WebsiteURL:                 req.WebsiteURL,
		Description:                req.Description,
		YearsExperience:            req.YearsExperience,
		Specialties:                req.Specialties,
		Certifications:             req.Certifications,
		LanguagesSpoken:            req.LanguagesSpoken,
		WorkingHours:               req.WorkingHours,
		// Set defaults
		AdvanceBookingDays:    30,
		MinBookingNoticeHours: 2,
		CommissionRate:        15.0,
		PayoutMethod:          "bank_transfer",
		Status:                "pending",
		Rating:                0.0,
		TotalReviews:          0,
		TotalBookings:         0,
	}

	// Create in database
	err := s.repo.Create(ctx, barber)
	if err != nil {
		return nil, fmt.Errorf("failed to create barber: %w", err)
	}

	// Cache the new barber if cache is available
	if s.cache != nil {
		_ = s.cache.CacheBarber(ctx, barber.ID, barber)
	}

	return barber, nil
}

// validateCreateRequest validates the create barber request
func (s *BarberService) validateCreateRequest(req CreateBarberRequest) error {
	var errors []string

	if req.UserID <= 0 {
		errors = append(errors, "user_id is required and must be positive")
	}
	if strings.TrimSpace(req.ShopName) == "" {
		errors = append(errors, "shop_name is required")
	}
	if strings.TrimSpace(req.Address) == "" {
		errors = append(errors, "address is required")
	}
	if strings.TrimSpace(req.City) == "" {
		errors = append(errors, "city is required")
	}
	if strings.TrimSpace(req.State) == "" {
		errors = append(errors, "state is required")
	}
	if strings.TrimSpace(req.Country) == "" {
		errors = append(errors, "country is required")
	}
	if strings.TrimSpace(req.PostalCode) == "" {
		errors = append(errors, "postal_code is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

// GetBarber is an alias for GetByID
func (s *BarberService) GetBarber(ctx context.Context, id int) (*models.Barber, error) {
	return s.GetByID(ctx, id)
}

// GetBarberByUUID retrieves a barber by UUID
func (s *BarberService) GetBarberByUUID(ctx context.Context, uuid string) (*models.Barber, error) {
	return s.repo.FindByUUID(ctx, uuid)
}

// Request DTOs

// CreateBarberRequest represents the create barber request
type CreateBarberRequest struct {
	UserID                     int                `json:"user_id" binding:"required"`
	ShopName                   string             `json:"shop_name" binding:"required"`
	BusinessName               *string            `json:"business_name"`
	BusinessRegistrationNumber *string            `json:"business_registration_number"`
	TaxID                      *string            `json:"tax_id"`
	Address                    string             `json:"address" binding:"required"`
	AddressLine2               *string            `json:"address_line_2"`
	City                       string             `json:"city" binding:"required"`
	State                      string             `json:"state" binding:"required"`
	Country                    string             `json:"country" binding:"required"`
	PostalCode                 string             `json:"postal_code" binding:"required"`
	Latitude                   *float64           `json:"latitude"`
	Longitude                  *float64           `json:"longitude"`
	Phone                      *string            `json:"phone"`
	BusinessEmail              *string            `json:"business_email"`
	WebsiteURL                 *string            `json:"website_url"`
	Description                *string            `json:"description"`
	YearsExperience            *int               `json:"years_experience"`
	Specialties                models.StringArray `json:"specialties"`
	Certifications             models.StringArray `json:"certifications"`
	LanguagesSpoken            models.StringArray `json:"languages_spoken"`
	WorkingHours               models.JSONMap     `json:"working_hours"`
}

// UpdateBarberRequest represents the update barber request
type UpdateBarberRequest struct {
	ShopName        *string            `json:"shop_name,omitempty"`
	BusinessName    *string            `json:"business_name,omitempty"`
	Address         *string            `json:"address,omitempty"`
	AddressLine2    *string            `json:"address_line_2,omitempty"`
	City            *string            `json:"city,omitempty"`
	State           *string            `json:"state,omitempty"`
	Country         *string            `json:"country,omitempty"`
	PostalCode      *string            `json:"postal_code,omitempty"`
	Phone           *string            `json:"phone,omitempty"`
	BusinessEmail   *string            `json:"business_email,omitempty"`
	WebsiteURL      *string            `json:"website_url,omitempty"`
	Description     *string            `json:"description,omitempty"`
	YearsExperience *int               `json:"years_experience,omitempty"`
	Specialties     models.StringArray `json:"specialties,omitempty"`
	Certifications  models.StringArray `json:"certifications,omitempty"`
	LanguagesSpoken models.StringArray `json:"languages_spoken,omitempty"`
	ProfileImageURL *string            `json:"profile_image_url,omitempty"`
	CoverImageURL   *string            `json:"cover_image_url,omitempty"`
	GalleryImages   models.StringArray `json:"gallery_images,omitempty"`
	WorkingHours    models.JSONMap     `json:"working_hours,omitempty"`
}
