package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"message-service/internal/domain"
	"message-service/internal/middleware"
	"message-service/internal/service"
)

// ConversationHandler handles conversation-related HTTP requests
type ConversationHandler struct {
	conversationService *service.ConversationService
	AuthMiddleware      *middleware.AuthorizationMiddleware
}

// NewConversationHandler creates a new conversation handler
func NewConversationHandler(
	conversationService *service.ConversationService,
	authMiddleware *middleware.AuthorizationMiddleware,
	) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
		AuthMiddleware:      authMiddleware,
	}
}

// CreateConversation handles POST /conversations
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	var req domain.CreateConversationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation, err := h.conversationService.CreateConversation(c.Request.Context(), &req)
	if err != nil {
		log.Printf("[ERROR] CreateConversation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, conversation)
}

// GetConversation handles GET /conversations/:id
func (h *ConversationHandler) GetConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := gocql.ParseUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID format"})
		return
	}

	conversation, err := h.conversationService.GetConversation(c.Request.Context(), id)
	if err != nil {
		log.Printf("[ERROR] GetConversation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if conversation == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

// GetUserConversations handles GET /users/conversations (authenticated)
func (h *ConversationHandler) GetUserConversations(c *gin.Context) {
	// Get userID from JWT context (UUID string from gateway)
	claims, err := middleware.GetClaimsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid JWT claims"})
		return
	}

	if claims.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user ID in JWT"})
		return
	}

	conversations, err := h.conversationService.GetUserConversations(c.Request.Context(), claims.UserID)
	if err != nil {
		log.Printf("[ERROR] GetUserConversations failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Assurer que la liste n'est jamais nil (retourner [] au lieu de null)
	if conversations == nil {
		conversations = []domain.Conversation{}
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// GetConversationMembers handles GET /conversations/:id/members
func (h *ConversationHandler) GetConversationMembers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := gocql.ParseUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID format"})
		return
	}

	members, err := h.conversationService.GetConversationMembers(c.Request.Context(), id)
	if err != nil {
		log.Printf("[ERROR] GetConversationMembers failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// ✅ Assurer que la liste n'est jamais nil
	if members == nil {
		members = []domain.ConversationMember{}
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// UpdateConversation handles PUT /conversations/:id
func (h *ConversationHandler) UpdateConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := gocql.ParseUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID format"})
		return
	}

	var req domain.UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation, err := h.conversationService.UpdateConversation(c.Request.Context(), id, &req)
	if err != nil {
		log.Printf("[ERROR] UpdateConversation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

// DeleteConversation handles DELETE /conversations/:id
func (h *ConversationHandler) DeleteConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := gocql.ParseUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID format"})
		return
	}

	err = h.conversationService.DeleteConversation(c.Request.Context(), id)
	if err != nil {
		log.Printf("[ERROR] DeleteConversation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted successfully"})
}

// AddMember handles POST /conversations/:id/members
func (h *ConversationHandler) AddMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := gocql.ParseUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID format"})
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`  // UUID string
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.conversationService.AddMember(c.Request.Context(), id, req.UserID)
	if err != nil {
		log.Printf("[ERROR] AddMember failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member added successfully"})
}

// RemoveMember handles DELETE /conversations/:id/members/:user_id
func (h *ConversationHandler) RemoveMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := gocql.ParseUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID format"})
		return
	}

	userIDStr := c.Param("user_id")  // UUID string
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user ID"})
		return
	}

	err = h.conversationService.RemoveMember(c.Request.Context(), id, userIDStr)
	if err != nil {
		log.Printf("[ERROR] RemoveMember failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}
