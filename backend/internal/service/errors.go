package service

import (
	"fmt"
	"log/slog"
	"os"
)

// Logger instance
var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// ServiceError represents a service layer error
type ServiceError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *ServiceError) Error() string {
	return e.Message
}

// Error types
const (
	ErrorTypeValidation = "validation_error"
	ErrorTypeNotFound   = "not_found_error"
	ErrorTypeConflict   = "conflict_error"
	ErrorTypeDatabase   = "database_error"
	ErrorTypeInternal   = "internal_error"
)

// NewValidationError creates a new validation error
func NewValidationError(message string) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeValidation,
		Message: message,
		Code:    400,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Code:    404,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(message string) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeConflict,
		Message: message,
		Code:    409,
	}
}

// NewDatabaseError creates a new database error
func NewDatabaseError(message string) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeDatabase,
		Message: message,
		Code:    500,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeInternal,
		Message: message,
		Code:    500,
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}