package router

import (
	"gateway-service/internal/handler"
	"gateway-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the gateway
func SetupRoutes(r *gin.Engine, healthHandler *handler.HealthHandler, messageHandler *handler.MessageHandler) {
	// Apply global middleware
	r.Use(middleware.LoggerMiddleware())

	// Health routes (no auth required)
	healthGroup := r.Group("/")
	{
		healthGroup.GET("/info", healthHandler.GetInfo)
	}

	// Protected routes with JWT authentication
	protectedAPI := r.Group("")
	protectedAPI.Use(middleware.JWTMiddleware())
	{
		// Conversation routes
		conversationGroup := protectedAPI.Group("/conversations")
		{
			conversationGroup.POST("", messageHandler.CreateConversation)
			conversationGroup.GET("/:id", messageHandler.GetConversation)
			conversationGroup.PUT("/:id", messageHandler.UpdateConversation)
			conversationGroup.DELETE("/:id", messageHandler.DeleteConversation)
			conversationGroup.GET("/:id/members", messageHandler.GetConversationMembers)
			conversationGroup.POST("/:id/members", messageHandler.AddMember)
			conversationGroup.DELETE("/:id/members/:user_id", messageHandler.RemoveMember)
		}

		// User conversations routes
		userGroup := protectedAPI.Group("/users")
		{
			userGroup.GET("/conversations", messageHandler.GetUserConversations)
		}

		// Message routes
		messageGroup := protectedAPI.Group("/messages")
		{
			messageGroup.POST("", messageHandler.CreateMessage)
			messageGroup.GET("/:id", messageHandler.GetMessage)
			messageGroup.PUT("/:id", messageHandler.UpdateMessage)
			messageGroup.DELETE("/:id", messageHandler.DeleteMessage)
			messageGroup.GET("/user/messages", messageHandler.GetUserMessages)
		}
	}
}
