package unit

import (
	"message-service/internal/domain"
	"message-service/internal/repository"
	"message-service/internal/service"
	"testing"
)

func TestCreateMessage(t *testing.T) {
	// Setup
	repo := repository.NewInMemoryMessageRepository()
	svc := service.NewMessageService(repo)

	req := &domain.CreateMessageRequest{
		SenderID:   "user1",
		ReceiverID: "user2",
		Content:    "Hello, World!",
	}

	// Execute
	msg, err := svc.CreateMessage(nil, req)

	// Verify
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if msg == nil {
		t.Error("expected message, got nil")
	}

	if msg.SenderID != "user1" {
		t.Errorf("expected sender_id 'user1', got %s", msg.SenderID)
	}

	if msg.ReceiverID != "user2" {
		t.Errorf("expected receiver_id 'user2', got %s", msg.ReceiverID)
	}

	if msg.Content != "Hello, World!" {
		t.Errorf("expected content 'Hello, World!', got %s", msg.Content)
	}
}

func TestGetMessage(t *testing.T) {
	// Setup
	repo := repository.NewInMemoryMessageRepository()
	svc := service.NewMessageService(repo)

	// Create a message first
	req := &domain.CreateMessageRequest{
		SenderID:   "user1",
		ReceiverID: "user2",
		Content:    "Hello",
	}

	created, _ := svc.CreateMessage(nil, req)

	// Execute
	retrieved, err := svc.GetMessage(nil, created.ID)

	// Verify
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if retrieved == nil {
		t.Error("expected message, got nil")
	}

	if retrieved.ID != created.ID {
		t.Errorf("expected id %s, got %s", created.ID, retrieved.ID)
	}
}

func TestDeleteMessage(t *testing.T) {
	// Setup
	repo := repository.NewInMemoryMessageRepository()
	svc := service.NewMessageService(repo)

	// Create a message
	req := &domain.CreateMessageRequest{
		SenderID:   "user1",
		ReceiverID: "user2",
		Content:    "Hello",
	}

	created, _ := svc.CreateMessage(nil, req)

	// Execute delete
	err := svc.DeleteMessage(nil, created.ID)

	// Verify
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify it's deleted
	retrieved, _ := svc.GetMessage(nil, created.ID)
	if retrieved != nil {
		t.Error("expected message to be deleted, but it still exists")
	}
}
