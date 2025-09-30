// internal/services/barber_service.go
package services

import (
	"barber-booking-system/internal/models"
	"barber-booking-system/internal/repository"
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
)

// BarberService handles barber business logic
type BarberService struct {
	barberRepo *repository.BarberRepository
}

// NewBarberService creates a new barber service
func NewBarberService(barberRepo *repository.BarberRepository) *BarberService {
	return &BarberService{
		barberRepo: barberRepo,
	}
}

// GetAllBarbers retrieves all barbers with filters - now using enhanced search
func (s *BarberService) GetAllBarbers(ctx context.Context, filters repository.BarberFilters) ([]models.Barber, error) {
	barbers, err := s.barberRepo.FindAllWithEnhancedSearch(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get barbers: %w", err)
	}
	return barbers, nil
}

// GetBarberByID retrieves a barber by ID
func (s *BarberService) GetBarberByID(ctx context.Context, id int) (*models.Barber, error) {
	barber, err := s.barberRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get barber: %w", err)
	}

	// Update last active
	_ = s.barberRepo.UpdateLastActive(ctx, id)

	return barber, nil
}

// GetBarberByUUID retrieves a barber by UUID
func (s *BarberService) GetBarberByUUID(ctx context.Context, uuid string) (*models.Barber, error) {
	barber, err := s.barberRepo.FindByUUID(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get barber: %w", err)
	}
	return barber, nil
}

// CreateBarber creates a new barber
func (s *BarberService) CreateBarber(ctx context.Context, req CreateBarberRequest) (*models.Barber, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Create barber model
	barber := &models.Barber{
		UserID:                req.UserID,
		UUID:                  uuid.New().String(),
		ShopName:              req.ShopName,
		BusinessName:          req.BusinessName,
		Address:               req.Address,
		AddressLine2:          req.AddressLine2,
		City:                  req.City,
		State:                 req.State,
		Country:               req.Country,
		PostalCode:            req.PostalCode,
		Phone:                 req.Phone,
		BusinessEmail:         req.BusinessEmail,
		WebsiteURL:            req.WebsiteURL,
		Description:           req.Description,
		YearsExperience:       req.YearsExperience,
		Specialties:           req.Specialties,
		Certifications:        req.Certifications,
		LanguagesSpoken:       req.LanguagesSpoken,
		WorkingHours:          req.WorkingHours,
		AdvanceBookingDays:    30,   // Default
		MinBookingNoticeHours: 2,    // Default
		CommissionRate:        15.0, // Default
		PayoutMethod:          "bank_transfer",
		Status:                "pending",
	}

	// Create barber
	if err := s.barberRepo.Create(ctx, barber); err != nil {
		return nil, fmt.Errorf("failed to create barber: %w", err)
	}

	return barber, nil
}

// UpdateBarber updates a barber
func (s *BarberService) UpdateBarber(ctx context.Context, id int, req UpdateBarberRequest) (*models.Barber, error) {
	// Get existing barber
	barber, err := s.barberRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find barber: %w", err)
	}

	// Update fields
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

	// Update barber
	if err := s.barberRepo.Update(ctx, barber); err != nil {
		return nil, fmt.Errorf("failed to update barber: %w", err)
	}

	return barber, nil
}

// DeleteBarber soft deletes a barber
func (s *BarberService) DeleteBarber(ctx context.Context, id int) error {
	if err := s.barberRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete barber: %w", err)
	}
	return nil
}

// UpdateBarberStatus updates barber status
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

	if err := s.barberRepo.UpdateStatus(ctx, id, status); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

// GetBarberStatistics retrieves barber statistics
func (s *BarberService) GetBarberStatistics(ctx context.Context, id int) (*repository.BarberStatistics, error) {
	stats, err := s.barberRepo.GetStatistics(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	return stats, nil
}

// SearchBarbers searches barbers by various criteria - now using enhanced search
func (s *BarberService) SearchBarbers(ctx context.Context, query string, filters repository.BarberFilters) ([]models.Barber, error) {
	filters.Search = query
	return s.barberRepo.FindAllWithEnhancedSearch(ctx, filters)
}

// GetNearbyBarbers retrieves barbers near a location
func (s *BarberService) GetNearbyBarbers(ctx context.Context, lat, lng float64, radiusKm float64) ([]models.Barber, error) {
	// This would require PostGIS or similar for proper geo queries
	// For now, return all barbers with coordinates
	filters := repository.BarberFilters{
		Status: "active",
		Limit:  50,
	}

	barbers, err := s.barberRepo.FindAllWithEnhancedSearch(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Filter by distance (simplified calculation)
	var nearbyBarbers []models.Barber
	for _, barber := range barbers {
		if barber.Latitude != nil && barber.Longitude != nil {
			distance := calculateDistance(lat, lng, *barber.Latitude, *barber.Longitude)
			if distance <= radiusKm {
				nearbyBarbers = append(nearbyBarbers, barber)
			}
		}
	}

	return nearbyBarbers, nil
}

// validateCreateRequest validates create barber request
func (s *BarberService) validateCreateRequest(req CreateBarberRequest) error {
	var errors []string

	if req.UserID <= 0 {
		errors = append(errors, "user_id is required")
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

// calculateDistance calculates distance between two points (Haversine formula)
func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371.0 // km

	dLat := toRadians(lat2 - lat1)
	dLng := toRadians(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

// Request/Response DTOs
type CreateBarberRequest struct {
	UserID          int                `json:"user_id" binding:"required"`
	ShopName        string             `json:"shop_name" binding:"required"`
	BusinessName    *string            `json:"business_name"`
	Address         string             `json:"address" binding:"required"`
	AddressLine2    *string            `json:"address_line_2"`
	City            string             `json:"city" binding:"required"`
	State           string             `json:"state" binding:"required"`
	Country         string             `json:"country" binding:"required"`
	PostalCode      string             `json:"postal_code" binding:"required"`
	Phone           *string            `json:"phone"`
	BusinessEmail   *string            `json:"business_email"`
	WebsiteURL      *string            `json:"website_url"`
	Description     *string            `json:"description"`
	YearsExperience *int               `json:"years_experience"`
	Specialties     models.StringArray `json:"specialties"`
	Certifications  models.StringArray `json:"certifications"`
	LanguagesSpoken models.StringArray `json:"languages_spoken"`
	WorkingHours    models.JSONMap     `json:"working_hours"`
}

type UpdateBarberRequest struct {
	ShopName        *string            `json:"shop_name"`
	BusinessName    *string            `json:"business_name"`
	Address         *string            `json:"address"`
	AddressLine2    *string            `json:"address_line_2"`
	City            *string            `json:"city"`
	State           *string            `json:"state"`
	Country         *string            `json:"country"`
	PostalCode      *string            `json:"postal_code"`
	Phone           *string            `json:"phone"`
	BusinessEmail   *string            `json:"business_email"`
	WebsiteURL      *string            `json:"website_url"`
	Description     *string            `json:"description"`
	YearsExperience *int               `json:"years_experience"`
	Specialties     models.StringArray `json:"specialties"`
	Certifications  models.StringArray `json:"certifications"`
	LanguagesSpoken models.StringArray `json:"languages_spoken"`
	ProfileImageURL *string            `json:"profile_image_url"`
	CoverImageURL   *string            `json:"cover_image_url"`
	GalleryImages   models.StringArray `json:"gallery_images"`
	WorkingHours    models.JSONMap     `json:"working_hours"`
}
