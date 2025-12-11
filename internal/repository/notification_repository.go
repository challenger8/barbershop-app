// internal/repository/notification_repository.go
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

// ========================================================================
// NOTIFICATION REPOSITORY - Data Access Layer for Notifications
// ========================================================================

// NotificationRepository handles notification data operations
type NotificationRepository struct {
	db *sqlx.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// ========================================================================
// FILTER STRUCTS
// ========================================================================

// NotificationFilters represents filter options for notification queries
type NotificationFilters struct {
	// Identity filters
	UserID int `form:"user_id"`

	// Type filters
	Type  string   `form:"type"`
	Types []string `form:"types"`

	// Status filters
	Status   string   `form:"status"`
	Statuses []string `form:"statuses"`

	// Priority filters
	Priority   string   `form:"priority"`
	Priorities []string `form:"priorities"`

	// Channel filters
	Channel string `form:"channel"`

	// Read status
	IsRead   *bool `form:"is_read"`
	IsUnread *bool `form:"is_unread"`

	// Related entity filters
	RelatedEntityType string `form:"related_entity_type"`
	RelatedEntityID   int    `form:"related_entity_id"`

	// Date range filters
	CreatedFrom  time.Time `form:"created_from" time_format:"2006-01-02T15:04:05Z07:00"`
	CreatedTo    time.Time `form:"created_to" time_format:"2006-01-02T15:04:05Z07:00"`
	ScheduledFor time.Time `form:"scheduled_for" time_format:"2006-01-02T15:04:05Z07:00"`

	// Search
	Search string `form:"search"`

	// Expired filter
	IncludeExpired bool `form:"include_expired"`

	// Sorting and pagination
	SortBy string `form:"sort_by"`
	Order  string `form:"order"`
	Limit  int    `form:"limit,default=50"`
	Offset int    `form:"offset,default=0"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalCount     int `json:"total_count" db:"total_count"`
	UnreadCount    int `json:"unread_count" db:"unread_count"`
	PendingCount   int `json:"pending_count" db:"pending_count"`
	SentCount      int `json:"sent_count" db:"sent_count"`
	DeliveredCount int `json:"delivered_count" db:"delivered_count"`
	FailedCount    int `json:"failed_count" db:"failed_count"`
}

// NOTE: Error variables are defined in errors.go to avoid duplication
// ErrNotificationNotFound, ErrInvalidNotificationType, ErrInvalidNotificationStatus,
// ErrNotificationExpired, ErrNotificationAlreadySent

// ========================================================================
// VALID STATUS VALUES
// ========================================================================

// ValidNotificationStatuses defines allowed notification statuses - using config constants
var ValidNotificationStatuses = []string{
	config.NotificationStatusPending,
	config.NotificationStatusSent,
	config.NotificationStatusDelivered,
	config.NotificationStatusRead,
	config.NotificationStatusFailed,
}

// ValidNotificationTypes defines allowed notification types - using config constants
var ValidNotificationTypes = []string{
	config.NotificationTypeBookingConfirmation,
	config.NotificationTypeBookingReminder,
	config.NotificationTypeBookingCancelled,
	config.NotificationTypeBookingRescheduled,
	config.NotificationTypeBookingCompleted,
	config.NotificationTypeReviewRequest,
	config.NotificationTypeReviewResponse,
	config.NotificationTypePaymentReceived,
	config.NotificationTypePaymentFailed,
	config.NotificationTypeAccountWelcome,
	config.NotificationTypeAccountVerification,
	config.NotificationTypePasswordReset,
	config.NotificationTypePromotion,
	config.NotificationTypeSystemAlert,
}

// ValidNotificationPriorities defines allowed priority levels - using config constants
var ValidNotificationPriorities = []string{
	config.NotificationPriorityLow,
	config.NotificationPriorityNormal,
	config.NotificationPriorityHigh,
	config.NotificationPriorityUrgent,
}

// ValidNotificationChannels defines allowed delivery channels - using config constants
var ValidNotificationChannels = []string{
	config.NotificationChannelApp,
	config.NotificationChannelEmail,
	config.NotificationChannelSMS,
	config.NotificationChannelPush,
}

// IsValidNotificationStatus checks if a status is valid
func IsValidNotificationStatus(status string) bool {
	return IsValidValue(status, ValidNotificationStatuses)
}

// IsValidNotificationType checks if a type is valid
func IsValidNotificationType(notifType string) bool {
	return IsValidValue(notifType, ValidNotificationTypes)
}
// IsValidNotificationChannel checks if a channel is valid
func IsValidNotificationChannel(channel string) bool {
	return IsValidValue(channel, ValidNotificationChannels)
}
// ========================================================================
// CREATE OPERATIONS
// ========================================================================

// Create inserts a new notification into the database
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (
			user_id, title, message, type,
			channels, status, priority,
			related_entity_type, related_entity_id,
			data, scheduled_for, expires_at,
			created_at
		) VALUES (
			:user_id, :title, :message, :type,
			:channels, :status, :priority,
			:related_entity_type, :related_entity_id,
			:data, :scheduled_for, :expires_at,
			:created_at
		) RETURNING id
	`

	// Set defaults using helpers
	notification.CreatedAt = time.Now()
	SetDefaultString(&notification.Status, config.NotificationStatusPending)
	SetDefaultString(&notification.Priority, config.NotificationPriorityNormal)

	rows, err := r.db.NamedQueryContext(ctx, query, notification)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&notification.ID); err != nil {
			return fmt.Errorf("failed to scan notification id: %w", err)
		}
	}

	return nil
}

