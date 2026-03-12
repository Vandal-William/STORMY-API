package main

import (
    "fmt"
    "log"
    "os"

    "message-service/internal/config"
    "message-service/internal/handler"
    "message-service/internal/infrastructure/cassandra"
    "message-service/internal/middleware"
    "message-service/internal/repository"
    "message-service/internal/router"
    "message-service/internal/service"

    "github.com/gin-gonic/gin"
)

func main() {
    log.SetOutput(os.Stderr)
    fmt.Fprintf(os.Stderr, "[STARTUP] Loading configuration\n")
    cfg := config.Load()
    fmt.Fprintf(os.Stderr, "[STARTUP] Attempting to create Cassandra client\n")
    
    cassandraClient, err := cassandra.NewClient(
        cassandra.Config{
            Hosts:    cfg.Cassandra.Hosts,
            Port:     cfg.Cassandra.Port,
            Keyspace: cfg.Cassandra.Keyspace,
        },
    )

    var messageRepo repository.MessageRepository
    var conversationRepo repository.ConversationRepository

    if err != nil {
        fmt.Fprintf(os.Stderr, "[STARTUP] WARNING: Cassandra failed: %v, using in-memory\n", err)
        messageRepo = repository.NewInMemoryMessageRepository()
        conversationRepo = repository.NewInMemoryConversationRepository()
    } else {
        defer cassandraClient.Close()
        fmt.Fprintf(os.Stderr, "[STARTUP] Cassandra connected\n")
        messageRepo = repository.NewCassandraMessageRepository(cassandraClient)
        conversationRepo = repository.NewCassandraConversationRepository(cassandraClient)
    }

    messageService := service.NewMessageService(messageRepo)
    conversationService := service.NewConversationService(conversationRepo, messageRepo)

    // Create authorization middleware
    authMiddleware := middleware.NewAuthorizationMiddleware(conversationRepo, messageRepo)

    healthHandler := handler.NewHealthHandler()
    messageHandler := handler.NewMessageHandler(messageService, conversationService, authMiddleware)
    conversationHandler := handler.NewConversationHandler(conversationService, authMiddleware)

    // Get JWT secret from environment
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "dev-secret-change-in-production"
    }

    fmt.Fprintf(os.Stderr, "[STARTUP] Setting up routes\n")
    r := gin.Default()
    router.SetupRoutes(r, healthHandler, messageHandler, conversationHandler, conversationRepo, messageRepo, jwtSecret)
    fmt.Fprintf(os.Stderr, "[STARTUP] Routes configured\n")

    fmt.Fprintf(os.Stderr, "[STARTUP] Starting server on %s\n", cfg.Server.GetAddr())
    if err := r.Run(cfg.Server.GetAddr()); err != nil {
        log.Fatalf("Server error: %v\n", err)
    }
}