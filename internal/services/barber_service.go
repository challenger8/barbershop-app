// internal/services/barber_service.go
package services

import (
	"context"
	"fmt"
	"strings"

	"barber-booking-system/internal/cache"
	"barber-booking-system/internal/config"
	"barber-booking-system/internal/logger"
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

// UpdateBarber is a wrapper that fetches, updates, and saves
func (s *BarberService) UpdateBarber(ctx context.Context, id int, req UpdateBarberRequest) (*models.Barber, error) {
	log := logger.FromContext(ctx)

	log.Debug("Updating barber").
		Int("barber_id", id).
		Send()

	// Fetch existing barber
	barber, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Warn("Barber not found for update").
			Int("barber_id", id).
			Err(err).
			Send()
		return nil, fmt.Errorf("failed to fetch barber: %w", err)
	}

	// Apply updates from request - only update if the field is provided
	if req.ShopName != nil {
		barber.ShopName = *req.ShopName
	}
	if req.Description != nil {
		barber.Description = req.Description
	}
	if req.Specialties != nil {
		barber.Specialties = req.Specialties
	}
	if req.Address != nil {
		barber.Address = *req.Address
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

	// Save updates
	if err := s.repo.Update(ctx, barber); err != nil {
		log.Error(err).
			Int("barber_id", id).
			Msg("Failed to update barber")
		return nil, fmt.Errorf("failed to update barber: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, id)
	}

	log.Info("Barber updated successfully").
		Int("barber_id", id).
		Str("shop_name", barber.ShopName).
		Send()

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

// UpdateStatus updates barber status
func (s *BarberService) UpdateStatus(ctx context.Context, id int, status string) error {
	log := logger.FromContext(ctx)

	log.Info("Updating barber status").
		Int("barber_id", id).
		Str("new_status", status).
		Send()

	// Get current barber for logging
	barber, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Warn("Barber not found for status update").
			Int("barber_id", id).
			Err(err).
			Send()
		return err
	}

	oldStatus := barber.Status

	// Update status
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		log.Error(err).
			Int("barber_id", id).
			Str("new_status", status).
			Msg("Failed to update barber status")
		return err
	}

	// Invalidate cache if available
	if s.cache != nil {
		_ = s.cache.InvalidateBarber(ctx, id)
	}

	log.Info("Barber status updated successfully").
		Int("barber_id", id).
		Str("old_status", oldStatus).
		Str("new_status", status).
		Send()

	return nil
}

// UpdateBarberStatus is an alias for UpdateStatus with validation
func (s *BarberService) UpdateBarberStatus(ctx context.Context, id int, status string) error {
	// Validate status using config constants
	validStatuses := []string{
		config.BarberStatusPending,
		config.BarberStatusActive,
		config.BarberStatusInactive,
		config.BarberStatusSuspended,
		config.BarberStatusRejected,
	}
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
	return s.repo.FindAll(ctx, filters)
}

// GetAllBarbers retrieves all barbers with filters
func (s *BarberService) GetAllBarbers(ctx context.Context, filters repository.BarberFilters) ([]models.Barber, error) {
	return s.repo.FindAll(ctx, filters)
}

// CreateBarber creates a new barber (UPDATED to use DTO pattern)
func (s *BarberService) CreateBarber(ctx context.Context, req CreateBarberRequest) (*models.Barber, error) {
	log := logger.FromContext(ctx)

	log.Debug("Creating barber").
		Int("user_id", req.UserID).
		Str("shop_name", req.ShopName).
		Send()

	// Build barber model from request
	barber := &models.Barber{
		UUID:        uuid.New().String(),
		UserID:      req.UserID,
		ShopName:    req.ShopName,
		Description: req.Description,
		Specialties: req.Specialties,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Country:     req.Country,
		PostalCode:  req.PostalCode,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Status:      config.BarberStatusPending,
	}

	// Create in database
	if err := s.repo.Create(ctx, barber); err != nil {
		log.Error(err).
			Int("user_id", req.UserID).
			Msg("Failed to create barber")
		return nil, fmt.Errorf("failed to create barber: %w", err)
	}

	log.Info("Barber created successfully").
		Int("barber_id", barber.ID).
		Str("uuid", barber.UUID).
		Str("shop_name", barber.ShopName).
		Send()

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
