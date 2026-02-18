package dto

// ErrorResponse is the standard JSON error envelope.
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewErrorResponse constructs an ErrorResponse from a message string.
func NewErrorResponse(msg string) ErrorResponse {
	return ErrorResponse{Error: msg}
}
