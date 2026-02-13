package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func safeServiceStatus(baseURL string) string {
	if baseURL == "" {
		return "not-configured"
	}

	resp, err := http.Get(baseURL + "/info")
	if err != nil || resp == nil {
		return "unreachable"
	}
	defer resp.Body.Close()

	return resp.Status
}

func createProxy(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		remote, err := url.Parse(target)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "invalid upstream URL"})
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = c.Request.URL.Path
			req.URL.RawQuery = c.Request.URL.RawQuery
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func jwtAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		// Pass user info to downstream services via headers
		if sub, ok := claims["sub"].(string); ok {
			c.Request.Header.Set("X-User-Id", sub)
		}
		if email, ok := claims["email"].(string); ok {
			c.Request.Header.Set("X-User-Email", email)
		}

		c.Next()
	}
}

func main() {
	r := gin.Default()

	messageURL := os.Getenv("MESSAGE_SERVICE_URL")
	presenceURL := os.Getenv("PRESENCE_SERVICE_URL")
	userURL := os.Getenv("USER_SERVICE_URL")
	notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	moderationURL := os.Getenv("MODERATION_SERVICE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

	// Health check
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"gateway":      "ok",
			"message":      safeServiceStatus(messageURL),
			"presence":     safeServiceStatus(presenceURL),
			"user":         safeServiceStatus(userURL),
			"notification": safeServiceStatus(notificationURL),
			"moderation":   safeServiceStatus(moderationURL),
		})
	})

	// Public auth routes (no JWT required)
	r.POST("/auth/register", createProxy(userURL))
	r.POST("/auth/login", createProxy(userURL))

	// Protected routes
	protected := r.Group("/")
	protected.Use(jwtAuthMiddleware(jwtSecret))
	{
		// Auth
		protected.GET("/auth/me", createProxy(userURL))

		// Message service
		protected.Any("/messages/*path", createProxy(messageURL))

		// Presence service
		protected.Any("/presence/*path", createProxy(presenceURL))

		// Notification service
		protected.Any("/notifications/*path", createProxy(notificationURL))

		// Moderation service
		protected.Any("/moderation/*path", createProxy(moderationURL))
	}

	// IMPORTANT : bind sur toutes les interfaces
	r.Run(":8080")
}
