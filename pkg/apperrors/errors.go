package apperrors

import "fmt"

// ErrorType defines the type of error
type ErrorType string

const (
	// BadRequest indicates a client error
	BadRequest ErrorType = "BAD_REQUEST"
	// NotFound indicates a resource was not found
	NotFound ErrorType = "NOT_FOUND"
	// Unauthorized indicates authentication is required
	Unauthorized ErrorType = "UNAUTHORIZED"
	// Forbidden indicates the user doesn't have permission
	Forbidden ErrorType = "FORBIDDEN"
	// Conflict indicates a resource conflict
	Conflict ErrorType = "CONFLICT"
	// InternalError indicates a server error
	InternalError ErrorType = "INTERNAL_ERROR"
)

// Error is a custom application error
type Error struct {
	errType ErrorType
	message string
}

// Error returns the error message
func (e *Error) Error() string {
	return e.message
}

// Type returns the error type
func (e *Error) Type() ErrorType {
	return e.errType
}

// NewError creates a new Error
func NewError(errType ErrorType, message string) *Error {
	return &Error{
		errType: errType,
		message: message,
	}
}

// BadRequest creates a bad request error
func BadRequest(message string) *Error {
	return NewError(BadRequest, message)
}

// NotFound creates a not found error
func NotFound(resource, id string) *Error {
	return NewError(NotFound, fmt.Sprintf("%s with ID %s not found", resource, id))
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *Error {
	return NewError(Unauthorized, message)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *Error {
	return NewError(Forbidden, message)
}

// Conflict creates a conflict error
func Conflict(message string) *Error {
	return NewError(Conflict, message)
}

// Internal creates an internal error
func Internal(message string) *Error {
	return NewError(InternalError, message)
}
