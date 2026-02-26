package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// UserIDKey is the context key for the user ID
	UserIDKey = "user_id"
	// CookieName is the name of the cookie storing the JWT token
	CookieName = "access_token"
)

// JWTMiddleware validates JWT tokens from cookies and extracts user information
func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie
		token, err := c.Cookie(CookieName)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization cookie"})
			c.Abort()
			return
		}

		// Validate JWT token signature and expiration
		userID := validateAndExtractUserID(token, jwtSecret)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Store user ID in context
		c.Set(UserIDKey, userID)

		// Continue to next handler
		c.Next()
	}
}

// validateAndExtractUserID validates JWT token and extracts user ID from claims
func validateAndExtractUserID(tokenString, secret string) string {
	if tokenString == "" || secret == "" {
		return ""
	}

	// Parse and validate JWT token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return ""
	}

	// Extract claims
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return ""
	}

	return claims.Subject
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
