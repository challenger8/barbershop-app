// internal/repository/service_repository.go
package repository

import (
	"barber-booking-system/internal/config"
	"barber-booking-system/internal/models"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ServiceRepository handles service data operations
type ServiceRepository struct {
	db *sqlx.DB
}

// NewServiceRepository creates a new service repository
func NewServiceRepository(db *sqlx.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

// ServiceFilters represents filter options for services
type ServiceFilters struct {
	CategoryID   int
	ServiceType  string
	IsActive     *bool
	IsApproved   *bool
	Search       string
	MinRating    float64
	Complexity   int
	TargetGender string
	SortBy       string
	Limit        int
	Offset       int
}

// BarberServiceFilters represents filter options for barber services
type BarberServiceFilters struct {
	BarberID   int
	ServiceID  int
	IsActive   *bool
	IsFeatured *bool
	MinPrice   float64
	MaxPrice   float64
	Search     string
	SortBy     string
	Limit      int
	Offset     int
}

func (r *ServiceRepository) FindAll(ctx context.Context, filters ServiceFilters) ([]models.Service, error) {
	// Define sort column mappings
	sortMap := map[string]string{
		"name":       "s.name ASC",
		"popularity": "s.global_popularity_score DESC",
		"rating":     "s.average_global_rating DESC",
		"duration":   "s.default_duration_min ASC",
		"complexity": "s.complexity ASC",
		"default":    "s.created_at DESC",
	}

	// Build query using QueryBuilder
	qb := BuildServiceQuery().
		WhereIf(filters.CategoryID > 0, "s.category_id = ?", filters.CategoryID).
		WhereIf(filters.ServiceType != "", "s.service_type = ?", filters.ServiceType).
		WhereIf(filters.IsActive != nil, "s.is_active = ?", *filters.IsActive).
		WhereIf(filters.IsApproved != nil, "s.is_approved = ?", *filters.IsApproved).
		WhereIf(filters.MinRating > 0, "s.average_global_rating >= ?", filters.MinRating).
		WhereIf(filters.Complexity > 0, "s.complexity = ?", filters.Complexity).
		WhereIf(filters.TargetGender != "", "(s.target_gender = ? OR s.target_gender = 'all')", filters.TargetGender)

	// Add search across multiple fields
	if filters.Search != "" {
		qb.Search([]string{
			"s.name",
			"s.short_description",
			"s.detailed_description",
		}, filters.Search).
			SearchILike([]string{
				"s.tags",
				"s.search_keywords",
			}, filters.Search)
	}

	// Add sorting and pagination
	query, args := qb.
		OrderByWithDefault(filters.SortBy, "default", sortMap).
		Paginate(filters.Limit, filters.Offset).
		Build()

	// Execute query
	var services []models.Service
	err := r.db.SelectContext(ctx, &services, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %w", err)
	}

	return services, nil
}

// FindByID retrieves a service by ID
func (r *ServiceRepository) FindByID(ctx context.Context, id int) (*models.Service, error) {
	query := `
		SELECT * FROM services
		WHERE id = $1
	`

	var service models.Service
	err := r.db.GetContext(ctx, &service, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("failed to fetch service: %w", err)
	}

	return &service, nil
}

// FindByUUID retrieves a service by UUID
func (r *ServiceRepository) FindByUUID(ctx context.Context, uuid string) (*models.Service, error) {
	query := `
		SELECT * FROM services
		WHERE uuid = $1
	`

	var service models.Service
	err := r.db.GetContext(ctx, &service, query, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("failed to fetch service: %w", err)
	}

	return &service, nil
}

// FindBySlug retrieves a service by slug
func (r *ServiceRepository) FindBySlug(ctx context.Context, slug string) (*models.Service, error) {
	query := `
		SELECT * FROM services
		WHERE slug = $1 AND is_active = true
	`

	var service models.Service
	err := r.db.GetContext(ctx, &service, query, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("failed to fetch service: %w", err)
	}

	return &service, nil
}

// Create creates a new service
func (r *ServiceRepository) Create(ctx context.Context, service *models.Service) error {
	query := `
		INSERT INTO services (
			uuid, name, slug, short_description, detailed_description, category_id,
			service_type, complexity, skill_level_required,
			default_duration_min, default_duration_max,
			suggested_price_min, suggested_price_max, currency,
			target_gender, target_age_min, target_age_max, hair_types,
			requires_consultation, required_tools, required_products, required_certifications,
			allergen_warnings, health_precautions, requires_health_check,
			image_url, gallery_images, video_url,
			tags, search_keywords, meta_description,
			has_variations, allows_add_ons,
			global_popularity_score, total_global_bookings, average_global_rating, total_global_reviews,
			is_active, is_approved, approval_notes,
			created_at, updated_at, created_by, version
		) VALUES (
			:uuid, :name, :slug, :short_description, :detailed_description, :category_id,
			:service_type, :complexity, :skill_level_required,
			:default_duration_min, :default_duration_max,
			:suggested_price_min, :suggested_price_max, :currency,
			:target_gender, :target_age_min, :target_age_max, :hair_types,
			:requires_consultation, :required_tools, :required_products, :required_certifications,
			:allergen_warnings, :health_precautions, :requires_health_check,
			:image_url, :gallery_images, :video_url,
			:tags, :search_keywords, :meta_description,
			:has_variations, :allows_add_ons,
			:global_popularity_score, :total_global_bookings, :average_global_rating, :total_global_reviews,
			:is_active, :is_approved, :approval_notes,
			:created_at, :updated_at, :created_by, :version
		) RETURNING id
	`

	// Set timestamps
	now := time.Now()
	service.CreatedAt = now
	service.UpdatedAt = now

	// Set defaults
	if service.Currency == "" {
		service.Currency = config.DefaultCurrency
	}
	if service.Version == 0 {
		service.Version = 1
	}

	rows, err := r.db.NamedQueryContext(ctx, query, service)
	if err != nil {
		// Check for duplicate slug or name
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique constraint") {
			if strings.Contains(strings.ToLower(err.Error()), "slug") {
				return ErrDuplicateSlug
			}
			if strings.Contains(strings.ToLower(err.Error()), "name") {
				return ErrDuplicateService
			}
		}
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&service.ID); err != nil {
			return fmt.Errorf("failed to scan service id: %w", err)
		}
	}

	return nil
}

