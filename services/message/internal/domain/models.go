package domain

import (
	"time"

	"github.com/gocql/gocql"
)

// ============== CONVERSATION MODELS ==============

// Conversation représente une conversation (privée ou groupe)
type Conversation struct {
	ID          gocql.UUID `json:"id"`
	Type        string     `json:"type"` // 'private' ou 'group'
	Name        string     `json:"name"`
	Description string     `json:"description"`
	AvatarURL   string     `json:"avatar_url"`
	CreatedBy   string     `json:"created_by"`  // UUID string
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// ConversationMember représente un membre dans une conversation
type ConversationMember struct {
	ID             gocql.UUID `json:"id"`
	ConversationID gocql.UUID `json:"conversation_id"`
	UserID         string     `json:"user_id"`  // UUID string
	Role           string     `json:"role"` // 'owner', 'admin', 'member'
	IsMuted        bool       `json:"is_muted"`
	JoinedAt       time.Time  `json:"joined_at"`
	LeftAt         *time.Time `json:"left_at,omitempty"`
}

// CreateConversationRequest pour créer une conversation
type CreateConversationRequest struct {
	Type        string   `json:"type" binding:"omitempty,oneof=private group"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	AvatarURL   string   `json:"avatar_url"`
	MemberIDs   []string `json:"member_ids"` // UUIDs of members to add
}

// UpdateConversationRequest pour modifier une conversation
type UpdateConversationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
}

// ============== MESSAGE MODELS ==============

// Message représente un message dans une conversation
type Message struct {
	ID             gocql.UUID              `json:"id"`
	ConversationID gocql.UUID              `json:"conversation_id"`
	SenderID       string                  `json:"sender_id"`  // UUID string
	Content        string                  `json:"content"`
	Type           string                  `json:"type"` // 'text', 'image', 'video', etc.
	ReplyToID      *gocql.UUID             `json:"reply_to_id,omitempty"`
	IsForwarded    bool                    `json:"is_forwarded"`
	IsEdited       bool                    `json:"is_edited"`
	IsDeleted      bool                    `json:"is_deleted"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      *time.Time              `json:"updated_at,omitempty"`
	Attachments    []MessageAttachment     `json:"attachments,omitempty"`
	Status         map[string]MessageStatus `json:"status,omitempty"` // user_id(UUID) -> status
}

// CreateMessageRequest pour créer un message
type CreateMessageRequest struct {
	ConversationID gocql.UUID                 `json:"conversation_id" binding:"required"`
	Content        string                     `json:"content" binding:"required"`
	Type           string                     `json:"type"`
	ReplyToID      *gocql.UUID                `json:"reply_to_id"`
	Attachments    []MessageAttachmentInput   `json:"attachments"`
}

// UpdateMessageRequest pour modifier un message
type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required"`
	Type    string `json:"type"`
}

// MessageAttachment représente une pièce jointe
type MessageAttachment struct {
	ID           gocql.UUID `json:"id"`
	MessageID    gocql.UUID `json:"message_id"`
	FileURL      string     `json:"file_url"`
	FileName     string     `json:"file_name"`
	FileType     string     `json:"file_type"`
	FileSize     int        `json:"file_size"`
	ThumbnailURL string     `json:"thumbnail_url,omitempty"`
}

// MessageAttachmentInput pour créer une pièce jointe
type MessageAttachmentInput struct {
	FileURL      string `json:"file_url" binding:"required"`
	FileName     string `json:"file_name"`
	FileType     string `json:"file_type"`
	FileSize     int    `json:"file_size"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// MessageStatus représente le statut d'un message pour un utilisateur
type MessageStatus struct {
	ID          gocql.UUID `json:"id"`
	MessageID   gocql.UUID `json:"message_id"`
	UserID      string     `json:"user_id"`  // UUID string
	Status      string     `json:"status"` // 'sent', 'delivered', 'read'
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
}

// UpdateMessageStatusRequest pour mettre à jour le statut
type UpdateMessageStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=sent delivered read"`
}

// ============== RESPONSE MODELS ==============

// ConversationResponse pour la réponse API
type ConversationResponse struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	AvatarURL   string     `json:"avatar_url"`
	CreatedBy   int32      `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	MemberCount int        `json:"member_count,omitempty"`
}

// MessageResponse pour la réponse API
type MessageResponse struct {
	ID             string                 `json:"id"`
	ConversationID string                 `json:"conversation_id"`
	SenderID       int32                  `json:"sender_id"`
	Content        string                 `json:"content"`
	Type           string                 `json:"type"`
	ReplyToID      *string                `json:"reply_to_id,omitempty"`
	IsForwarded    bool                   `json:"is_forwarded"`
	IsEdited       bool                   `json:"is_edited"`
	IsDeleted      bool                   `json:"is_deleted"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      *time.Time             `json:"updated_at,omitempty"`
	Attachments    []MessageAttachment    `json:"attachments,omitempty"`
	Status         map[int32]interface{}  `json:"status,omitempty"`
}

// ConversationsResponse pour la réponse avec plusieurs conversations
type ConversationsResponse struct {
	Message       string                   `json:"message"`
	Conversations []ConversationResponse   `json:"conversations"`
	Total         int                      `json:"total"`
}

// MessagesResponse pour la réponse avec plusieurs messages
type MessagesResponse struct {
	Message  string            `json:"message"`
	Messages []MessageResponse `json:"messages"`
	Total    int               `json:"total"`
}
