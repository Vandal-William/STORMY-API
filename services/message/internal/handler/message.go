package handler

import (
	"message-service/internal/domain"
	"message-service/internal/middleware"
	"message-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

// MessageHandler handles message-related requests
type MessageHandler struct {
	messageService      *service.MessageService
    conversationService *service.ConversationService
    AuthMiddleware      *middleware.AuthorizationMiddleware
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	messageService *service.MessageService,
    conversationService *service.ConversationService,
    authMiddleware *middleware.AuthorizationMiddleware,
	) *MessageHandler {
	return &MessageHandler{
		messageService:      messageService,
        conversationService: conversationService,
        AuthMiddleware:      authMiddleware,
	}
}

// CreateMessage creates a new message
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	var req domain.CreateMessageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.messageService.CreateMessage(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetMessages retrieves all messages for a conversation
func (h *MessageHandler) GetMessages(c *gin.Context) {
    conversationIDStr := c.Param("conversation_id")
    conversationID, err := gocql.ParseUUID(conversationIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID format"})
        return
    }

    messages, err := h.messageService.GetByConversationID(c.Request.Context(), conversationID, 50)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// GetMessage retrieves a message by ID
func (h *MessageHandler) GetMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := gocql.ParseUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message ID format"})
		return
	}

	message, err := h.messageService.GetMessage(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if message == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	c.JSON(http.StatusOK, message)
}

// GetUserMessages retrieves all messages for a user
func (h *MessageHandler) GetUserMessages(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userIDInt, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	messages, err := h.messageService.GetUserMessages(c.Request.Context(), int32(userIDInt), 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// UpdateMessage updates an existing message
func (h *MessageHandler) UpdateMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := gocql.ParseUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message ID format"})
		return
	}

	var req domain.UpdateMessageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.messageService.UpdateMessage(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, message)
}

// DeleteMessage deletes  a message
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := gocql.ParseUUID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message ID format"})
		return
	}

	err = h.messageService.DeleteMessage(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message deleted successfully"})
}
