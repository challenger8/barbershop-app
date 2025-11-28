// internal/repository/constants.go
package repository

// ========================================================================
// SQL COLUMN CONSTANTS - Eliminates repetition across repository queries
// ========================================================================

const (
	// USER_COLUMNS - Complete user table column list
	USER_COLUMNS = `
		id, uuid, email, password_hash, name, phone, user_type, status,
		email_verified, phone_verified, two_factor_enabled, 
		failed_login_attempts, locked_until,
		date_of_birth, gender, profile_picture_url,
		address, city, state, country, postal_code, latitude, longitude,
		preferences, notification_settings,
		created_at, updated_at, last_login_at, created_by, deleted_at
	`

	// BARBER_COLUMNS - Complete barber table column list
	BARBER_COLUMNS = `
		id, user_id, uuid, shop_name, business_name, business_registration_number,
		tax_id, address, address_line_2, city, state, country, postal_code,
		latitude, longitude, phone, business_email, website_url, description,
		years_experience, specialties, certifications, languages_spoken,
		profile_image_url, cover_image_url, gallery_images, working_hours,
		rating, total_reviews, total_bookings, response_time_minutes,
		acceptance_rate, cancellation_rate, status, is_verified,
		verification_date, verification_notes,
		advance_booking_days, min_booking_notice_hours, auto_accept_bookings,
		instant_booking_enabled, commission_rate, payout_method, payout_details,
		created_at, updated_at, last_active_at, deleted_at
	`

	// SERVICE_COLUMNS - Complete service table column list
	SERVICE_COLUMNS = `
		id, uuid, name, slug, short_description, detailed_description,
		category_id, subcategory_id, default_duration_min, complexity,
		target_gender, is_active, requires_consultation, available_for_home_service,
		preparation_time_min, cleanup_time_min, tags, search_keywords,
		image_url, gallery_images, video_url,
		average_global_rating, total_global_reviews, global_popularity_score,
		total_bookings_count, trending_score,
		seo_title, seo_description, seo_keywords,
		created_at, updated_at, created_by, last_modified_by
	`

	// BOOKING_COLUMNS - Complete booking table column list
	BOOKING_COLUMNS = `
		id, uuid, booking_number, customer_id, barber_id, time_slot_id,
		scheduled_start_time, scheduled_end_time, actual_start_time, actual_end_time,
		status, payment_status, payment_method, subtotal, tax_amount, discount_amount,
		total_price, currency, booking_source, special_requests, customer_notes,
		barber_notes, internal_notes, cancellation_reason, cancelled_by, cancelled_at,
		no_show, no_show_reason, reminder_sent_at, confirmation_sent_at,
		created_at, updated_at, deleted_at
	`

	// COMMON_WHERE_ACTIVE - Common WHERE clause for non-deleted records
	WHERE_ACTIVE = "WHERE deleted_at IS NULL"
)
