// internal/services/user_service.go
package services

import (
	"barber-booking-system/internal/config"
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/models"
	"barber-booking-system/internal/repository"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the bcrypt hashing cost (local constant as only used here)
	BcryptCost = 10
)

// UserService handles user business logic
type UserService struct {
	userRepo      *repository.UserRepository
	jwtSecret     string
	jwtExpiration time.Duration
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository, jwtSecret string, jwtExpiration time.Duration) *UserService {
	return &UserService{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

// Request structs

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Name     string  `json:"name" binding:"required,min=2,max=100"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=8"`
	Phone    *string `json:"phone" binding:"omitempty"`
	UserType string  `json:"user_type" binding:"omitempty,oneof=customer barber admin"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest represents profile update data
type UpdateProfileRequest struct {
	Name              *string                `json:"name" binding:"omitempty,min=2,max=100"`
	Phone             *string                `json:"phone" binding:"omitempty"`
	DateOfBirth       *time.Time             `json:"date_of_birth" binding:"omitempty"`
	Gender            *string                `json:"gender" binding:"omitempty,oneof=male female other prefer_not_to_say"`
	ProfilePictureURL *string                `json:"profile_picture_url" binding:"omitempty,url"`
	Address           *string                `json:"address" binding:"omitempty"`
	City              *string                `json:"city" binding:"omitempty"`
	State             *string                `json:"state" binding:"omitempty"`
	Country           *string                `json:"country" binding:"omitempty"`
	PostalCode        *string                `json:"postal_code" binding:"omitempty"`
	Preferences       map[string]interface{} `json:"preferences" binding:"omitempty"`
}

// ChangePasswordRequest represents password change data
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// Response structs

// AuthResponse represents authentication response
type AuthResponse struct {
	Token     string              `json:"token"`
	ExpiresAt time.Time           `json:"expires_at"`
	User      UserProfileResponse `json:"user"`
}

// UserProfileResponse represents user profile data (without password)
type UserProfileResponse struct {
	ID                int                    `json:"id"`
	UUID              string                 `json:"uuid"`
	Email             string                 `json:"email"`
	Name              string                 `json:"name"`
	Phone             *string                `json:"phone"`
	UserType          string                 `json:"user_type"`
	Status            string                 `json:"status"`
	EmailVerified     bool                   `json:"email_verified"`
	PhoneVerified     bool                   `json:"phone_verified"`
	DateOfBirth       *time.Time             `json:"date_of_birth"`
	Gender            *string                `json:"gender"`
	ProfilePictureURL *string                `json:"profile_picture_url"`
	Address           *string                `json:"address"`
	City              *string                `json:"city"`
	State             *string                `json:"state"`
	Country           *string                `json:"country"`
	PostalCode        *string                `json:"postal_code"`
	Preferences       map[string]interface{} `json:"preferences"`
	CreatedAt         time.Time              `json:"created_at"`
	LastLoginAt       *time.Time             `json:"last_login_at"`
}

// Register registers a new user
func (s *UserService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Validate password strength
	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if email already exists
	exists, err := s.userRepo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default user type if not provided
	userType := req.UserType
	if userType == "" {
		userType = "customer"
	}

	// Create user model
	user := &models.User{
		UUID:                uuid.New().String(),
		Email:               email,
		PasswordHash:        hashedPassword,
		Name:                req.Name,
		Phone:               req.Phone,
		UserType:            userType,
		Status:              config.UserStatusActive,
		EmailVerified:       false,
		PhoneVerified:       false,
		TwoFactorEnabled:    false,
		FailedLoginAttempts: 0,
		Preferences: models.JSONMap{
			"language": "en",
			"timezone": "UTC",
		},
		NotificationSettings: models.JSONMap{
			"email": true,
			"sms":   false,
			"push":  true,
		},
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, user.Email, user.UserType, s.jwtSecret, s.jwtExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return auth response
	return &AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(s.jwtExpiration),
		User:      s.toProfileResponse(user),
	}, nil
}

// Login authenticates a user and returns JWT token
func (s *UserService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if account is locked
	isLocked, err := s.userRepo.IsAccountLocked(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check account status: %w", err)
	}
	if isLocked {
		return nil, fmt.Errorf("account is temporarily locked due to too many failed login attempts. Please try again later")
	}

	// Check if account is inactive
	if user.Status != config.UserStatusActive {
		return nil, fmt.Errorf("account is %s. Please contact support", user.Status)
	}

	// Validate password
	if err := s.ValidatePassword(req.Password, user.PasswordHash); err != nil {
		// Handle failed login
		if handleErr := s.handleFailedLogin(ctx, user); handleErr != nil {
			return nil, fmt.Errorf("failed to handle failed login: %w", handleErr)
		}
		return nil, fmt.Errorf("invalid email or password")
	}

	// Reset failed login attempts on successful login
	if err := s.userRepo.ResetFailedLoginAttempts(ctx, user.ID); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Warning: failed to reset login attempts: %v\n", err)
	}

	// Update last login timestamp
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Warning: failed to update last login: %v\n", err)
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, user.Email, user.UserType, s.jwtSecret, s.jwtExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return auth response
	return &AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(s.jwtExpiration),
		User:      s.toProfileResponse(user),
	}, nil
}

// RefreshToken generates a new token from an existing token
func (s *UserService) RefreshToken(ctx context.Context, oldToken string) (*AuthResponse, error) {
	// Generate new token using middleware function
	newToken, err := middleware.RefreshToken(oldToken, s.jwtSecret, s.jwtExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// We need to extract user ID from the old token to get fresh user data
	// Since parseToken is private, we'll use a workaround by parsing the token ourselves
	token, err := jwt.ParseWithClaims(oldToken, &middleware.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*middleware.Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Get updated user info
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &AuthResponse{
		Token:     newToken,
		ExpiresAt: time.Now().Add(s.jwtExpiration),
		User:      s.toProfileResponse(user),
	}, nil
}

// GetProfile retrieves user profile
func (s *UserService) GetProfile(ctx context.Context, userID int) (*UserProfileResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	profile := s.toProfileResponse(user)
	return &profile, nil
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(ctx context.Context, userID int, req UpdateProfileRequest) (*UserProfileResponse, error) {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.DateOfBirth != nil {
		user.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		user.Gender = req.Gender
	}
	if req.ProfilePictureURL != nil {
		user.ProfilePictureURL = req.ProfilePictureURL
	}
	if req.Address != nil {
		user.Address = req.Address
	}
	if req.City != nil {
		user.City = req.City
	}
	if req.State != nil {
		user.State = req.State
	}
	if req.Country != nil {
		user.Country = req.Country
	}
	if req.PostalCode != nil {
		user.PostalCode = req.PostalCode
	}
	if req.Preferences != nil {
		user.Preferences = req.Preferences
	}

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	profile := s.toProfileResponse(user)
	return &profile, nil
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(ctx context.Context, userID int, req ChangePasswordRequest) error {
	// Validate new password matches confirmation
	if req.NewPassword != req.ConfirmPassword {
		return fmt.Errorf("new password and confirmation do not match")
	}

	// Validate new password strength
	if err := s.validatePassword(req.NewPassword); err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Validate old password
	if err := s.ValidatePassword(req.OldPassword, user.PasswordHash); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Check if new password is same as old password
	if err := s.ValidatePassword(req.NewPassword, user.PasswordHash); err == nil {
		return fmt.Errorf("new password must be different from current password")
	}

	// Hash new password
	hashedPassword, err := s.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Helper functions

// HashPassword hashes a password using bcrypt
func (s *UserService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// ValidatePassword validates a password against a hash
func (s *UserService) ValidatePassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// validatePassword validates password strength
func (s *UserService) validatePassword(password string) error {
	if len(password) < config.MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", config.MinPasswordLength)
	}

	// Additional password policy checks can be added here
	// For example: require uppercase, lowercase, numbers, special characters

	return nil
}

// handleFailedLogin handles failed login attempts and account locking
func (s *UserService) handleFailedLogin(ctx context.Context, user *models.User) error {
	// Increment failed login attempts
	if err := s.userRepo.IncrementFailedLoginAttempts(ctx, user.ID); err != nil {
		return err
	}

	// Check if we need to lock the account
	newAttempts := user.FailedLoginAttempts + 1
	if newAttempts >= config.MaxFailedLoginAttempts {
		if err := s.userRepo.LockAccount(ctx, user.ID, config.AccountLockDuration); err != nil {
			return err
		}
	}

	return nil
}

// toProfileResponse converts User model to UserProfileResponse
func (s *UserService) toProfileResponse(user *models.User) UserProfileResponse {
	var preferences map[string]interface{}
	if user.Preferences != nil {
		preferences = user.Preferences
	}

	return UserProfileResponse{
		ID:                user.ID,
		UUID:              user.UUID,
		Email:             user.Email,
		Name:              user.Name,
		Phone:             user.Phone,
		UserType:          user.UserType,
		Status:            user.Status,
		EmailVerified:     user.EmailVerified,
		PhoneVerified:     user.PhoneVerified,
		DateOfBirth:       user.DateOfBirth,
		Gender:            user.Gender,
		ProfilePictureURL: user.ProfilePictureURL,
		Address:           user.Address,
		City:              user.City,
		State:             user.State,
		Country:           user.Country,
		PostalCode:        user.PostalCode,
		Preferences:       preferences,
		CreatedAt:         user.CreatedAt,
		LastLoginAt:       user.LastLoginAt,
	}
}