// CreateBatch inserts multiple notifications at once
func (r *NotificationRepository) CreateBatch(ctx context.Context, notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	query := `
		INSERT INTO notifications (
			user_id, title, message, type,
			channels, status, priority,
			related_entity_type, related_entity_id,
			data, scheduled_for, expires_at,
			created_at
		) VALUES (
			:user_id, :title, :message, :type,
			:channels, :status, :priority,
			:related_entity_type, :related_entity_id,
			:data, :scheduled_for, :expires_at,
			:created_at
		)
	`

	now := time.Now()
	for _, n := range notifications {
		n.CreatedAt = now
		SetDefaultString(&n.Status, config.NotificationStatusPending)
		SetDefaultString(&n.Priority, config.NotificationPriorityNormal)
	}

	_, err := r.db.NamedExecContext(ctx, query, notifications)
	if err != nil {
		return fmt.Errorf("failed to create notifications batch: %w", err)
	}

	return nil
}

// ========================================================================
// READ OPERATIONS - FindByID
// ========================================================================

// FindByID retrieves a notification by its ID
func (r *NotificationRepository) FindByID(ctx context.Context, id int) (*models.Notification, error) {
	query := `SELECT * FROM notifications WHERE id = $1`

	var notification models.Notification
	err := r.db.GetContext(ctx, &notification, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to find notification by id: %w", err)
	}

	return &notification, nil
}

// ========================================================================
// READ OPERATIONS - FindAll with Filters
// ========================================================================

