package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// AuthorizationHeader is the HTTP header containing the JWT token
	AuthorizationHeader = "Authorization"
	// BearerScheme is the scheme for bearer tokens
	BearerScheme = "Bearer"
	// UserIDKey is the context key for the user ID
	UserIDKey = "user_id"
)

// JWTMiddleware validates JWT tokens and extracts user information
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Parse bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != BearerScheme {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// TODO: Validate JWT token signature and expiration
		// For now, we'll extract the user_id from a simple format:
		// The token should be a UUID representing the user
		// In production, use a proper JWT library (github.com/golang-jwt/jwt)
		userID := validateAndExtractUserID(token)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Store user ID in context
		c.Set(UserIDKey, userID)

		// Continue to next handler
		c.Next()
	}
}

// TODO: Implement proper JWT validation
// For now, we accept any UUID as valid
func validateAndExtractUserID(token string) string {
	// Simple validation: token should be a non-empty string
	// In production, validate JWT signature and claims
	if token == "" {
		return ""
	}
	// Return the token as user ID (should be UUID)
	return token
}

// GetUserIDFromContext extracts the user ID from the request context
func GetUserIDFromContext(c *gin.Context) string {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return ""
	}
	if uid, ok := userID.(string); ok {
		return uid
	}
	return ""
}
