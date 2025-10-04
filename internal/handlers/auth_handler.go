// internal/handlers/auth_handler.go
package handlers

import (
	"barber-booking-system/internal/middleware"
	"barber-booking-system/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userService *services.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param user body services.RegisterRequest true "Registration data"
// @Success 201 {object} SuccessResponse{data=services.AuthResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, middleware.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	// Register user
	authResponse, err := h.userService.Register(c.Request.Context(), req)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "email already registered") {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Registration failed",
				Message: "Email address is already registered",
			})
			return
		}

		if strings.Contains(err.Error(), "password must be") {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Validation failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Registration failed",
			Message: "Unable to create account. Please try again later",
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    authResponse,
		Message: "Account created successfully",
	})
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body services.LoginRequest true "Login credentials"
// @Success 200 {object} SuccessResponse{data=services.AuthResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, middleware.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	// Authenticate user
	authResponse, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "invalid email or password") {
			c.JSON(http.StatusUnauthorized, middleware.ErrorResponse{
				Error:   "Authentication failed",
				Message: "Invalid email or password",
			})
			return
		}

		if strings.Contains(err.Error(), "account is temporarily locked") {
			c.JSON(http.StatusForbidden, middleware.ErrorResponse{
				Error:   "Account locked",
				Message: err.Error(),
			})
			return
		}

		if strings.Contains(err.Error(), "account is") {
			c.JSON(http.StatusForbidden, middleware.ErrorResponse{
				Error:   "Account inactive",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Login failed",
			Message: "Unable to authenticate. Please try again later",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    authResponse,
		Message: "Login successful",
	})
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Generate a new JWT token from an existing valid token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse{data=services.AuthResponse}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Extract token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, middleware.ErrorResponse{
			Error:   "Missing token",
			Message: "Authorization header is required",
		})
		return
	}

	// Remove "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.JSON(http.StatusUnauthorized, middleware.ErrorResponse{
			Error:   "Invalid token format",
			Message: "Token must be in Bearer format",
		})
		return
	}

	// Refresh token
	authResponse, err := h.userService.RefreshToken(c.Request.Context(), tokenString)
	if err != nil {
		if strings.Contains(err.Error(), "invalid token") || strings.Contains(err.Error(), "expired") {
			c.JSON(http.StatusUnauthorized, middleware.ErrorResponse{
				Error:   "Invalid token",
				Message: "Token is invalid or expired",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Token refresh failed",
			Message: "Unable to refresh token. Please try again later",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    authResponse,
		Message: "Token refreshed successfully",
	})
}

// GetMe godoc
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse{data=services.UserProfileResponse}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get user profile
	profile, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "User not found",
				Message: "User account no longer exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Failed to fetch profile",
			Message: "Unable to retrieve profile. Please try again later",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    profile,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the profile of the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body services.UpdateProfileRequest true "Profile data"
// @Success 200 {object} SuccessResponse{data=services.UserProfileResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, middleware.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	// Update profile
	profile, err := h.userService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "User not found",
				Message: "User account no longer exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Update failed",
			Message: "Unable to update profile. Please try again later",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    profile,
		Message: "Profile updated successfully",
	})
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change the password of the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body services.ChangePasswordRequest true "Password data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, middleware.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, middleware.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	// Change password
	err := h.userService.ChangePassword(c.Request.Context(), userID, req)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "do not match") ||
			strings.Contains(err.Error(), "password must be") ||
			strings.Contains(err.Error(), "current password is incorrect") ||
			strings.Contains(err.Error(), "must be different") {
			c.JSON(http.StatusBadRequest, middleware.ErrorResponse{
				Error:   "Password change failed",
				Message: err.Error(),
			})
			return
		}

		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, middleware.ErrorResponse{
				Error:   "User not found",
				Message: "User account no longer exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, middleware.ErrorResponse{
			Error:   "Password change failed",
			Message: "Unable to change password. Please try again later",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// Logout godoc
// @Summary User logout
// @Description Logout the current user (client should delete the token)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Note: Since we're using JWT, logout is handled client-side by deleting the token
	// In a more complex system, you might want to blacklist tokens in Redis
	// For now, we just acknowledge the logout request

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}
