package service

import (
	"context"
	"testing"

	"github.com/gocql/gocql"
	"message-service/internal/domain"
	"message-service/internal/repository"
)

func TestCreateConversation(t *testing.T) {
	ctx := context.Background()
	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	service := NewConversationService(conversationRepo, messageRepo)

	// Create member IDs
	member1 := gocql.TimeUUID()
	member2 := gocql.TimeUUID()
	creator := gocql.TimeUUID()

	req := &domain.CreateConversationRequest{
		Name:      "Test Conversation",
		Type:      "group",
		CreatedBy: creator,
		MemberIDs: []gocql.UUID{member1, member2},
	}

	conversation, err := service.CreateConversation(ctx, req)
	if err != nil {
		t.Fatalf("CreateConversation failed: %v", err)
	}

	if conversation == nil {
		t.Fatal("Expected conversation, got nil")
	}

	if conversation.Name != "Test Conversation" {
		t.Errorf("Expected name 'Test Conversation', got '%s'", conversation.Name)
	}

	if conversation.Type != "group" {
		t.Errorf("Expected type 'group', got '%s'", conversation.Type)
	}

	if conversation.CreatedBy != creator {
		t.Errorf("Expected creator %s, got %s", creator, conversation.CreatedBy)
	}

	if !conversation.IsActive {
		t.Error("Expected conversation to be active")
	}
}

func TestGetConversation(t *testing.T) {
	ctx := context.Background()
	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	service := NewConversationService(conversationRepo, messageRepo)

	// Create a conversation first
	member1 := gocql.TimeUUID()
	member2 := gocql.TimeUUID()
	creator := gocql.TimeUUID()

	req := &domain.CreateConversationRequest{
		Name:      "Test Conversation",
		Type:      "direct",
		CreatedBy: creator,
		MemberIDs: []gocql.UUID{member1, member2},
	}

	created, _ := service.CreateConversation(ctx, req)

	// Now retrieve it
	conversation, err := service.GetConversation(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetConversation failed: %v", err)
	}

	if conversation == nil {
		t.Fatal("Expected conversation, got nil")
	}

	if conversation.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, conversation.ID)
	}
}

func TestGetUserConversations(t *testing.T) {
	ctx := context.Background()
	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	service := NewConversationService(conversationRepo, messageRepo)

	creator := gocql.TimeUUID()
	member := gocql.TimeUUID()

	// Create multiple conversations
	for i := 0; i < 3; i++ {
		req := &domain.CreateConversationRequest{
			Name:      "Test Conversation",
			Type:      "group",
			CreatedBy: creator,
			MemberIDs: []gocql.UUID{member, gocql.TimeUUID()},
		}
		service.CreateConversation(ctx, req)
	}

	// Retrieve user conversations
	conversations, err := service.GetUserConversations(ctx, creator)
	if err != nil {
		t.Fatalf("GetUserConversations failed: %v", err)
	}

	if len(conversations) != 3 {
		t.Errorf("Expected 3 conversations, got %d", len(conversations))
	}
}

func TestAddMember(t *testing.T) {
	ctx := context.Background()
	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	service := NewConversationService(conversationRepo, messageRepo)

	creator := gocql.TimeUUID()
	member1 := gocql.TimeUUID()
	member2 := gocql.TimeUUID()
	newMember := gocql.TimeUUID()

	// Create conversation
	req := &domain.CreateConversationRequest{
		Name:      "Test Conversation",
		Type:      "group",
		CreatedBy: creator,
		MemberIDs: []gocql.UUID{member1, member2},
	}
	conversation, _ := service.CreateConversation(ctx, req)

	// Add new member
	err := service.AddMember(ctx, conversation.ID, newMember)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	// Verify member was added
	members, err := service.GetConversationMembers(ctx, conversation.ID)
	if err != nil {
		t.Fatalf("GetConversationMembers failed: %v", err)
	}

	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}

	// Check if new member is in the list
	found := false
	for _, m := range members {
		if m.UserID == newMember {
			found = true
			break
		}
	}

	if !found {
		t.Error("New member not found in conversation")
	}
}

func TestRemoveMember(t *testing.T) {
	ctx := context.Background()
	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	service := NewConversationService(conversationRepo, messageRepo)

	creator := gocql.TimeUUID()
	member1 := gocql.TimeUUID()
	member2 := gocql.TimeUUID()

	// Create conversation
	req := &domain.CreateConversationRequest{
		Name:      "Test Conversation",
		Type:      "group",
		CreatedBy: creator,
		MemberIDs: []gocql.UUID{member1, member2},
	}
	conversation, _ := service.CreateConversation(ctx, req)

	// Remove member
	err := service.RemoveMember(ctx, conversation.ID, member2)
	if err != nil {
		t.Fatalf("RemoveMember failed: %v", err)
	}

	// Verify member was removed
	members, _ := service.GetConversationMembers(ctx, conversation.ID)

	found := false
	for _, m := range members {
		if m.UserID == member2 {
			found = true
			break
		}
	}

	if found {
		t.Error("Member should have been removed but is still in conversation")
	}
}

func TestDeleteConversation(t *testing.T) {
	ctx := context.Background()
	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	service := NewConversationService(conversationRepo, messageRepo)

	creator := gocql.TimeUUID()
	member1 := gocql.TimeUUID()
	member2 := gocql.TimeUUID()

	// Create conversation
	req := &domain.CreateConversationRequest{
		Name:      "Test Conversation",
		Type:      "group",
		CreatedBy: creator,
		MemberIDs: []gocql.UUID{member1, member2},
	}
	conversation, _ := service.CreateConversation(ctx, req)

	// Delete conversation
	err := service.DeleteConversation(ctx, conversation.ID)
	if err != nil {
		t.Fatalf("DeleteConversation failed: %v", err)
	}

	// Verify it's deleted
	retrieved, _ := service.GetConversation(ctx, conversation.ID)
	if retrieved != nil {
		t.Error("Conversation should be deleted but still exists")
	}
}

func TestUpdateConversation(t *testing.T) {
	ctx := context.Background()
	conversationRepo := repository.NewInMemoryConversationRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	service := NewConversationService(conversationRepo, messageRepo)

	creator := gocql.TimeUUID()
	member1 := gocql.TimeUUID()
	member2 := gocql.TimeUUID()

	// Create conversation
	req := &domain.CreateConversationRequest{
		Name:      "Test Conversation",
		Type:      "group",
		CreatedBy: creator,
		MemberIDs: []gocql.UUID{member1, member2},
	}
	conversation, _ := service.CreateConversation(ctx, req)

	// Update conversation
	updateReq := &domain.UpdateConversationRequest{
		Name: "Updated Conversation",
	}
	updated, err := service.UpdateConversation(ctx, conversation.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateConversation failed: %v", err)
	}

	if updated.Name != "Updated Conversation" {
		t.Errorf("Expected name 'Updated Conversation', got '%s'", updated.Name)
	}
}
