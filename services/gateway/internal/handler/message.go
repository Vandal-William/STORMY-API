package handler

import (
	"gateway-service/internal/middleware"
	"gateway-service/pkg/client"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MessageHandler handles message service proxy requests
type MessageHandler struct {
	httpClient *client.HTTPClient
	messageURL string
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(httpClient *client.HTTPClient, messageURL string) *MessageHandler {
	return &MessageHandler{
		httpClient: httpClient,
		messageURL: messageURL,
	}
}

// CreateConversation proxies POST /conversations request to message service
// @Summary      Create a new conversation
// @Description  Creates a new conversation with the specified title and members
// @Tags         conversations
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        body body ConversationRequest true "Conversation data"
// @Success      201  {object} map[string]interface{}
// @Failure      400  {object} ErrorResponse
// @Failure      401  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /conversations [post]
func (h *MessageHandler) CreateConversation(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Forward request to message service
	resp, err := h.httpClient.Do("POST", h.messageURL+"/conversations", 
		c.Request.Header.Get("Content-Type"), body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create conversation"})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// GetConversation proxies GET /conversations/:id request to message service
// @Summary      Get conversation details
// @Description  Retrieves the details of a specific conversation
// @Tags         conversations
// @Security     CookieAuth
// @Produce      json
// @Param        id path string true "Conversation ID"
// @Success      200  {object} map[string]interface{}
// @Failure      401  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /conversations/{id} [get]
func (h *MessageHandler) GetConversation(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	conversationID := c.Param("id")
	
	resp, err := h.httpClient.Do("GET", h.messageURL+"/conversations/"+conversationID, 
		"application/json", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get conversation"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// UpdateConversation proxies PUT /conversations/:id request to message service
// @Summary      Update conversation
// @Description  Updates a specific conversation
// @Tags         conversations
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Conversation ID"
// @Param        body body ConversationRequest true "Updated conversation data"
// @Success      200  {object} map[string]interface{}
// @Failure      400  {object} ErrorResponse
// @Failure      401  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /conversations/{id} [put]
func (h *MessageHandler) UpdateConversation(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	conversationID := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	resp, err := h.httpClient.Do("PUT", h.messageURL+"/conversations/"+conversationID, 
		c.Request.Header.Get("Content-Type"), body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update conversation"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// DeleteConversation proxies DELETE /conversations/:id request to message service
// @Summary      Delete conversation
// @Description  Deletes a specific conversation
// @Tags         conversations
// @Security     CookieAuth
// @Param        id path string true "Conversation ID"
// @Success      204
// @Failure      401  {object} ErrorResponse
// @Failure      500  {object} ErrorResponse
// @Router       /conversations/{id} [delete]
func (h *MessageHandler) DeleteConversation(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	conversationID := c.Param("id")
	
	resp, err := h.httpClient.Do("DELETE", h.messageURL+"/conversations/"+conversationID, 
		"application/json", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete conversation"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// GetUserConversations proxies GET /users/:user_id/conversations request to message service
func (h *MessageHandler) GetUserConversations(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Always fetch the authenticated user's conversations
	resp, err := h.httpClient.Do("GET", h.messageURL+"/users/"+userID+"/conversations", 
		"application/json", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get conversations"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// GetConversationMembers proxies GET /conversations/:id/members request
func (h *MessageHandler) GetConversationMembers(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	conversationID := c.Param("id")
	
	resp, err := h.httpClient.Do("GET", h.messageURL+"/conversations/"+conversationID+"/members", 
		"application/json", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get members"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// AddMember proxies POST /conversations/:id/members request
func (h *MessageHandler) AddMember(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	conversationID := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	resp, err := h.httpClient.Do("POST", h.messageURL+"/conversations/"+conversationID+"/members", 
		c.Request.Header.Get("Content-Type"), body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// RemoveMember proxies DELETE /conversations/:id/members/:user_id request
func (h *MessageHandler) RemoveMember(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	conversationID := c.Param("id")
	memberUserID := c.Param("user_id")
	
	resp, err := h.httpClient.Do("DELETE", 
		h.messageURL+"/conversations/"+conversationID+"/members/"+memberUserID, 
		"application/json", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove member"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// CreateMessage proxies POST /messages request to message service
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Ensure sender_id matches authenticated user
	// TODO: Validate and enforce sender_id = authenticated user_id

	resp, err := h.httpClient.Do("POST", h.messageURL+"/messages", 
		c.Request.Header.Get("Content-Type"), body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// GetMessage proxies GET /messages/:id request to message service
func (h *MessageHandler) GetMessage(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	messageID := c.Param("id")
	
	resp, err := h.httpClient.Do("GET", h.messageURL+"/messages/"+messageID, 
		"application/json", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get message"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// UpdateMessage proxies PUT /messages/:id request to message service
func (h *MessageHandler) UpdateMessage(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	messageID := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	resp, err := h.httpClient.Do("PUT", h.messageURL+"/messages/"+messageID, 
		c.Request.Header.Get("Content-Type"), body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update message"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// DeleteMessage proxies DELETE /messages/:id request to message service
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	messageID := c.Param("id")
	
	resp, err := h.httpClient.Do("DELETE", h.messageURL+"/messages/"+messageID, 
		"application/json", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete message"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// GetUserMessages proxies GET /users/:user_id/messages request
func (h *MessageHandler) GetUserMessages(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Always fetch the authenticated user's messages
	resp, err := h.httpClient.Do("GET", h.messageURL+"/users/"+userID+"/messages", 
		"application/json", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}
