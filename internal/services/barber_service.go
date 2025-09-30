// internal/services/barber_service.go
package services

import (
	"context"
	"fmt"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/models"
	"barber-booking-system/internal/repository"
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
func (s *BarberService) UpdateBarber(ctx context.Context, id int, req *UpdateBarberRequest) (*models.Barber, error) {
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
		barber.Phone = req.Phone // Both are *string
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

// CreateBarber creates a new barber
func (s *BarberService) CreateBarber(ctx context.Context, barber *models.Barber) error {
	err := s.repo.Create(ctx, barber)
	if err != nil {
		return fmt.Errorf("failed to create barber: %w", err)
	}

	// Cache the new barber if cache is available
	if s.cache != nil {
		_ = s.cache.CacheBarber(ctx, barber.ID, barber)
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

// UpdateBarberRequest represents the update request structure
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
