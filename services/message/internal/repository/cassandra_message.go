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

// CassandraMessageRepository implémente MessageRepository avec Cassandra
type CassandraMessageRepository struct {
	client *cassandra.Client
}

// NewCassandraMessageRepository crée une nouvelle instance du repository
func NewCassandraMessageRepository(client *cassandra.Client) MessageRepository {
	return &CassandraMessageRepository{
		client: client,
	}
}

// Create crée un nouveau message
func (r *CassandraMessageRepository) Create(ctx context.Context, msg *domain.Message) (*domain.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	// Générer un UUID si non fourni
	if len(msg.ID) == 0 {
		msg.ID = gocql.TimeUUID()
	}

	now := time.Now()
	msg.CreatedAt = now
	msg.UpdatedAt = &now

	// Insérer le message
	query := session.Query(
		`INSERT INTO messages 
		(id, conversation_id, sender_id, content, type, reply_to_id, is_forwarded, is_edited, is_deleted, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.ConversationID, msg.SenderID, msg.Content, msg.Type, msg.ReplyToID, 
		msg.IsForwarded, msg.IsEdited, msg.IsDeleted, msg.CreatedAt, msg.UpdatedAt,
	).WithContext(ctx)

	if err := query.Exec(); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Insérer dans la table de dénormalisation pour la pagination
	queryDenorm := session.Query(
		`INSERT INTO conversation_messages 
		(conversation_id, created_at, id, sender_id, content, type)
		VALUES (?, ?, ?, ?, ?, ?)`,
		msg.ConversationID, msg.CreatedAt, msg.ID, msg.SenderID, msg.Content, msg.Type,
	).WithContext(ctx)

	if err := queryDenorm.Exec(); err != nil {
		// Log l'erreur mais ne fail pas - la table principale est prioritaire
		fmt.Printf("Warning: failed to insert into denormalized table: %v\n", err)
	}

	// Créer les statuts de livraison initiaux si des IDs utilisateurs fournis
	// (peut être fait asynchrone via événement NATS)

	return msg, nil
}

// GetByID récupère un message par ID
func (r *CassandraMessageRepository) GetByID(ctx context.Context, id gocql.UUID) (*domain.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	var msg domain.Message

	query := session.Query(
		`SELECT id, conversation_id, sender_id, content, type, reply_to_id, 
				is_forwarded, is_edited, is_deleted, created_at, updated_at
		 FROM messages WHERE id = ?`,
		id,
	).WithContext(ctx)

	if err := query.Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, 
		&msg.Type, &msg.ReplyToID, &msg.IsForwarded, &msg.IsEdited, &msg.IsDeleted,
		&msg.CreatedAt, &msg.UpdatedAt); err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrMessageNotFound
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Récupérer les pièces jointes
	attachments, err := r.getAttachments(ctx, msg.ID)
	if err == nil && len(attachments) > 0 {
		msg.Attachments = attachments
	}

	return &msg, nil
}

// GetByConversationID récupère les messages d'une conversation
func (r *CassandraMessageRepository) GetByConversationID(ctx context.Context, conversationID gocql.UUID, limit int, pageState []byte) ([]domain.Message, []byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	var messages []domain.Message

	query := session.Query(
		`SELECT id, sender_id, content, type, created_at 
		 FROM conversation_messages 
		 WHERE conversation_id = ?
		 LIMIT ?`,
		conversationID, limit,
	).WithContext(ctx).PageSize(limit)

	if len(pageState) > 0 {
		query = query.PageState(pageState)
	}

	iter := query.Iter()
	pageState = iter.PageState()

	var id gocql.UUID
	var senderID int32
	var content string
	var msgType string
	var createdAt time.Time

	for iter.Scan(&id, &senderID, &content, &msgType, &createdAt) {
		msg := domain.Message{
			ID:             id,
			ConversationID: conversationID,
			SenderID:       senderID,
			Content:        content,
			Type:           msgType,
			CreatedAt:      createdAt,
		}
		messages = append(messages, msg)
	}

	if err := iter.Close(); err != nil {
		return nil, nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	return messages, pageState, nil
}

// GetByUserID récupère les messages envoyés par un utilisateur
func (r *CassandraMessageRepository) GetByUserID(ctx context.Context, userID int32, limit int) ([]domain.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var messages []domain.Message

	query := session.Query(
		`SELECT id, conversation_id, sender_id, content, type, is_deleted, created_at
		 FROM messages WHERE sender_id = ?
		 LIMIT ?`, // Note: Cassandra nécessite un index pour cette requête
		userID, limit,
	).WithContext(ctx)

	iter := query.Iter()

	var id gocql.UUID
	var conversationID gocql.UUID
	var senderID int32
	var content string
	var msgType string
	var isDeleted bool
	var createdAt time.Time

	for iter.Scan(&id, &conversationID, &senderID, &content, &msgType, &isDeleted, &createdAt) {
		if !isDeleted {
			messages = append(messages, domain.Message{
				ID:             id,
				ConversationID: conversationID,
				SenderID:       senderID,
				Content:        content,
				Type:           msgType,
				CreatedAt:      createdAt,
			})
		}
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get user messages: %w", err)
	}

	return messages, nil
}

// Update met à jour un message
func (r *CassandraMessageRepository) Update(ctx context.Context, id gocql.UUID, updates *domain.Message) (*domain.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	now := time.Now()
	updates.UpdatedAt = &now

	query := session.Query(
		`UPDATE messages 
		 SET content = ?, is_edited = true, updated_at = ?
		 WHERE id = ?`,
		updates.Content, now, id,
	).WithContext(ctx)

	if err := query.Exec(); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Récupérer le message mis à jour
	return r.GetByID(ctx, id)
}

// Delete marque un message comme supprimé (soft delete)
func (r *CassandraMessageRepository) Delete(ctx context.Context, id gocql.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	session := r.client.GetSession()
	if session == nil {
		return errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	now := time.Now()

	query := session.Query(
		`UPDATE messages 
		 SET is_deleted = true, updated_at = ?
		 WHERE id = ?`,
		now, id,
	).WithContext(ctx)

	return query.Exec()
}

// ============== HELPER METHODS ==============

// getAttachments récupère les pièces jointes d'un message
func (r *CassandraMessageRepository) getAttachments(ctx context.Context, messageID gocql.UUID) ([]domain.MessageAttachment, error) {
	session := r.client.GetSession()
	if session == nil {
		return nil, errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	var attachments []domain.MessageAttachment

	query := session.Query(
		`SELECT id, message_id, file_url, file_name, file_type, file_size, thumbnail_url
		 FROM message_attachments WHERE message_id = ?`,
		messageID,
	).WithContext(ctx)

	iter := query.Iter()
	var id gocql.UUID
	var msgID gocql.UUID
	var fileURL string
	var fileName string
	var fileType string
	var fileSize int
	var thumbnailURL string

	for iter.Scan(&id, &msgID, &fileURL, &fileName, &fileType, &fileSize, &thumbnailURL) {
		attachments = append(attachments, domain.MessageAttachment{
			ID:           id,
			MessageID:    msgID,
			FileURL:      fileURL,
			FileName:     fileName,
			FileType:     fileType,
			FileSize:     fileSize,
			ThumbnailURL: thumbnailURL,
		})
	}

	return attachments, iter.Close()
}

// CreateAttachment crée une pièce jointe
func (r *CassandraMessageRepository) CreateAttachment(ctx context.Context, attachment *domain.MessageAttachment) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	session := r.client.GetSession()
	if session == nil {
		return errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	if len(attachment.ID) == 0 {
		attachment.ID = gocql.TimeUUID()
	}

	query := session.Query(
		`INSERT INTO message_attachments 
		(id, message_id, file_url, file_name, file_type, file_size, thumbnail_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		attachment.ID, attachment.MessageID, attachment.FileURL, attachment.FileName,
		attachment.FileType, attachment.FileSize, attachment.ThumbnailURL,
	).WithContext(ctx)

	return query.Exec()
}

// DeleteAttachment supprime une pièce jointe
func (r *CassandraMessageRepository) DeleteAttachment(ctx context.Context, attachmentID gocql.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	session := r.client.GetSession()
	if session == nil {
		return errors.NewAppError(errors.ErrInternalServer, "cassandra connection lost", "")
	}

	query := session.Query(
		`DELETE FROM message_attachments WHERE id = ?`,
		attachmentID,
	).WithContext(ctx)

	return query.Exec()
}
