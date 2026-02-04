package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	messageURL := os.Getenv("MESSAGE_SERVICE_URL")
	presenceURL := os.Getenv("PRESENCE_SERVICE_URL")
	userURL := os.Getenv("USER_SERVICE_URL")
	notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	moderationURL := os.Getenv("MODERATION_SERVICE_URL")

	// Route publique
	r.GET("/info", func(c *gin.Context) {

		messageResp, _ := http.Get(messageURL + "/info")
		presenceResp, _ := http.Get(presenceURL + "/info")
		userResp, _ := http.Get(userURL + "/info")
		notificationResp, _ := http.Get(notificationURL + "/info")
		moderationResp, _ := http.Get(moderationURL + "/info")

		c.JSON(http.StatusOK, gin.H{
			"gateway":  "ok",
			"message":  messageResp.Status,
			"presence": presenceResp.Status,
			"user":     userResp.Status,
			"notification": notificationResp.Status,
			"moderation":   moderationResp.Status,
		})
	})

	r.Run(":8080")
}
