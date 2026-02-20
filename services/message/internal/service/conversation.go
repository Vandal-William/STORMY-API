package service

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"message-service/internal/domain"
	"message-service/internal/repository"
)

// ConversationService handles business logic for conversations
type ConversationService struct {
	conversationRepo repository.ConversationRepository
	messageRepo      repository.MessageRepository
}

// NewConversationService creates a new conversation service
func NewConversationService(
	conversationRepo repository.ConversationRepository,
	messageRepo repository.MessageRepository,
) *ConversationService {
	return &ConversationService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
	}
}

// CreateConversation creates a new conversation
func (s *ConversationService) CreateConversation(ctx context.Context, req *domain.CreateConversationRequest) (*domain.Conversation, error) {
	if len(req.MemberIDs) < 2 || req.CreatedBy == 0 {
		return nil, fmt.Errorf("conversation must have at least 2 members and a valid creator")
	}

	conversation := &domain.Conversation{
		ID:        gocql.TimeUUID(),
		CreatedBy: req.CreatedBy,
		Name:      req.Name,
		Type:      req.Type,
		Description: req.Description,
		AvatarURL: req.AvatarURL,
	}

	// Create conversation
	createdConv, err := s.conversationRepo.Create(ctx, conversation)
	if err != nil {
		return nil, err
	}

	// Add members
	for _, memberID := range req.MemberIDs {
		member := &domain.ConversationMember{
			ConversationID: createdConv.ID,
			UserID:         memberID,
			JoinedAt:       createdConv.CreatedAt,
		}
		if err := s.conversationRepo.AddMember(ctx, member); err != nil {
			return nil, err
		}
	}

	return createdConv, nil
}

// GetConversation retrieves a conversation by ID
func (s *ConversationService) GetConversation(ctx context.Context, id gocql.UUID) (*domain.Conversation, error) {
	return s.conversationRepo.GetByID(ctx, id)
}

// GetUserConversations retrieves all conversations for a user
func (s *ConversationService) GetUserConversations(ctx context.Context, userID int32) ([]domain.Conversation, error) {
	return s.conversationRepo.GetByUserID(ctx, userID)
}

// GetConversationMembers retrieves all members of a conversation
func (s *ConversationService) GetConversationMembers(ctx context.Context, conversationID gocql.UUID) ([]domain.ConversationMember, error) {
	return s.conversationRepo.GetMembers(ctx, conversationID)
}

// UpdateConversation updates a conversation
func (s *ConversationService) UpdateConversation(ctx context.Context, id gocql.UUID, req *domain.UpdateConversationRequest) (*domain.Conversation, error) {
	conv, err := s.conversationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	if req.Name != "" {
		conv.Name = req.Name
	}

	return s.conversationRepo.Update(ctx, id, conv)
}

// DeleteConversation deletes a conversation
func (s *ConversationService) DeleteConversation(ctx context.Context, id gocql.UUID) error {
	conv, err := s.conversationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if conv == nil {
		return fmt.Errorf("conversation not found")
	}

	return s.conversationRepo.Delete(ctx, id)
}

// AddMember adds a member to a conversation
func (s *ConversationService) AddMember(ctx context.Context, conversationID gocql.UUID, userID int32) error {
	conv, err := s.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return err
	}

	if conv == nil {
		return fmt.Errorf("conversation not found")
	}

	member := &domain.ConversationMember{
		ConversationID: conversationID,
		UserID:         userID,
	}

	return s.conversationRepo.AddMember(ctx, member)
}

// RemoveMember removes a member from a conversation
func (s *ConversationService) RemoveMember(ctx context.Context, conversationID gocql.UUID, userID int32) error {
	return s.conversationRepo.RemoveMember(ctx, conversationID, userID)
}
