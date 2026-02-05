package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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

func main() {
	r := gin.Default()

	messageURL := os.Getenv("MESSAGE_SERVICE_URL")
	presenceURL := os.Getenv("PRESENCE_SERVICE_URL")
	userURL := os.Getenv("USER_SERVICE_URL")
	notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	moderationURL := os.Getenv("MODERATION_SERVICE_URL")

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

	// IMPORTANT : bind sur toutes les interfaces
	r.Run(":8080")
}
