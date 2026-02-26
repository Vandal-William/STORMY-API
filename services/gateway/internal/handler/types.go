package handler

// RegisterRequest represents the registration request payload
// @Description RegisterRequest structure for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3" example:"john_doe"`
	Password string `json:"password" binding:"required,min=8" example:"securepass123"`
	Phone    string `json:"phone" binding:"required,min=6" example:"1234567890"`
	Email    string `json:"email" binding:"email" example:"john@example.com"`
}

// LoginRequest represents the login request payload
// @Description LoginRequest structure for user authentication
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"securepass123"`
}

// ErrorResponse represents an error response
// @Description ErrorResponse structure for API errors
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid credentials"`
}

// ConversationRequest represents a conversation creation request
// @Description ConversationRequest structure for creating conversations
type ConversationRequest struct {
	Title   string   `json:"title" binding:"required" example:"Team Discussion"`
	Members []string `json:"members" binding:"required" example:"user1,user2"`
}

// MessageRequest represents a message creation request
// @Description MessageRequest structure for creating messages
type MessageRequest struct {
	ConversationID string `json:"conversationId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Content        string `json:"content" binding:"required" example:"Hello, world!"`
}
