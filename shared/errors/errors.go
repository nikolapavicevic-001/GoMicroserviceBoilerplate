package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists is returned when a resource already exists
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when authorization fails
	ErrForbidden = errors.New("forbidden")

	// ErrInternal is returned for internal server errors
	ErrInternal = errors.New("internal server error")

	// ErrConflict is returned when there's a conflict
	ErrConflict = errors.New("conflict")
)

// AppError represents a custom application error
type AppError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements the errors.Unwrap interface
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource, id string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s with id %s not found", resource, id),
		Err:     ErrNotFound,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Err:     ErrInvalidInput,
	}
}

// NewAlreadyExistsError creates an already exists error
func NewAlreadyExistsError(resource, field, value string) *AppError {
	return &AppError{
		Code:    "ALREADY_EXISTS",
		Message: fmt.Sprintf("%s with %s=%s already exists", resource, field, value),
		Err:     ErrAlreadyExists,
	}
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if the error is an already exists error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if the error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}