// FindAll retrieves notifications with optional filters
func (r *NotificationRepository) FindAll(ctx context.Context, filters NotificationFilters) ([]models.Notification, error) {
	query := `SELECT * FROM notifications WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// User filter
	if filters.UserID > 0 {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, filters.UserID)
		argCount++
	}

	// Single type filter
	if filters.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, filters.Type)
		argCount++
	}

	// Multiple types filter
	if len(filters.Types) > 0 {
		placeholders := make([]string, len(filters.Types))
		for i, t := range filters.Types {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, t)
			argCount++
		}
		query += fmt.Sprintf(" AND type IN (%s)", strings.Join(placeholders, ", "))
	}

	// Single status filter
	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	// Multiple statuses filter
	if len(filters.Statuses) > 0 {
		placeholders := make([]string, len(filters.Statuses))
		for i, s := range filters.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, s)
			argCount++
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ", "))
	}

	// Priority filter
	if filters.Priority != "" {
		query += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, filters.Priority)
		argCount++
	}

	// Multiple priorities filter
	if len(filters.Priorities) > 0 {
		placeholders := make([]string, len(filters.Priorities))
		for i, p := range filters.Priorities {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, p)
			argCount++
		}
		query += fmt.Sprintf(" AND priority IN (%s)", strings.Join(placeholders, ", "))
	}

	// Channel filter (checks if channel is in array)
	if filters.Channel != "" {
		query += fmt.Sprintf(" AND $%d = ANY(channels)", argCount)
		args = append(args, filters.Channel)
		argCount++
	}

	// Read status filter
	if filters.IsRead != nil {
		if *filters.IsRead {
			query += " AND read_at IS NOT NULL"
		}
	}
	if filters.IsUnread != nil {
		if *filters.IsUnread {
			query += " AND read_at IS NULL"
		}
	}

	// Related entity filters
	if filters.RelatedEntityType != "" {
		query += fmt.Sprintf(" AND related_entity_type = $%d", argCount)
		args = append(args, filters.RelatedEntityType)
		argCount++
	}
	if filters.RelatedEntityID > 0 {
		query += fmt.Sprintf(" AND related_entity_id = $%d", argCount)
		args = append(args, filters.RelatedEntityID)
		argCount++
	}

	// Date range filters
	if !filters.CreatedFrom.IsZero() {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, filters.CreatedFrom)
		argCount++
	}
	if !filters.CreatedTo.IsZero() {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, filters.CreatedTo)
		argCount++
	}

	// Scheduled for filter
	if !filters.ScheduledFor.IsZero() {
		query += fmt.Sprintf(" AND scheduled_for <= $%d", argCount)
		args = append(args, filters.ScheduledFor)
		argCount++
	}

	// Search filter
	if filters.Search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR message ILIKE $%d)", argCount, argCount+1)
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argCount += 2
	}

	// Expired filter
	if !filters.IncludeExpired {
		query += " AND (expires_at IS NULL OR expires_at > NOW())"
	}

	// Sorting
	orderBy := "created_at DESC" // Default sort
	if filters.SortBy != "" {
		order := "DESC"
		if filters.Order != "" && (filters.Order == "ASC" || filters.Order == "asc") {
			order = "ASC"
		}
		switch filters.SortBy {
		case "priority":
			// Priority order: urgent > high > normal > low
			orderBy = fmt.Sprintf("CASE priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'normal' THEN 3 ELSE 4 END %s", order)
		case "scheduled_for":
			orderBy = fmt.Sprintf("scheduled_for %s NULLS LAST", order)
		case "created_at":
			orderBy = fmt.Sprintf("created_at %s", order)
		}
	}
	query += " ORDER BY " + orderBy

	// Pagination
	limit := 50
	if filters.Limit > 0 {
		limit = filters.Limit
	}
	offset := 0
	if filters.Offset > 0 {
		offset = filters.Offset
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	var notifications []models.Notification
	err := r.db.SelectContext(ctx, &notifications, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find notifications: %w", err)
	}

	return notifications, nil
}

// ========================================================================
// READ OPERATIONS - Specific Queries
// ========================================================================

// FindByUserID retrieves all notifications for a user
func (r *NotificationRepository) FindByUserID(ctx context.Context, userID int, filters NotificationFilters) ([]models.Notification, error) {
	filters.UserID = userID
	return r.FindAll(ctx, filters)
}

// GetUnreadNotifications retrieves unread notifications for a user
func (r *NotificationRepository) GetUnreadNotifications(ctx context.Context, userID int, limit int) ([]models.Notification, error) {
	isUnread := true
	filters := NotificationFilters{
		UserID:   userID,
		IsUnread: &isUnread,
		Limit:    limit,
		SortBy:   "created_at",
		Order:    "DESC",
	}
	return r.FindAll(ctx, filters)
}

// GetPendingNotifications retrieves notifications ready to be sent
func (r *NotificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]models.Notification, error) {
	query := `
		SELECT * FROM notifications
		WHERE status = 'pending'
		AND (scheduled_for IS NULL OR scheduled_for <= NOW())
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY
			CASE priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'normal' THEN 3 ELSE 4 END,
			created_at ASC
		LIMIT $1
	`

	var notifications []models.Notification
	err := r.db.SelectContext(ctx, &notifications, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}

	return notifications, nil
}

// GetByRelatedEntity retrieves notifications for a specific entity
func (r *NotificationRepository) GetByRelatedEntity(ctx context.Context, entityType string, entityID int) ([]models.Notification, error) {
	filters := NotificationFilters{
		RelatedEntityType: entityType,
		RelatedEntityID:   entityID,
	}
	return r.FindAll(ctx, filters)
}

// ========================================================================
// UPDATE OPERATIONS
// ========================================================================

// UpdateStatus updates the status of a notification
func (r *NotificationRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	if !IsValidNotificationStatus(status) {
		return ErrInvalidNotificationStatus
	}

	query := `UPDATE notifications SET status = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return CheckRowsAffected(result, ErrNotificationNotFound)
}

