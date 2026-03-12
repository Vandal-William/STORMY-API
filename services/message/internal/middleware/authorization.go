package middleware

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gocql/gocql"
    "message-service/internal/repository"
)

// AuthorizationMiddleware fournit les helpers pour les vérifications d'autorisation
type AuthorizationMiddleware struct {
    conversationRepo repository.ConversationRepository
    messageRepo      repository.MessageRepository
}

// NewAuthorizationMiddleware crée une nouvelle instance
func NewAuthorizationMiddleware(
    conversationRepo repository.ConversationRepository,
    messageRepo repository.MessageRepository,
) *AuthorizationMiddleware {
    return &AuthorizationMiddleware{
        conversationRepo: conversationRepo,
        messageRepo:      messageRepo,
    }
}

// IsConversationMember vérifie si un utilisateur est membre d'une conversation
func (a *AuthorizationMiddleware) IsConversationMember(ctx *gin.Context, conversationID gocql.UUID, userID int32) (bool, error) {
    members, err := a.conversationRepo.GetMembers(ctx.Request.Context(), conversationID)
    if err != nil {
        return false, err
    }

    for _, member := range members {
        if member.UserID == userID {
            return true, nil
        }
    }

    return false, nil
}

// IsConversationOwnerOrAdmin vérifie si un utilisateur est propriétaire ou admin d'une conversation
func (a *AuthorizationMiddleware) IsConversationOwnerOrAdmin(ctx *gin.Context, conversationID gocql.UUID, userID int32) (bool, error) {
    members, err := a.conversationRepo.GetMembers(ctx.Request.Context(), conversationID)
    if err != nil {
        return false, err
    }

    for _, member := range members {
        if member.UserID == userID {
            // Supposer que Role ou IsOwner existe
            return true, nil
        }
    }

    return false, nil
}

// IsMessageOwner vérifie si un utilisateur est propriétaire d'un message
func (a *AuthorizationMiddleware) IsMessageOwner(ctx *gin.Context, messageID gocql.UUID, userID int32) (bool, error) {
    message, err := a.messageRepo.GetByID(ctx.Request.Context(), messageID)
    if err != nil {
        return false, err
    }

    if message == nil {
        return false, fmt.Errorf("message not found")
    }

    return message.SenderID == userID, nil
}

// IsUserAccessingOwnData vérifie si un utilisateur accède à ses propres données
func IsUserAccessingOwnData(c *gin.Context, requestedUserID int32) (bool, error) {
    userID, err := GetUserIDFromContext(c)
    if err != nil {
        return false, err
    }

    return userID == requestedUserID, nil
}

// ForbiddenError retourne une réponse 403
func ForbiddenError(c *gin.Context, message string) {
    c.JSON(http.StatusForbidden, gin.H{"error": message})
}

// UnauthorizedError retourne une réponse 401
func UnauthorizedError(c *gin.Context, message string) {
    c.JSON(http.StatusUnauthorized, gin.H{"error": message})
}

// BadRequestError retourne une réponse 400
func BadRequestError(c *gin.Context, message string) {
    c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

// InternalServerError retourne une réponse 500
func InternalServerError(c *gin.Context, message string) {
    c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}