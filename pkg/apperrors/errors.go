package apperrors

import (
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

// Error types - using different names than the function constructors
const (
	ErrTypeBadRequest   ErrorType = "BAD_REQUEST"
	ErrTypeNotFound     ErrorType = "NOT_FOUND"
	ErrTypeUnauthorized ErrorType = "UNAUTHORIZED"
	ErrTypeForbidden    ErrorType = "FORBIDDEN"
	ErrTypeConflict     ErrorType = "CONFLICT"
	ErrTypeInternal     ErrorType = "INTERNAL"
)

// Error represents an application error
type Error struct {
	errorType ErrorType
	message   string
}

// Error returns the error message
func (e *Error) Error() string {
	return e.message
}

// Type returns the error type
func (e *Error) Type() ErrorType {
	return e.errorType
}

// Status returns the HTTP status code
func (e *Error) Status() int {
	switch e.errorType {
	case ErrTypeBadRequest:
		return http.StatusBadRequest
	case ErrTypeNotFound:
		return http.StatusNotFound
	case ErrTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrTypeForbidden:
		return http.StatusForbidden
	case ErrTypeConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// Error factory functions
func BadRequest(message string) *Error {
	return &Error{
		errorType: ErrTypeBadRequest,
		message:   message,
	}
}

func NotFound(message string) *Error {
	return &Error{
		errorType: ErrTypeNotFound,
		message:   message,
	}
}

func Unauthorized(message string) *Error {
	return &Error{
		errorType: ErrTypeUnauthorized,
		message:   message,
	}
}

func Forbidden(message string) *Error {
	return &Error{
		errorType: ErrTypeForbidden,
		message:   message,
	}
}

func Conflict(message string) *Error {
	return &Error{
		errorType: ErrTypeConflict,
		message:   message,
	}
}

func InternalError(message string) *Error {
	return &Error{
		errorType: ErrTypeInternal,
		message:   message,
	}
}

// Wrap wraps an error with a new error type
func Wrap(err error, errorType ErrorType, message string) *Error {
	if message == "" {
		message = err.Error()
	} else {
		message = fmt.Sprintf("%s: %s", message, err.Error())
	}
	return &Error{
		errorType: errorType,
		message:   message,
	}
}