// MarkAsSent marks a notification as sent
func (r *NotificationRepository) MarkAsSent(ctx context.Context, id int) error {
	now := time.Now()
	query := `
		UPDATE notifications SET
			status = 'sent',
			sent_at = $1
		WHERE id = $2 AND status = 'pending'
	`

	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as sent: %w", err)
	}

	return CheckRowsAffected(result, ErrNotificationAlreadySent)
}

// MarkAsDelivered marks a notification as delivered
func (r *NotificationRepository) MarkAsDelivered(ctx context.Context, id int) error {
	now := time.Now()
	query := `
		UPDATE notifications SET
			status = 'delivered',
			delivered_at = $1
		WHERE id = $2 AND status IN ('pending', 'sent')
	`

	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as delivered: %w", err)
	}

	return CheckRowsAffected(result, ErrNotificationNotFound)
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id int) error {
	now := time.Now()
	query := `
		UPDATE notifications SET
			status = 'read',
			read_at = $1
		WHERE id = $2 AND read_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	// Either not found or already read - not an error
	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID int) (int, error) {
	now := time.Now()
	query := `
		UPDATE notifications SET
			status = 'read',
			read_at = $1
		WHERE user_id = $2 AND read_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// MarkAsFailed marks a notification as failed
func (r *NotificationRepository) MarkAsFailed(ctx context.Context, id int, errorMsg string) error {
	query := `
		UPDATE notifications SET
			status = 'failed',
			data = data || jsonb_build_object('error', $1, 'failed_at', $2)
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, errorMsg, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as failed: %w", err)
	}

	return CheckRowsAffected(result, ErrNotificationNotFound)
}

// ========================================================================
// DELETE OPERATIONS
// ========================================================================

// Delete removes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return CheckRowsAffected(result, ErrNotificationNotFound)
}

// DeleteOldNotifications removes notifications older than specified duration
func (r *NotificationRepository) DeleteOldNotifications(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	query := `DELETE FROM notifications WHERE created_at < $1 AND status = 'read'`

	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old notifications: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// DeleteExpiredNotifications removes expired notifications that were never sent
func (r *NotificationRepository) DeleteExpiredNotifications(ctx context.Context) (int, error) {
	query := `DELETE FROM notifications WHERE expires_at < NOW() AND status = 'pending'`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired notifications: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// ========================================================================
// STATISTICS
// ========================================================================

// GetUserStats retrieves notification statistics for a user
func (r *NotificationRepository) GetUserStats(ctx context.Context, userID int) (*NotificationStats, error) {
	query := `
		SELECT
			COUNT(*) as total_count,
			COUNT(CASE WHEN read_at IS NULL THEN 1 END) as unread_count,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_count,
			COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent_count,
			COUNT(CASE WHEN status = 'delivered' THEN 1 END) as delivered_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count
		FROM notifications
		WHERE user_id = $1
		AND (expires_at IS NULL OR expires_at > NOW())
	`

	var stats NotificationStats
	err := r.db.GetContext(ctx, &stats, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user notification stats: %w", err)
	}

	return &stats, nil
}

// Count returns the total number of notifications matching the filters
func (r *NotificationRepository) Count(ctx context.Context, filters NotificationFilters) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if filters.UserID > 0 {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, filters.UserID)
		argCount++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if filters.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, filters.Type)
		argCount++
	}

	if !filters.IncludeExpired {
		query += " AND (expires_at IS NULL OR expires_at > NOW())"
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	return count, nil
}

// GetUnreadCount returns the unread notification count for a user
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID int) (int, error) {
	query := `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1
		AND read_at IS NULL
		AND (expires_at IS NULL OR expires_at > NOW())
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}
