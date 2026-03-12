package router

import (
"log"

"message-service/internal/handler"
"message-service/internal/middleware"
"message-service/internal/repository"

"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the message service
// Protected routes require JWT authentication from the gateway
func SetupRoutes(
r *gin.Engine,
healthHandler *handler.HealthHandler,
messageHandler *handler.MessageHandler,
conversationHandler *handler.ConversationHandler,
conversationRepo repository.ConversationRepository,
messageRepo repository.MessageRepository,
jwtSecret string,
) {
// Initialize authorization middleware
authMiddleware := middleware.NewAuthorizationMiddleware(conversationRepo, messageRepo)

// Assign authMiddleware to handlers
messageHandler.AuthMiddleware = authMiddleware
conversationHandler.AuthMiddleware = authMiddleware

// Apply global middleware
r.Use(middleware.LoggerMiddleware())

// Health routes (public)
healthGroup := r.Group("/")
{
healthGroup.GET("/info", healthHandler.GetInfo)
}

// Protected routes - apply JWT middleware
protected := r.Group("")
protected.Use(middleware.JWTMiddleware(jwtSecret))
{
// Conversation routes (all protected with authorization checks)
conversationGroup := protected.Group("/conversations")
{
conversationGroup.POST("", conversationHandler.CreateConversation)
conversationGroup.GET("/:id", conversationHandler.GetConversation)
conversationGroup.PUT("/:id", conversationHandler.UpdateConversation)
conversationGroup.DELETE("/:id", conversationHandler.DeleteConversation)
conversationGroup.GET("/:id/members", conversationHandler.GetConversationMembers)
conversationGroup.POST("/:id/members", conversationHandler.AddMember)
conversationGroup.DELETE("/:id/members/:user_id", conversationHandler.RemoveMember)
}

// User conversation routes (protected with authorization checks)
userGroup := protected.Group("/users")
{
userGroup.GET("/conversations", conversationHandler.GetUserConversations)
}

// Message routes (all protected with authorization checks)
messageGroup := protected.Group("/messages")
{
messageGroup.POST("", messageHandler.CreateMessage)
messageGroup.GET("/:id", messageHandler.GetMessage)
messageGroup.GET("/conversation/:conversation_id", messageHandler.GetMessages)
messageGroup.PUT("/:id", messageHandler.UpdateMessage)
messageGroup.DELETE("/:id", messageHandler.DeleteMessage)
messageGroup.GET("/user/:user_id", messageHandler.GetUserMessages)
}
}

// Debug: Log all registered routes
for _, route := range r.Routes() {
log.Printf("[ROUTE] %s %s", route.Method, route.Path)
}
}