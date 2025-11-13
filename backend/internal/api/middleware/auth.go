package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/services"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("profile_id", claims.ProfileID)
		c.Set("is_admin", claims.IsAdmin)

		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, &AuthError{Message: "user ID not found in context"}
	}
	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, &AuthError{Message: "invalid user ID in context"}
	}
	return uid, nil
}

// GetProfileID extracts profile ID from context
func GetProfileID(c *gin.Context) (uuid.UUID, error) {
	profileID, exists := c.Get("profile_id")
	if !exists {
		return uuid.Nil, &AuthError{Message: "profile ID not found in context"}
	}
	pid, ok := profileID.(uuid.UUID)
	if !ok {
		return uuid.Nil, &AuthError{Message: "invalid profile ID in context"}
	}
	return pid, nil
}

// AuthError represents an authentication/authorization error
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// MustGetProfileID extracts profile ID and returns 401 if not found
func MustGetProfileID(c *gin.Context) (uuid.UUID, bool) {
	profileID, err := GetProfileID(c)
	if err != nil || profileID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing profile ID"})
		return uuid.Nil, false
	}
	return profileID, true
}

// MustGetUserID extracts user ID and returns 401 if not found
func MustGetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, err := GetUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing user ID"})
		return uuid.Nil, false
	}
	return userID, true
}

// RequireAdmin middleware ensures user is an admin
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
