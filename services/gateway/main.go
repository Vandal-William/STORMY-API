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

	// Route publique
	r.GET("/info", func(c *gin.Context) {

		messageResp, _ := http.Get(messageURL + "/info")
		presenceResp, _ := http.Get(presenceURL + "/info")

		c.JSON(http.StatusOK, gin.H{
			"gateway":  "ok",
			"message":  messageResp.Status,
			"presence": presenceResp.Status,
		})
	})

	r.Run(":8080")
}
