package handler

import (
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/labstack/echo/v4"
	"net/http"
)

// Request types

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

func (r *CreateUserRequest) Validate() error {
	if r.Email == "" {
		return fmt.Errorf("email is required")
	}
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

type UpdateUserRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Response types

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListUsersResponse struct {
	Users      []UserResponse     `json:"users"`
	Pagination PaginationResponse `json:"pagination"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Helper function to convert gRPC errors to HTTP errors
func handleGRPCError(c echo.Context, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal server error",
		})
	}

	switch st.Code() {
	case codes.NotFound:
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Message: st.Message(),
			Code:    "NOT_FOUND",
		})
	case codes.AlreadyExists:
		return c.JSON(http.StatusConflict, ErrorResponse{
			Message: st.Message(),
			Code:    "ALREADY_EXISTS",
		})
	case codes.InvalidArgument:
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: st.Message(),
			Code:    "INVALID_ARGUMENT",
		})
	case codes.Unauthenticated:
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Message: st.Message(),
			Code:    "UNAUTHENTICATED",
		})
	case codes.PermissionDenied:
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Message: st.Message(),
			Code:    "PERMISSION_DENIED",
		})
	default:
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal server error",
			Code:    "INTERNAL",
		})
	}
}
