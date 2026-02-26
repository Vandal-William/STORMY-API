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
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	_ "gateway-service/docs"
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
	authHandler := handler.NewAuthHandler(httpClient, cfg.Services.UserURL)

	// Setup router
	r := gin.Default()
	
	// Setup Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/swagger/index.html")
	})
	
	router.SetupRoutes(r, healthHandler, messageHandler, authHandler, cfg.JWT.Secret)

	// Start server
	log.Printf("Starting gateway on %s\n", cfg.Server.GetAddr())
	if err := r.Run(cfg.Server.GetAddr()); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}