// Update updates a service
func (r *ServiceRepository) Update(ctx context.Context, service *models.Service) error {
	service.UpdatedAt = time.Now()
	service.Version++

	query := `
		UPDATE services SET
			name = :name,
			slug = :slug,
			short_description = :short_description,
			detailed_description = :detailed_description,
			category_id = :category_id,
			service_type = :service_type,
			complexity = :complexity,
			skill_level_required = :skill_level_required,
			default_duration_min = :default_duration_min,
			default_duration_max = :default_duration_max,
			suggested_price_min = :suggested_price_min,
			suggested_price_max = :suggested_price_max,
			currency = :currency,
			target_gender = :target_gender,
			target_age_min = :target_age_min,
			target_age_max = :target_age_max,
			hair_types = :hair_types,
			requires_consultation = :requires_consultation,
			required_tools = :required_tools,
			required_products = :required_products,
			required_certifications = :required_certifications,
			allergen_warnings = :allergen_warnings,
			health_precautions = :health_precautions,
			requires_health_check = :requires_health_check,
			image_url = :image_url,
			gallery_images = :gallery_images,
			video_url = :video_url,
			tags = :tags,
			search_keywords = :search_keywords,
			meta_description = :meta_description,
			has_variations = :has_variations,
			allows_add_ons = :allows_add_ons,
			is_active = :is_active,
			is_approved = :is_approved,
			approval_notes = :approval_notes,
			updated_at = :updated_at,
			last_modified_by = :last_modified_by,
			version = :version
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, service)
	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrServiceNotFound
	}

	return nil
}

// Delete deletes a service (soft delete by setting is_active = false)
func (r *ServiceRepository) Delete(ctx context.Context, id int) error {
	query := `
		UPDATE services
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrServiceNotFound
	}

	return nil
}

