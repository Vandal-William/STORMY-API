package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// UserIDKey is the context key for the user ID
	UserIDKey = "user_id"
	// CookieName is the name of the cookie storing the JWT token
	CookieName = "access_token"
)

// JWTMiddleware validates JWT tokens from cookies OR Authorization headers and extracts user information
func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		var err error

		// 1. Try to get token from cookie first
		token, err = c.Cookie(CookieName)
		if err != nil {
			// 2. If no cookie, try Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization cookie or header"})
				c.Abort()
				return
			}

			// Extract token from "Bearer <token>" format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
				c.Abort()
				return
			}
			token = parts[1]
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
