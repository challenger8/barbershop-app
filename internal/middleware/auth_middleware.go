// internal/middleware/auth_middleware.go
package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig defines the configuration for authentication middleware
type AuthConfig struct {
	SecretKey     string
	TokenLookup   string // "header:Authorization,cookie:token"
	TokenHeadName string // "Bearer"
	SkipPaths     []string
	RequiredRole  string // Optional: require specific role
	AllowedRoles  []string
}

// DefaultAuthConfig returns default authentication configuration
func DefaultAuthConfig(secretKey string) AuthConfig {
	return AuthConfig{
		SecretKey:     secretKey,
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		SkipPaths:     []string{"/health", "/metrics", "/api/v1/auth/login", "/api/v1/auth/register"},
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	UserType string `json:"user_type"` // customer, barber, admin
	jwt.RegisteredClaims
}

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(config AuthConfig) gin.HandlerFunc {
	// Create skip paths map for O(1) lookup
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// Skip authentication for certain paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Extract token from request
		token, err := extractToken(c, config.TokenLookup, config.TokenHeadName)
		if err != nil {
			RespondWithError(c, NewUnauthorizedError("Missing or invalid authorization token"))
			c.Abort()
			return
		}

		// Parse and validate token
		claims, err := parseToken(token, config.SecretKey)
		if err != nil {
			RespondWithError(c, NewUnauthorizedError("Invalid or expired token"))
			c.Abort()
			return
		}

		// Check role if required
		if config.RequiredRole != "" && claims.UserType != config.RequiredRole {
			RespondWithError(c, NewForbiddenError("Insufficient permissions"))
			c.Abort()
			return
		}

		// Check allowed roles
		if len(config.AllowedRoles) > 0 {
			allowed := false
			for _, role := range config.AllowedRoles {
				if claims.UserType == role {
					allowed = true
					break
				}
			}
			if !allowed {
				RespondWithError(c, NewForbiddenError("Insufficient permissions"))
				c.Abort()
				return
			}
		}

		// Store claims in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("user_type", claims.UserType)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireAuth is a simple auth middleware that requires authentication
func RequireAuth(secretKey string) gin.HandlerFunc {
	return AuthMiddleware(DefaultAuthConfig(secretKey))
}

// RequireRole creates middleware that requires a specific role
func RequireRole(secretKey string, role string) gin.HandlerFunc {
	config := DefaultAuthConfig(secretKey)
	config.RequiredRole = role
	return AuthMiddleware(config)
}

// RequireAnyRole creates middleware that requires any of the specified roles
func RequireAnyRole(secretKey string, roles ...string) gin.HandlerFunc {
	config := DefaultAuthConfig(secretKey)
	config.AllowedRoles = roles
	return AuthMiddleware(config)
}

// OptionalAuth adds user info to context if token is present, but doesn't require it
func OptionalAuth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to extract token
		token, err := extractToken(c, "header:Authorization", "Bearer")
		if err != nil {
			// No token, continue without auth
			c.Next()
			return
		}

		// Try to parse token
		claims, err := parseToken(token, secretKey)
		if err != nil {
			// Invalid token, continue without auth
			c.Next()
			return
		}

		// Store claims in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("user_type", claims.UserType)
		c.Set("claims", claims)

		c.Next()
	}
}

// extractToken extracts JWT token from request
func extractToken(c *gin.Context, tokenLookup string, tokenHeadName string) (string, error) {
	// Parse token lookup string (e.g., "header:Authorization,cookie:token")
	parts := strings.Split(tokenLookup, ",")

	for _, part := range parts {
		lookupParts := strings.Split(strings.TrimSpace(part), ":")
		if len(lookupParts) != 2 {
			continue
		}

		method := lookupParts[0]
		key := lookupParts[1]

		var token string
		switch method {
		case "header":
			// Get token from header
			authHeader := c.GetHeader(key)
			if authHeader == "" {
				continue
			}

			// Remove token head name (e.g., "Bearer ")
			if tokenHeadName != "" {
				prefix := tokenHeadName + " "
				if strings.HasPrefix(authHeader, prefix) {
					token = strings.TrimPrefix(authHeader, prefix)
				} else {
					token = authHeader
				}
			} else {
				token = authHeader
			}

			if token != "" {
				return token, nil
			}

		case "cookie":
			// Get token from cookie
			token, err := c.Cookie(key)
			if err == nil && token != "" {
				return token, nil
			}

		case "query":
			// Get token from query parameter
			token = c.Query(key)
			if token != "" {
				return token, nil
			}
		}
	}

	return "", fmt.Errorf("token not found")
}

// parseToken parses and validates JWT token
func parseToken(tokenString string, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateToken generates a new JWT token
func GenerateToken(userID int, email string, userType string, secretKey string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Email:    email,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// RefreshToken generates a new token from an existing token
func RefreshToken(tokenString string, secretKey string, expiresIn time.Duration) (string, error) {
	claims, err := parseToken(tokenString, secretKey)
	if err != nil {
		return "", err
	}

	return GenerateToken(claims.UserID, claims.Email, claims.UserType, secretKey, expiresIn)
}

// Helper functions to get user info from context

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(int)
	return id, ok
}

// MustGetUserID retrieves user ID or panics
func MustGetUserID(c *gin.Context) int {
	userID, ok := GetUserID(c)
	if !ok {
		panic("user_id not found in context")
	}
	return userID
}

// GetUserEmail retrieves user email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}

	e, ok := email.(string)
	return e, ok
}

// GetUserType retrieves user type from context
func GetUserType(c *gin.Context) (string, bool) {
	userType, exists := c.Get("user_type")
	if !exists {
		return "", false
	}

	t, ok := userType.(string)
	return t, ok
}

// GetClaims retrieves full claims from context
func GetClaims(c *gin.Context) (*Claims, bool) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	claims, ok := claimsVal.(*Claims)
	return claims, ok
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// IsAdmin checks if user is admin
func IsAdmin(c *gin.Context) bool {
	userType, ok := GetUserType(c)
	return ok && userType == "admin"
}

// IsBarber checks if user is barber
func IsBarber(c *gin.Context) bool {
	userType, ok := GetUserType(c)
	return ok && userType == "barber"
}

// IsCustomer checks if user is customer
func IsCustomer(c *gin.Context) bool {
	userType, ok := GetUserType(c)
	return ok && userType == "customer"
}

// RequireAdmin middleware that requires admin role
func RequireAdmin(secretKey string) gin.HandlerFunc {
	return RequireRole(secretKey, "admin")
}

// RequireBarber middleware that requires barber role
func RequireBarber(secretKey string) gin.HandlerFunc {
	return RequireRole(secretKey, "barber")
}

// RequireCustomer middleware that requires customer role
func RequireCustomer(secretKey string) gin.HandlerFunc {
	return RequireRole(secretKey, "customer")
}

// RequireBarberOrAdmin middleware that requires barber or admin role
func RequireBarberOrAdmin(secretKey string) gin.HandlerFunc {
	return RequireAnyRole(secretKey, "barber", "admin")
}