// ==================== Service Categories ====================

// FindAllCategories retrieves all service categories
func (r *ServiceRepository) FindAllCategories(ctx context.Context, activeOnly bool) ([]models.ServiceCategory, error) {
	query := `
		SELECT * FROM service_categories
		WHERE 1=1
	`

	if activeOnly {
		query += " AND is_active = true"
	}

	query += " ORDER BY sort_order ASC, name ASC"

	var categories []models.ServiceCategory
	err := r.db.SelectContext(ctx, &categories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	return categories, nil
}

// FindCategoryByID retrieves a category by ID
func (r *ServiceRepository) FindCategoryByID(ctx context.Context, id int) (*models.ServiceCategory, error) {
	query := `SELECT * FROM service_categories WHERE id = $1`

	var category models.ServiceCategory
	err := r.db.GetContext(ctx, &category, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	return &category, nil
}

// CreateCategory creates a new service category
func (r *ServiceRepository) CreateCategory(ctx context.Context, category *models.ServiceCategory) error {
	query := `
		INSERT INTO service_categories (
			name, slug, description, parent_category_id, level, category_path,
			icon_url, color_hex, image_url, sort_order,
			is_active, is_featured,
			meta_title, meta_description, keywords,
			service_count, barber_count, average_price, popularity_score,
			created_at, updated_at
		) VALUES (
			:name, :slug, :description, :parent_category_id, :level, :category_path,
			:icon_url, :color_hex, :image_url, :sort_order,
			:is_active, :is_featured,
			:meta_title, :meta_description, :keywords,
			:service_count, :barber_count, :average_price, :popularity_score,
			:created_at, :updated_at
		) RETURNING id
	`

	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	rows, err := r.db.NamedQueryContext(ctx, query, category)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique constraint") {
			return ErrDuplicateCategory
		}
		return fmt.Errorf("failed to create category: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&category.ID); err != nil {
			return fmt.Errorf("failed to scan category id: %w", err)
		}
	}

	return nil
}

