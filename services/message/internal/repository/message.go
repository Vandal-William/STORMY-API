package repository

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"message-service/internal/domain"
)

// MessageRepository définit l'interface pour l'accès aux messages
type MessageRepository interface {
	// Create stocke un nouveau message
	Create(ctx context.Context, message *domain.Message) (*domain.Message, error)

	// GetByID récupère un message par ID
	GetByID(ctx context.Context, id gocql.UUID) (*domain.Message, error)

	// GetByConversationID récupère les messages d'une conversation avec pagination
	GetByConversationID(ctx context.Context, conversationID gocql.UUID, limit int, pageState []byte) ([]domain.Message, []byte, error)

	// GetByUserID récupère tous les messages d'un utilisateur
	GetByUserID(ctx context.Context, userID string, limit int) ([]domain.Message, error)

	// Update modifie un message existant
	Update(ctx context.Context, id gocql.UUID, message *domain.Message) (*domain.Message, error)

	// Delete supprime (soft delete) un message
	Delete(ctx context.Context, id gocql.UUID) error

	// CreateAttachment crée une pièce jointe
	CreateAttachment(ctx context.Context, attachment *domain.MessageAttachment) error

	// DeleteAttachment supprime une pièce jointe
	DeleteAttachment(ctx context.Context, attachmentID gocql.UUID) error
}

// ConversationRepository définit l'interface pour l'accès aux conversations
type ConversationRepository interface {
	// Create crée une nouvelle conversation
	Create(ctx context.Context, conversation *domain.Conversation) (*domain.Conversation, error)

	// GetByID récupère une conversation par ID
	GetByID(ctx context.Context, id gocql.UUID) (*domain.Conversation, error)

	// GetByUserID récupère les conversations d'un utilisateur avec pagination
	GetByUserID(ctx context.Context, userID string) ([]domain.Conversation, error)

	// Update met à jour une conversation
	Update(ctx context.Context, id gocql.UUID, conversation *domain.Conversation) (*domain.Conversation, error)

	// Delete supprime une conversation
	Delete(ctx context.Context, id gocql.UUID) error

	// AddMember ajoute un membre à une conversation
	AddMember(ctx context.Context, member *domain.ConversationMember) error

	// RemoveMember retire un membre d'une conversation
	RemoveMember(ctx context.Context, conversationID gocql.UUID, userID string) error

	// GetMembers récupère les membres d'une conversation
	GetMembers(ctx context.Context, conversationID gocql.UUID) ([]domain.ConversationMember, error)

	// GetUserRoleInConversation récupère le rôle d'un utilisateur dans une conversation
	GetUserRoleInConversation(ctx context.Context, conversationID gocql.UUID, userID string) (string, error)
}

// ============== IN-MEMORY IMPLEMENTATIONS FOR DEVELOPMENT ==============

// InMemoryMessageRepository implémente MessageRepository en mémoire
type InMemoryMessageRepository struct {
	messages    map[string]*domain.Message
	attachments map[string]*domain.MessageAttachment
}

// NewInMemoryMessageRepository crée une nouvelle instance du repository en mémoire
func NewInMemoryMessageRepository() MessageRepository {
	return &InMemoryMessageRepository{
		messages:    make(map[string]*domain.Message),
		attachments: make(map[string]*domain.MessageAttachment),
	}
}

func (r *InMemoryMessageRepository) Create(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	id := message.ID.String()
	r.messages[id] = message
	return message, nil
}

func (r *InMemoryMessageRepository) GetByID(ctx context.Context, id gocql.UUID) (*domain.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if msg, ok := r.messages[id.String()]; ok {
		return msg, nil
	}
	return nil, nil
}

func (r *InMemoryMessageRepository) GetByConversationID(ctx context.Context, conversationID gocql.UUID, limit int, pageState []byte) ([]domain.Message, []byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	var messages []domain.Message
	for _, msg := range r.messages {
		if msg.ConversationID == conversationID && !msg.IsDeleted {
			messages = append(messages, *msg)
		}
	}
	return messages, nil, nil
}

func (r *InMemoryMessageRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]domain.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var messages []domain.Message
	for _, msg := range r.messages {
		if msg.SenderID == userID && !msg.IsDeleted {
			messages = append(messages, *msg)
		}
	}
	return messages, nil
}

