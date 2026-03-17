package service

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"message-service/internal/domain"
	"message-service/internal/repository"
)

// MessageService handles business logic for messages
type MessageService struct {
	repo repository.MessageRepository
}

// NewMessageService creates a new message service
func NewMessageService(repo repository.MessageRepository) *MessageService {
	return &MessageService{
		repo: repo,
	}
}

// CreateMessage creates a new message
func (s *MessageService) CreateMessage(ctx context.Context, senderID string, req *domain.CreateMessageRequest) (*domain.Message, error) {
	var emptyUUID gocql.UUID
	if req.ConversationID == emptyUUID || senderID == "" || req.Content == "" {
		return nil, fmt.Errorf("invalid message data")
	}

	// Default type to "text" if not provided
	msgType := req.Type
	if msgType == "" {
		msgType = "text"
	}

	message := &domain.Message{
		ID:             gocql.TimeUUID(),
		ConversationID: req.ConversationID,
		SenderID:       senderID,
		Content:        req.Content,
		Type:           msgType,
		ReplyToID:      req.ReplyToID,
		IsForwarded:    false,
		IsEdited:       false,
		IsDeleted:      false,
	}

	return s.repo.Create(ctx, message)
}

// GetMessage retrieves a message by ID
func (s *MessageService) GetMessage(ctx context.Context, id gocql.UUID) (*domain.Message, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByConversationID retrieves messages from a conversation
func (s *MessageService) GetByConversationID(ctx context.Context, conversationID gocql.UUID, limit int) ([]domain.Message, error) {
	messages, _, err := s.repo.GetByConversationID(ctx, conversationID, limit, nil)
	return messages, err
}

// GetUserMessages retrieves all messages for a user
func (s *MessageService) GetUserMessages(ctx context.Context, userID string, limit int) ([]domain.Message, error) {
	return s.repo.GetByUserID(ctx, userID, limit)
}

// UpdateMessage updates an existing message
func (s *MessageService) UpdateMessage(ctx context.Context, id gocql.UUID, req *domain.UpdateMessageRequest) (*domain.Message, error) {
	msg, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, fmt.Errorf("message not found")
	}

	msg.Content = req.Content
	
	// Default type to "text" if not provided
	if req.Type != "" {
		msg.Type = req.Type
	} else {
		msg.Type = "text"
	}
	
	msg.IsEdited = true

	return s.repo.Update(ctx, id, msg)
}

// DeleteMessage deletes a message
func (s *MessageService) DeleteMessage(ctx context.Context, id gocql.UUID) error {
	msg, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if msg == nil {
		return fmt.Errorf("message not found")
	}

	return s.repo.Delete(ctx, id)
}
