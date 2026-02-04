package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "message",
			"status":  "ok",
		})
	})

	r.Run(":8080")
}
