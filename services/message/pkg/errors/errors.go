package errors

// AppError represents an application-specific error
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Error codes
const (
	ErrNotFound       = 404
	ErrInvalidInput   = 400
	ErrUnauthorized   = 401
	ErrForbidden      = 403
	ErrConflict       = 409
	ErrInternalServer = 500
)

// Common errors
var (
	ErrMessageNotFound = &AppError{Code: ErrNotFound, Message: "Message not found"}
	ErrInvalidPayload  = &AppError{Code: ErrInvalidInput, Message: "Invalid payload"}
	ErrInternal        = &AppError{Code: ErrInternalServer, Message: "Internal server error"}
)

// NewAppError creates a new app error
func NewAppError(code int, message, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}
