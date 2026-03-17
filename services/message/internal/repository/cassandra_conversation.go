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

	fmt.Printf("[CONVERSATION-INSERT] ==========================================\n")
	fmt.Printf("[CONVERSATION-INSERT] CreatedBy: %s (type: %T)\n", conversation.CreatedBy, conversation.CreatedBy)
	fmt.Printf("[CONVERSATION-INSERT] ID: %s\n", conversation.ID)
	fmt.Printf("[CONVERSATION-INSERT] Name: %s\n", conversation.Name)
	fmt.Printf("[CONVERSATION-INSERT] ==========================================\n")

	// Insérer la conversation
	query := session.Query(
		`INSERT INTO conversations 
		(id, conversation_type, name, description, avatar_url, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		conversation.ID, conversation.Type, conversation.Name, conversation.Description,
		conversation.AvatarURL, conversation.CreatedBy, conversation.CreatedAt, conversation.UpdatedAt,
	).WithContext(ctx)

	if err := query.Exec(); err != nil {
		fmt.Printf("[CONVERSATION-INSERT] ❌ INSERT FAILED: %v\n", err)
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}
	fmt.Printf("[CONVERSATION-INSERT] ✓ INSERT SUCCESS\n")

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
		fmt.Printf("[CONVERSATION-INSERT] ❌ Failed to add creator as member: %v\n", err)
		return nil, fmt.Errorf("failed to add creator as conversation member: %w", err)
	}
	fmt.Printf("[CONVERSATION-INSERT] ✓ Creator added as member\n")

	// Retourner la conversation créée
	return conversation, nil
}

// GetByID récupère une conversation par son ID
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
		`SELECT id, conversation_type, name, description, avatar_url, created_by, created_at, updated_at
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
func (r *CassandraConversationRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Conversation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	fmt.Printf("[DEBUG] GetByUserID - Recherche conversations pour userID: %s\n", userID)

	// Convertir le string userID en UUID pour Cassandra
	userUUID, err := gocql.ParseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	var conversations []domain.Conversation

	query := session.Query(
		`SELECT conversation_id 
		 FROM user_conversations 
		 WHERE user_id = ?
		 ORDER BY last_activity DESC`,
		userUUID,
	).WithContext(ctx)

	iter := query.Iter()
	var convID gocql.UUID

	for iter.Scan(&convID) {
		fmt.Printf("[DEBUG] GetByUserID - Chargement conversation: %s\n", convID.String())
		if conv, err := r.GetByID(ctx, convID); err == nil && conv != nil {
			conversations = append(conversations, *conv)
		}
	}

	fmt.Printf("[DEBUG] GetByUserID - Total trouvé: %d conversations\n", len(conversations))

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

	// D'abord récupérer tous les membres pour nettoyer l'index user_conversations
	members, err := r.GetMembers(ctx, id)
	if err != nil {
		fmt.Printf("Warning: failed to get members for cleanup: %v\n", err)
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

	// Nettoyer l'index user_conversations pour chaque ancien membre
	for _, member := range members {
		queryUserIndex := session.Query(
			`DELETE FROM user_conversations WHERE user_id = ? AND conversation_id = ?`,
			member.UserID, id,
		).WithContext(ctx)

		if err := queryUserIndex.Exec(); err != nil {
			fmt.Printf("Warning: failed to delete user index for %s: %v\n", member.UserID, err)
		}
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
func (r *CassandraConversationRepository) RemoveMember(ctx context.Context, conversationID gocql.UUID, userID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	session := r.client.GetSession()
	if session == nil {
		return errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	// D'abord récupérer le ID du membre
	var memberID gocql.UUID
	query := session.Query(
		`SELECT id FROM conversation_members 
		 WHERE conversation_id = ? AND user_id = ?
		 ALLOW FILTERING`,
		conversationID, userID,
	).WithContext(ctx)

	if err := query.Scan(&memberID); err != nil {
		if err == gocql.ErrNotFound {
			return fmt.Errorf("member not found in conversation")
		}
		return fmt.Errorf("failed to find member: %w", err)
	}

	// Maintenant supprimer le membre avec son ID (clé primaire)
	deleteQuery := session.Query(
		`DELETE FROM conversation_members WHERE id = ?`,
		memberID,
	).WithContext(ctx)

	return deleteQuery.Exec()
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
	var userID string
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

// GetUserRoleInConversation récupère le rôle d'un utilisateur dans une conversation
func (r *CassandraConversationRepository) GetUserRoleInConversation(ctx context.Context, conversationID gocql.UUID, userID string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	session := r.client.GetSession()
	if session == nil {
		return "", errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	var role string

	query := session.Query(
		`SELECT role FROM conversation_members 
		 WHERE conversation_id = ? AND user_id = ?
		 ALLOW FILTERING`,
		conversationID, userID,
	).WithContext(ctx)

	if err := query.Scan(&role); err != nil {
		if err == gocql.ErrNotFound {
			return "", nil // User not found in conversation
		}
		return "", fmt.Errorf("failed to get user role: %w", err)
	}

	return role, nil
}
