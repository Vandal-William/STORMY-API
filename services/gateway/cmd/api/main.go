package main

import (
	"gateway-service/internal/config"
	"gateway-service/internal/handler"
	"gateway-service/internal/service"
	"gateway-service/internal/router"
	"gateway-service/pkg/client"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize HTTP client
	httpClient := client.NewHTTPClient(5 * time.Second)

	// Initialize services
	healthService := service.NewHealthService(httpClient, &cfg.Services)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(healthService)
	messageHandler := handler.NewMessageHandler(httpClient, cfg.Services.MessageURL)

	// Setup router
	r := gin.Default()
	router.SetupRoutes(r, healthHandler, messageHandler)

	// Start server
	log.Printf("Starting gateway on %s\n", cfg.Server.GetAddr())
	if err := r.Run(cfg.Server.GetAddr()); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}