// UpdateCategory updates a service category
func (r *ServiceRepository) UpdateCategory(ctx context.Context, category *models.ServiceCategory) error {
	category.UpdatedAt = time.Now()

	query := `
		UPDATE service_categories SET
			name = :name,
			slug = :slug,
			description = :description,
			parent_category_id = :parent_category_id,
			level = :level,
			category_path = :category_path,
			icon_url = :icon_url,
			color_hex = :color_hex,
			image_url = :image_url,
			sort_order = :sort_order,
			is_active = :is_active,
			is_featured = :is_featured,
			meta_title = :meta_title,
			meta_description = :meta_description,
			keywords = :keywords,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, category)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// DeleteCategory deletes a service category (soft delete)
func (r *ServiceRepository) DeleteCategory(ctx context.Context, id int) error {
	query := `
		UPDATE service_categories
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// ==================== Barber Services ====================

// FindBarberServices retrieves all services offered by a barber
func (r *ServiceRepository) FindBarberServices(ctx context.Context, filters BarberServiceFilters) ([]models.BarberService, error) {
	query := `
		SELECT bs.*, s.name as service_name, s.service_type, s.category_id
		FROM barber_services bs
		LEFT JOIN services s ON bs.service_id = s.id
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filters.BarberID > 0 {
		query += fmt.Sprintf(" AND bs.barber_id = $%d", argCount)
		args = append(args, filters.BarberID)
		argCount++
	}

	if filters.ServiceID > 0 {
		query += fmt.Sprintf(" AND bs.service_id = $%d", argCount)
		args = append(args, filters.ServiceID)
		argCount++
	}

	if filters.IsActive != nil {
		query += fmt.Sprintf(" AND bs.is_active = $%d", argCount)
		args = append(args, *filters.IsActive)
		argCount++
	}

	if filters.IsFeatured != nil {
		query += fmt.Sprintf(" AND bs.is_featured = $%d", argCount)
		args = append(args, *filters.IsFeatured)
		argCount++
	}

	if filters.MinPrice > 0 {
		query += fmt.Sprintf(" AND bs.price >= $%d", argCount)
		args = append(args, filters.MinPrice)
		argCount++
	}

	if filters.MaxPrice > 0 {
		query += fmt.Sprintf(" AND bs.price <= $%d", argCount)
		args = append(args, filters.MaxPrice)
		argCount++
	}

	// Sorting
	orderBy := "bs.display_order ASC, bs.created_at DESC"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "price":
			orderBy = "bs.price ASC"
		case "price_desc":
			orderBy = "bs.price DESC"
		case "rating":
			orderBy = "bs.average_rating DESC"
		case "popularity":
			orderBy = "bs.popularity_score DESC"
		case "bookings":
			orderBy = "bs.total_bookings DESC"
		}
	}
	query += " ORDER BY " + orderBy

	// Pagination
	limit := config.BarberServicesPageLimit

	offset := 0
	if filters.Limit > 0 {
		limit = filters.Limit
	}
	if filters.Offset > 0 {
		offset = filters.Offset
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	var barberServices []models.BarberService
	err := r.db.SelectContext(ctx, &barberServices, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch barber services: %w", err)
	}

	return barberServices, nil
}

// FindBarberServiceByID retrieves a barber service by ID
func (r *ServiceRepository) FindBarberServiceByID(ctx context.Context, id int) (*models.BarberService, error) {
	query := `
		SELECT bs.*, s.name as service_name
		FROM barber_services bs
		LEFT JOIN services s ON bs.service_id = s.id
		WHERE bs.id = $1
	`

	var barberService models.BarberService
	err := r.db.GetContext(ctx, &barberService, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBarberServiceNotFound
		}
		return nil, fmt.Errorf("failed to fetch barber service: %w", err)
	}

	return &barberService, nil
}

// CreateBarberService creates a new barber service
func (r *ServiceRepository) CreateBarberService(ctx context.Context, bs *models.BarberService) error {
	query := `
		INSERT INTO barber_services (
			barber_id, service_id, custom_name, custom_description,
			price, max_price, currency, discount_price, discount_valid_until,
			estimated_duration_min, estimated_duration_max, buffer_time_minutes,
			advance_notice_hours, max_advance_booking_days, available_days, available_time_slots,
			requires_consultation, consultation_duration, pre_service_instructions, post_service_care,
			min_customer_age, max_customer_age,
			is_seasonal, seasonal_start_month, seasonal_end_month,
			portfolio_images, before_after_images,
			total_bookings, total_revenue, average_rating, total_reviews,
			cancellation_rate, customer_satisfaction, repeat_customer_rate,
			bookings_last_30_days, revenue_last_30_days, popularity_score, demand_level,
			is_promotional, promotional_text, promotion_start_date, promotion_end_date, is_featured,
			display_order, service_note, is_active, paused_reason, paused_until,
			created_at, updated_at
		) VALUES (
			:barber_id, :service_id, :custom_name, :custom_description,
			:price, :max_price, :currency, :discount_price, :discount_valid_until,
			:estimated_duration_min, :estimated_duration_max, :buffer_time_minutes,
			:advance_notice_hours, :max_advance_booking_days, :available_days, :available_time_slots,
			:requires_consultation, :consultation_duration, :pre_service_instructions, :post_service_care,
			:min_customer_age, :max_customer_age,
			:is_seasonal, :seasonal_start_month, :seasonal_end_month,
			:portfolio_images, :before_after_images,
			:total_bookings, :total_revenue, :average_rating, :total_reviews,
			:cancellation_rate, :customer_satisfaction, :repeat_customer_rate,
			:bookings_last_30_days, :revenue_last_30_days, :popularity_score, :demand_level,
			:is_promotional, :promotional_text, :promotion_start_date, :promotion_end_date, :is_featured,
			:display_order, :service_note, :is_active, :paused_reason, :paused_until,
			:created_at, :updated_at
		) RETURNING id
	`

	now := time.Now()
	bs.CreatedAt = now
	bs.UpdatedAt = now

	if bs.Currency == "" {
		bs.Currency = config.DefaultCurrency
	}

	rows, err := r.db.NamedQueryContext(ctx, query, bs)
	if err != nil {
		// Check for duplicate barber_id + service_id combination
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("barber already offers this service: %w", err)
		}
		return fmt.Errorf("failed to create barber service: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&bs.ID); err != nil {
			return fmt.Errorf("failed to scan barber service id: %w", err)
		}
	}

	return nil
}

// UpdateBarberService updates a barber service
func (r *ServiceRepository) UpdateBarberService(ctx context.Context, bs *models.BarberService) error {
	bs.UpdatedAt = time.Now()

	query := `
		UPDATE barber_services SET
			custom_name = :custom_name,
			custom_description = :custom_description,
			price = :price,
			max_price = :max_price,
			currency = :currency,
			discount_price = :discount_price,
			discount_valid_until = :discount_valid_until,
			estimated_duration_min = :estimated_duration_min,
			estimated_duration_max = :estimated_duration_max,
			buffer_time_minutes = :buffer_time_minutes,
			advance_notice_hours = :advance_notice_hours,
			max_advance_booking_days = :max_advance_booking_days,
			available_days = :available_days,
			available_time_slots = :available_time_slots,
			requires_consultation = :requires_consultation,
			consultation_duration = :consultation_duration,
			pre_service_instructions = :pre_service_instructions,
			post_service_care = :post_service_care,
			min_customer_age = :min_customer_age,
			max_customer_age = :max_customer_age,
			is_seasonal = :is_seasonal,
			seasonal_start_month = :seasonal_start_month,
			seasonal_end_month = :seasonal_end_month,
			portfolio_images = :portfolio_images,
			before_after_images = :before_after_images,
			is_promotional = :is_promotional,
			promotional_text = :promotional_text,
			promotion_start_date = :promotion_start_date,
			promotion_end_date = :promotion_end_date,
			is_featured = :is_featured,
			display_order = :display_order,
			service_note = :service_note,
			is_active = :is_active,
			paused_reason = :paused_reason,
			paused_until = :paused_until,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, bs)
	if err != nil {
		return fmt.Errorf("failed to update barber service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrBarberServiceNotFound
	}

	return nil
}

// DeleteBarberService deletes a barber service (soft delete)
func (r *ServiceRepository) DeleteBarberService(ctx context.Context, id int) error {
	query := `
		UPDATE barber_services
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete barber service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrBarberServiceNotFound
	}

	return nil
}

// GetServicesByBarberID retrieves all services for a specific barber
func (r *ServiceRepository) GetServicesByBarberID(ctx context.Context, barberID int) ([]models.BarberService, error) {
	filters := BarberServiceFilters{
		BarberID: barberID,
		Limit:    100,
	}
	isActive := true
	filters.IsActive = &isActive

	return r.FindBarberServices(ctx, filters)
}

// GetBarbersByServiceID retrieves all barbers offering a specific service
func (r *ServiceRepository) GetBarbersByServiceID(ctx context.Context, serviceID int) ([]models.BarberService, error) {
	filters := BarberServiceFilters{
		ServiceID: serviceID,
		Limit:     100,
	}
	isActive := true
	filters.IsActive = &isActive

	return r.FindBarberServices(ctx, filters)
}
