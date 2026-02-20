package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"message-service/internal/domain"
	"message-service/internal/infrastructure/cassandra"
	"message-service/pkg/errors"
)

// CassandraConversationRepository implémente ConversationRepository avec Cassandra
type CassandraConversationRepository struct {
	client *cassandra.Client
}

// NewCassandraConversationRepository crée une nouvelle instance du repository
func NewCassandraConversationRepository(client *cassandra.Client) ConversationRepository {
	return &CassandraConversationRepository{
		client: client,
	}
}

// Create crée une nouvelle conversation
func (r *CassandraConversationRepository) Create(ctx context.Context, conversation *domain.Conversation) (*domain.Conversation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	// Générer un UUID si non fourni
	if len(conversation.ID) == 0 {
		conversation.ID = gocql.TimeUUID()
	}

	now := time.Now()
	conversation.CreatedAt = now
	conversation.UpdatedAt = &now

	// Insérer la conversation
	query := session.Query(
		`INSERT INTO conversations 
		(id, type, name, description, avatar_url, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		conversation.ID, conversation.Type, conversation.Name, conversation.Description,
		conversation.AvatarURL, conversation.CreatedBy, conversation.CreatedAt, conversation.UpdatedAt,
	).WithContext(ctx)

	if err := query.Exec(); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Ajouter le créateur comme premier membre (owner)
	ownerMember := &domain.ConversationMember{
		ID:             gocql.TimeUUID(),
		ConversationID: conversation.ID,
		UserID:         conversation.CreatedBy,
		Role:           "owner",
		IsMuted:        false,
		JoinedAt:       now,
	}

	if err := r.AddMember(ctx, ownerMember); err != nil {
		// Log l'erreur mais ne fail pas - la conversation est créée
		fmt.Printf("Warning: failed to add creator as member: %v\n", err)
	}

	// Ajouter à l'index dénormalisé pour les recherches rapides
	queryDenorm := session.Query(
		`INSERT INTO user_conversations 
		(user_id, conversation_id, conversation_type, last_activity)
		VALUES (?, ?, ?, ?)`,
		conversation.CreatedBy, conversation.ID, conversation.Type, now,
	).WithContext(ctx)

	if err := queryDenorm.Exec(); err != nil {
		fmt.Printf("Warning: failed to add to user_conversations index: %v\n", err)
	}

	return conversation, nil
}

// GetByID récupère une conversation par ID
func (r *CassandraConversationRepository) GetByID(ctx context.Context, id gocql.UUID) (*domain.Conversation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	var conversation domain.Conversation

	query := session.Query(
		`SELECT id, type, name, description, avatar_url, created_by, created_at, updated_at
		 FROM conversations WHERE id = ?`,
		id,
	).WithContext(ctx)

	if err := query.Scan(&conversation.ID, &conversation.Type, &conversation.Name, 
		&conversation.Description, &conversation.AvatarURL, &conversation.CreatedBy, 
		&conversation.CreatedAt, &conversation.UpdatedAt); err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.NewAppError(errors.ErrNotFound, "conversation not found", "")
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return &conversation, nil
}

// GetByUserID récupère les conversations d'un utilisateur
func (r *CassandraConversationRepository) GetByUserID(ctx context.Context, userID int32) ([]domain.Conversation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	var conversations []domain.Conversation

	query := session.Query(
		`SELECT conversation_id 
		 FROM user_conversations 
		 WHERE user_id = ?
		 ORDER BY last_activity DESC`,
		userID,
	).WithContext(ctx)

	iter := query.Iter()
	var convID gocql.UUID

	for iter.Scan(&convID) {
		if conv, err := r.GetByID(ctx, convID); err == nil && conv != nil {
			conversations = append(conversations, *conv)
		}
	}

	return conversations, iter.Close()
}

// Update met à jour une conversation
func (r *CassandraConversationRepository) Update(ctx context.Context, id gocql.UUID, conversation *domain.Conversation) (*domain.Conversation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	now := time.Now()
	conversation.UpdatedAt = &now

	query := session.Query(
		`UPDATE conversations 
		 SET name = ?, description = ?, avatar_url = ?, updated_at = ?
		 WHERE id = ?`,
		conversation.Name, conversation.Description, conversation.AvatarURL, now, id,
	).WithContext(ctx)

	if err := query.Exec(); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Delete supprime une conversation
func (r *CassandraConversationRepository) Delete(ctx context.Context, id gocql.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	session := r.client.GetSession()
	if session == nil {
		return errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	// Supprimer tous les messages de la conversation
	queryMessages := session.Query(
		`DELETE FROM messages WHERE conversation_id = ?`,
		id,
	).WithContext(ctx)

	if err := queryMessages.Exec(); err != nil {
		fmt.Printf("Warning: failed to delete messages: %v\n", err)
	}

	// Supprimer tous les membres
	queryMembers := session.Query(
		`DELETE FROM conversation_members WHERE conversation_id = ?`,
		id,
	).WithContext(ctx)

	if err := queryMembers.Exec(); err != nil {
		fmt.Printf("Warning: failed to delete members: %v\n", err)
	}

	// Supprimer la conversation
	query := session.Query(
		`DELETE FROM conversations WHERE id = ?`,
		id,
	).WithContext(ctx)

	return query.Exec()
}

// AddMember ajoute un membre à une conversation
func (r *CassandraConversationRepository) AddMember(ctx context.Context, member *domain.ConversationMember) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	session := r.client.GetSession()
	if session == nil {
		return errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	if len(member.ID) == 0 {
		member.ID = gocql.TimeUUID()
	}

	query := session.Query(
		`INSERT INTO conversation_members 
		(id, conversation_id, user_id, role, is_muted, joined_at, left_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		member.ID, member.ConversationID, member.UserID, member.Role, 
		member.IsMuted, member.JoinedAt, member.LeftAt,
	).WithContext(ctx)

	if err := query.Exec(); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Ajouter à l'index utilisateur si c'est un ajout
	if member.LeftAt == nil {
		queryIndex := session.Query(
			`INSERT INTO user_conversations 
			(user_id, conversation_id, conversation_type, last_activity)
			VALUES (?, ?, ?, ?)
			IF NOT EXISTS`,
			member.UserID, member.ConversationID, "group", time.Now(),
		).WithContext(ctx)

		if err := queryIndex.Exec(); err != nil {
			fmt.Printf("Warning: failed to update user index: %v\n", err)
		}
	}

	return nil
}

// RemoveMember retire un membre d'une conversation
func (r *CassandraConversationRepository) RemoveMember(ctx context.Context, conversationID gocql.UUID, userID int32) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	session := r.client.GetSession()
	if session == nil {
		return errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	now := time.Now()

	// Marquer le membre comme quitté (soft delete)
	query := session.Query(
		`UPDATE conversation_members 
		 SET left_at = ?
		 WHERE conversation_id = ? AND user_id = ?`,
		now, conversationID, userID,
	).WithContext(ctx)

	return query.Exec()
}

// GetMembers récupère les membres actifs d'une conversation
func (r *CassandraConversationRepository) GetMembers(ctx context.Context, conversationID gocql.UUID) ([]domain.ConversationMember, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	var members []domain.ConversationMember

	query := session.Query(
		`SELECT id, conversation_id, user_id, role, is_muted, joined_at, left_at
		 FROM conversation_members 
		 WHERE conversation_id = ?`,
		conversationID,
	).WithContext(ctx)

	iter := query.Iter()
	var id gocql.UUID
	var convID gocql.UUID
	var userID int32
	var role string
	var isMuted bool
	var joinedAt time.Time
	var leftAt *time.Time

	for iter.Scan(&id, &convID, &userID, &role, &isMuted, &joinedAt, &leftAt) {
		// Inclure seulement les membres actifs (left_at = null)
		if leftAt == nil {
			members = append(members, domain.ConversationMember{
				ID:             id,
				ConversationID: convID,
				UserID:         userID,
				Role:           role,
				IsMuted:        isMuted,
				JoinedAt:       joinedAt,
				LeftAt:         leftAt,
			})
		}
	}

	return members, iter.Close()
}
