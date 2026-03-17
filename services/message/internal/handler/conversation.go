package handler

import (
	"fmt"
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

	// Extract creator ID from JWT context
	claims, err := middleware.GetClaimsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid JWT claims"})
		return
	}

	if claims.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user ID in JWT"})
		return
	}

	// Pass request and createdBy to service
	conversation, err := h.conversationService.CreateConversation(c.Request.Context(), claims.UserID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	fmt.Printf("[DEBUG] GetUserConversations - claims.UserID: %s\n", claims.UserID)

	conversations, err := h.conversationService.GetUserConversations(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[DEBUG] GetUserConversations - résultats: %d conversations trouvées\n", len(conversations))

	// ✅ Assurer que la liste n'est jamais nil (retourner [] au lieu de null)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	// Extract user ID from JWT
	claims, err := middleware.GetClaimsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid JWT claims"})
		return
	}

	if claims.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user ID in JWT"})
		return
	}

	// Verify user is owner of the conversation
	role, err := h.AuthMiddleware.GetConversationRepository().GetUserRoleInConversation(c.Request.Context(), id, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to verify user role: %v", err)})
		return
	}

	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "user is not a member of this conversation"})
		return
	}

	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only conversation owner can add members"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	// Extract user ID from JWT
	claims, err := middleware.GetClaimsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid JWT claims"})
		return
	}

	if claims.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user ID in JWT"})
		return
	}

	// Verify user is owner of the conversation
	role, err := h.AuthMiddleware.GetConversationRepository().GetUserRoleInConversation(c.Request.Context(), id, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to verify user role: %v", err)})
		return
	}

	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "user is not a member of this conversation"})
		return
	}

	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only conversation owner can remove members"})
		return
	}

	userIDStr := c.Param("user_id")  // UUID string
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user ID"})
		return
	}

	err = h.conversationService.RemoveMember(c.Request.Context(), id, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}
