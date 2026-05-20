package core

// ErrorResponse is the shape of every error body returned by the API.
type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// MessageResponse is the shape of simple success acknowledgements.
type MessageResponse struct {
	Message string `json:"message"`
}

// PageMeta carries pagination info included in list responses.
type PageMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}
