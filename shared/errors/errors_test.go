package errors

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		want     string
	}{
		{
			name: "error with wrapped error",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Message: "test message",
				Err:     errors.New("wrapped error"),
			},
			want: "TEST_ERROR: test message: wrapped error",
		},
		{
			name: "error without wrapped error",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Message: "test message",
			},
			want: "TEST_ERROR: test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appError.Error(); got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	appErr := &AppError{
		Code:    "TEST_ERROR",
		Message: "test message",
		Err:     wrappedErr,
	}

	if got := appErr.Unwrap(); got != wrappedErr {
		t.Errorf("AppError.Unwrap() = %v, want %v", got, wrappedErr)
	}
}

func TestNew(t *testing.T) {
	appErr := New("TEST_CODE", "test message")

	if appErr.Code != "TEST_CODE" {
		t.Errorf("New() Code = %v, want TEST_CODE", appErr.Code)
	}
	if appErr.Message != "test message" {
		t.Errorf("New() Message = %v, want test message", appErr.Message)
	}
	if appErr.Err != nil {
		t.Errorf("New() Err = %v, want nil", appErr.Err)
	}
}

func TestWrap(t *testing.T) {
	t.Run("wrap non-nil error", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrapped := Wrap(originalErr, "context message")

		if wrapped == nil {
			t.Error("Wrap() returned nil for non-nil error")
		}
		if !errors.Is(wrapped, originalErr) {
			t.Error("Wrap() should preserve the original error")
		}
	})

	t.Run("wrap nil error", func(t *testing.T) {
		wrapped := Wrap(nil, "context message")
		if wrapped != nil {
			t.Errorf("Wrap() = %v, want nil", wrapped)
		}
	})
}

func TestNewNotFoundError(t *testing.T) {
	appErr := NewNotFoundError("user", "123")

	if appErr.Code != "NOT_FOUND" {
		t.Errorf("NewNotFoundError() Code = %v, want NOT_FOUND", appErr.Code)
	}
	if appErr.Message != "user with id 123 not found" {
		t.Errorf("NewNotFoundError() Message = %v, want 'user with id 123 not found'", appErr.Message)
	}
	if !errors.Is(appErr, ErrNotFound) {
		t.Error("NewNotFoundError() should wrap ErrNotFound")
	}
}

func TestNewValidationError(t *testing.T) {
	appErr := NewValidationError("invalid email format")

	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("NewValidationError() Code = %v, want VALIDATION_ERROR", appErr.Code)
	}
	if appErr.Message != "invalid email format" {
		t.Errorf("NewValidationError() Message = %v, want 'invalid email format'", appErr.Message)
	}
	if !errors.Is(appErr, ErrInvalidInput) {
		t.Error("NewValidationError() should wrap ErrInvalidInput")
	}
}

func TestNewAlreadyExistsError(t *testing.T) {
	appErr := NewAlreadyExistsError("user", "email", "test@example.com")

	if appErr.Code != "ALREADY_EXISTS" {
		t.Errorf("NewAlreadyExistsError() Code = %v, want ALREADY_EXISTS", appErr.Code)
	}
	if appErr.Message != "user with email=test@example.com already exists" {
		t.Errorf("NewAlreadyExistsError() Message = %v", appErr.Message)
	}
	if !errors.Is(appErr, ErrAlreadyExists) {
		t.Error("NewAlreadyExistsError() should wrap ErrAlreadyExists")
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrNotFound directly",
			err:  ErrNotFound,
			want: true,
		},
		{
			name: "wrapped ErrNotFound",
			err:  NewNotFoundError("user", "123"),
			want: true,
		},
		{
			name: "different error",
			err:  ErrAlreadyExists,
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrAlreadyExists directly",
			err:  ErrAlreadyExists,
			want: true,
		},
		{
			name: "wrapped ErrAlreadyExists",
			err:  NewAlreadyExistsError("user", "email", "test@example.com"),
			want: true,
		},
		{
			name: "different error",
			err:  ErrNotFound,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAlreadyExists(tt.err); got != tt.want {
				t.Errorf("IsAlreadyExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrInvalidInput directly",
			err:  ErrInvalidInput,
			want: true,
		},
		{
			name: "wrapped ErrInvalidInput",
			err:  NewValidationError("invalid input"),
			want: true,
		},
		{
			name: "different error",
			err:  ErrNotFound,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidation(tt.err); got != tt.want {
				t.Errorf("IsValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	if !IsUnauthorized(ErrUnauthorized) {
		t.Error("IsUnauthorized(ErrUnauthorized) should return true")
	}
	if IsUnauthorized(ErrNotFound) {
		t.Error("IsUnauthorized(ErrNotFound) should return false")
	}
}

func TestIsForbidden(t *testing.T) {
	if !IsForbidden(ErrForbidden) {
		t.Error("IsForbidden(ErrForbidden) should return true")
	}
	if IsForbidden(ErrNotFound) {
		t.Error("IsForbidden(ErrNotFound) should return false")
	}
}

