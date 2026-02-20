package router

import (
	"log"

	"message-service/internal/handler"
	"message-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the message service
func SetupRoutes(r *gin.Engine, healthHandler *handler.HealthHandler, messageHandler *handler.MessageHandler, conversationHandler *handler.ConversationHandler) {
	// Apply global middleware
	r.Use(middleware.LoggerMiddleware())

	// Health routes
	healthGroup := r.Group("/")
	{
		healthGroup.GET("/info", healthHandler.GetInfo)
	}

	// Conversation routes
	conversationGroup := r.Group("/conversations")
	{
		conversationGroup.POST("", conversationHandler.CreateConversation)
		conversationGroup.GET("/:id", conversationHandler.GetConversation)
		conversationGroup.PUT("/:id", conversationHandler.UpdateConversation)
		conversationGroup.DELETE("/:id", conversationHandler.DeleteConversation)
		conversationGroup.GET("/:id/members", conversationHandler.GetConversationMembers)
		conversationGroup.POST("/:id/members", conversationHandler.AddMember)
		conversationGroup.DELETE("/:id/members/:user_id", conversationHandler.RemoveMember)
	}

	// User conversation routes
	userGroup := r.Group("/users/:user_id")
	{
		userGroup.GET("/conversations", conversationHandler.GetUserConversations)
	}

	// Message routes
	messageGroup := r.Group("/messages")
	{
		messageGroup.POST("", messageHandler.CreateMessage)
		messageGroup.GET("/:id", messageHandler.GetMessage)
		messageGroup.PUT("/:id", messageHandler.UpdateMessage)
		messageGroup.DELETE("/:id", messageHandler.DeleteMessage)
		messageGroup.GET("/user/:user_id", messageHandler.GetUserMessages)
	}

	// Debug: Log all registered routes
	for _, route := range r.Routes() {
		log.Printf("[ROUTE] %s %s", route.Method, route.Path)
	}
}