func (r *InMemoryMessageRepository) Update(ctx context.Context, id gocql.UUID, message *domain.Message) (*domain.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	idStr := id.String()
	if _, ok := r.messages[idStr]; ok {
		now := time.Now()
		message.UpdatedAt = &now
		message.IsEdited = true
		r.messages[idStr] = message
		return message, nil
	}
	return nil, nil
}

func (r *InMemoryMessageRepository) Delete(ctx context.Context, id gocql.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	idStr := id.String()
	if msg, ok := r.messages[idStr]; ok {
		msg.IsDeleted = true
		now := time.Now()
		msg.UpdatedAt = &now
	}
	return nil
}

func (r *InMemoryMessageRepository) CreateAttachment(ctx context.Context, attachment *domain.MessageAttachment) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.attachments[attachment.ID.String()] = attachment
	return nil
}

func (r *InMemoryMessageRepository) DeleteAttachment(ctx context.Context, attachmentID gocql.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(r.attachments, attachmentID.String())
	return nil
}

// ============== IN-MEMORY CONVERSATION REPOSITORY ==============

// InMemoryConversationRepository implémente ConversationRepository en mémoire
type InMemoryConversationRepository struct {
	conversations map[string]*domain.Conversation
	members       map[string]*domain.ConversationMember
}

// NewInMemoryConversationRepository crée une nouvelle instance du repository
func NewInMemoryConversationRepository() ConversationRepository {
	return &InMemoryConversationRepository{
		conversations: make(map[string]*domain.Conversation),
		members:       make(map[string]*domain.ConversationMember),
	}
}

func (r *InMemoryConversationRepository) Create(ctx context.Context, conversation *domain.Conversation) (*domain.Conversation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	id := conversation.ID.String()
	now := time.Now()
	conversation.CreatedAt = now
	conversation.UpdatedAt = &now
	r.conversations[id] = conversation
	return conversation, nil
}

func (r *InMemoryConversationRepository) GetByID(ctx context.Context, id gocql.UUID) (*domain.Conversation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if conv, ok := r.conversations[id.String()]; ok {
		return conv, nil
	}
	return nil, nil
}

func (r *InMemoryConversationRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Conversation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var conversations []domain.Conversation
	for _, member := range r.members {
		if member.UserID == userID {
			if conv, ok := r.conversations[member.ConversationID.String()]; ok {
				conversations = append(conversations, *conv)
			}
		}
	}
	return conversations, nil
}

func (r *InMemoryConversationRepository) Update(ctx context.Context, id gocql.UUID, conversation *domain.Conversation) (*domain.Conversation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	idStr := id.String()
	if _, ok := r.conversations[idStr]; ok {
		now := time.Now()
		conversation.UpdatedAt = &now
		r.conversations[idStr] = conversation
		return conversation, nil
	}
	return nil, nil
}

func (r *InMemoryConversationRepository) Delete(ctx context.Context, id gocql.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(r.conversations, id.String())
	return nil
}

func (r *InMemoryConversationRepository) AddMember(ctx context.Context, member *domain.ConversationMember) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	key := member.ConversationID.String() + ":" + member.UserID
	r.members[key] = member
	return nil
}

func (r *InMemoryConversationRepository) RemoveMember(ctx context.Context, conversationID gocql.UUID, userID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	key := conversationID.String() + ":" + userID
	if member, ok := r.members[key]; ok {
		now := time.Now()
		member.LeftAt = &now
	}
	return nil
}

func (r *InMemoryConversationRepository) GetMembers(ctx context.Context, conversationID gocql.UUID) ([]domain.ConversationMember, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var members []domain.ConversationMember
	convIDStr := conversationID.String()
	for _, member := range r.members {
		if member.ConversationID.String() == convIDStr && member.LeftAt == nil {
			members = append(members, *member)
		}
	}
	return members, nil
}

// GetUserRoleInConversation récupère le rôle d'un utilisateur dans une conversation
func (r *InMemoryConversationRepository) GetUserRoleInConversation(ctx context.Context, conversationID gocql.UUID, userID string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	convIDStr := conversationID.String()
	for _, member := range r.members {
		if member.ConversationID.String() == convIDStr && member.UserID == userID && member.LeftAt == nil {
			return member.Role, nil
		}
	}
	return "", nil // Not found
}
