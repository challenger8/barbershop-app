package models

import (
	"errors"
	"time"
)

// User represents a user account (customer, barber, or admin)
type User struct {
	ID           int     `json:"id" db:"id"`
	UUID         string  `json:"uuid" db:"uuid"`
	Email        string  `json:"email" db:"email"`
	PasswordHash string  `json:"-" db:"password_hash"` // Never include in JSON
	Name         string  `json:"name" db:"name"`
	Phone        *string `json:"phone" db:"phone"`
	UserType     string  `json:"user_type" db:"user_type"` // customer, barber, admin
	Status       string  `json:"status" db:"status"`       // active, inactive, suspended, deleted

	// Verification and security
	EmailVerified       bool       `json:"email_verified" db:"email_verified"`
	PhoneVerified       bool       `json:"phone_verified" db:"phone_verified"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until" db:"locked_until"`

	// Personal information
	DateOfBirth       *time.Time `json:"date_of_birth" db:"date_of_birth"`
	Gender            *string    `json:"gender" db:"gender"` // male, female, other, prefer_not_to_say
	ProfilePictureURL *string    `json:"profile_picture_url" db:"profile_picture_url"`

	// Location information
	Address    *string  `json:"address" db:"address"`
	City       *string  `json:"city" db:"city"`
	State      *string  `json:"state" db:"state"`
	Country    *string  `json:"country" db:"country"`
	PostalCode *string  `json:"postal_code" db:"postal_code"`
	Latitude   *float64 `json:"latitude" db:"latitude"`
	Longitude  *float64 `json:"longitude" db:"longitude"`

	// User preferences and settings
	Preferences          JSONMap `json:"preferences" db:"preferences"`
	NotificationSettings JSONMap `json:"notification_settings" db:"notification_settings"`

	// Audit fields
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedBy   *int       `json:"created_by" db:"created_by"`
	DeletedAt   *time.Time `json:"deleted_at" db:"deleted_at"`
}

// UserSession represents active user sessions for JWT management
type UserSession struct {
	ID              int       `json:"id" db:"id"`
	UserID          int       `json:"user_id" db:"user_id"`
	SessionToken    string    `json:"session_token" db:"session_token"`
	RefreshToken    *string   `json:"refresh_token" db:"refresh_token"`
	DeviceType      *string   `json:"device_type" db:"device_type"` // mobile, web, tablet
	DeviceID        *string   `json:"device_id" db:"device_id"`     // Unique device identifier
	IPAddress       *string   `json:"ip_address" db:"ip_address"`   // Store as string for simplicity
	UserAgent       *string   `json:"user_agent" db:"user_agent"`
	LocationCity    *string   `json:"location_city" db:"location_city"`
	LocationCountry *string   `json:"location_country" db:"location_country"`
	ExpiresAt       time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	LastActivityAt  time.Time `json:"last_activity_at" db:"last_activity_at"`
	IsActive        bool      `json:"is_active" db:"is_active"`
}

// VerificationCode represents email/phone verification codes
type VerificationCode struct {
	ID        int        `json:"id" db:"id"`
	UserID    int        `json:"user_id" db:"user_id"`
	Code      string     `json:"code" db:"code"`
	Type      string     `json:"type" db:"type"` // email_verification, phone_verification, password_reset, 2fa
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at" db:"used_at"`
	Attempts  int        `json:"attempts" db:"attempts"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// ========================================================================
// USER HELPER METHODS
// ========================================================================

// IsCustomer returns true if the user is a customer
func (u *User) IsCustomer() bool {
	return u.UserType == "customer"
}

// IsBarber returns true if the user is a barber
func (u *User) IsBarber() bool {
	return u.UserType == "barber"
}

// IsAdmin returns true if the user is an admin
func (u *User) IsAdmin() bool {
	return u.UserType == "admin"
}

// IsActive returns true if the user is active
func (u *User) IsActive() bool {
	return u.Status == "active"
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.Name
}

// GetDisplayLocation returns a formatted location string
func (u *User) GetDisplayLocation() string {
	if u.City != nil && u.State != nil {
		return *u.City + ", " + *u.State
	} else if u.City != nil {
		return *u.City
	}
	return ""
}

// Validate validates user fields
func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("name is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.UserType == "" {
		return errors.New("user type is required")
	}
	return nil
}
