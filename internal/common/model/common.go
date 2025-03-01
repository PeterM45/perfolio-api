package model

// Pagination is used for paginated responses
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// Response is a generic API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse is used for error responses
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
