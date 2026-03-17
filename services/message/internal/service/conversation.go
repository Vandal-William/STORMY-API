package service

import (
	"context"
	"fmt"
	"time"

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
func (s *ConversationService) CreateConversation(ctx context.Context, createdBy string, req *domain.CreateConversationRequest) (*domain.Conversation, error) {
	if createdBy == "" {
		return nil, fmt.Errorf("creator must be valid")
	}

	// Default type to "group" if not provided
	conversationType := req.Type
	if conversationType == "" {
		conversationType = "group"
	}

	conversation := &domain.Conversation{
		ID:        gocql.TimeUUID(),
		CreatedBy: createdBy,
		Name:      req.Name,
		Type:      conversationType,
		Description: req.Description,
		AvatarURL: req.AvatarURL,
	}

	// Create conversation
	createdConv, err := s.conversationRepo.Create(ctx, conversation)
	if err != nil {
		return nil, err
	}

	// Le créateur est déjà ajouté comme "owner" par le repository
	// Ajouter seulement les autres membres
	for _, memberID := range req.MemberIDs {
		member := &domain.ConversationMember{
			ID:             gocql.TimeUUID(),
			ConversationID: createdConv.ID,
			UserID:         memberID,
			Role:           "member",
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
func (s *ConversationService) GetUserConversations(ctx context.Context, userID string) ([]domain.Conversation, error) {
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
	
	if req.Description != "" {
		conv.Description = req.Description
	}
	
	if req.AvatarURL != "" {
		conv.AvatarURL = req.AvatarURL
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
func (s *ConversationService) AddMember(ctx context.Context, conversationID gocql.UUID, userID string) error {
	conv, err := s.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return err
	}

	if conv == nil {
		return fmt.Errorf("conversation not found")
	}

	member := &domain.ConversationMember{
		ID:             gocql.TimeUUID(),
		ConversationID: conversationID,
		UserID:         userID,
		Role:           "member",
		IsMuted:        false,
		JoinedAt:       time.Now(),
	}

	return s.conversationRepo.AddMember(ctx, member)
}

// RemoveMember removes a member from a conversation
func (s *ConversationService) RemoveMember(ctx context.Context, conversationID gocql.UUID, userID string) error {
	return s.conversationRepo.RemoveMember(ctx, conversationID, userID)
}
