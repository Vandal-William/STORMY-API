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
func (s *MessageService) CreateMessage(ctx context.Context, req *domain.CreateMessageRequest) (*domain.Message, error) {
	var emptyUUID gocql.UUID
	if req.ConversationID == emptyUUID || req.SenderID == 0 || req.Content == "" {
		return nil, fmt.Errorf("invalid message data")
	}

	message := &domain.Message{
		ID:             gocql.TimeUUID(),
		ConversationID: req.ConversationID,
		SenderID:       req.SenderID,
		Content:        req.Content,
		Type:           req.Type,
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
func (s *MessageService) GetUserMessages(ctx context.Context, userID int32, limit int) ([]domain.Message, error) {
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
