package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"message-service/internal/domain"
	"message-service/internal/repository"
	"message-service/internal/service"
)

func setupConversationTest() (*ConversationHandler, context.Context) {
	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	conversationService := service.NewConversationService(conversationRepo, messageRepo)
	handler := NewConversationHandler(conversationService)
	return handler, context.Background()
}

func TestCreateConversationHandler(t *testing.T) {
	handler, _ := setupConversationTest()

	member1 := gocql.TimeUUID().String()
	member2 := gocql.TimeUUID().String()
	creator := gocql.TimeUUID().String()

	body := map[string]interface{}{
		"name":      "Test Conversation",
		"type":      "group",
		"created_by": creator,
		"member_ids": []string{member1, member2},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/conversations", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateConversation(c)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestGetConversationHandler(t *testing.T) {
	handler, ctx := setupConversationTest()

	// Create a conversation first
	creator := gocql.TimeUUID()
	member1 := gocql.TimeUUID()
	member2 := gocql.TimeUUID()

	req := &domain.CreateConversationRequest{
		Name:      "Test Conversation",
		Type:      "group",
		CreatedBy: creator,
		MemberIDs: []gocql.UUID{member1, member2},
	}

	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	conversationService := service.NewConversationService(conversationRepo, messageRepo)
	conversation, _ := conversationService.CreateConversation(ctx, req)

	// Now test the handler
	httpReq := httptest.NewRequest("GET", "/conversations/"+conversation.ID.String(), nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: conversation.ID.String()}}

	handler.GetConversation(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetConversationMembersHandler(t *testing.T) {
	handler, ctx := setupConversationTest()

	// Create a conversation first
	creator := gocql.TimeUUID()
	member1 := gocql.TimeUUID()
	member2 := gocql.TimeUUID()

	req := &domain.CreateConversationRequest{
		Name:      "Test Conversation",
		Type:      "group",
		CreatedBy: creator,
		MemberIDs: []gocql.UUID{member1, member2},
	}

	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	conversationService := service.NewConversationService(conversationRepo, messageRepo)
	conversation, _ := conversationService.CreateConversation(ctx, req)

	// Now test the handler
	httpReq := httptest.NewRequest("GET", "/conversations/"+conversation.ID.String()+"/members", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: conversation.ID.String()}}

	handler.GetConversationMembers(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteConversationHandler(t *testing.T) {
	handler, ctx := setupConversationTest()

	// Create a conversation first
	creator := gocql.TimeUUID()
	member1 := gocql.TimeUUID()
	member2 := gocql.TimeUUID()

	req := &domain.CreateConversationRequest{
		Name:      "Test Conversation",
		Type:      "group",
		CreatedBy: creator,
		MemberIDs: []gocql.UUID{member1, member2},
	}

	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	conversationService := service.NewConversationService(conversationRepo, messageRepo)
	conversation, _ := conversationService.CreateConversation(ctx, req)

	// Now test the handler
	httpReq := httptest.NewRequest("DELETE", "/conversations/"+conversation.ID.String(), nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: conversation.ID.String()}}

	handler.DeleteConversation(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
