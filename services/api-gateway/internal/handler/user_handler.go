package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	commonv1 "github.com/yourorg/boilerplate/shared/proto/gen/common/v1"
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	"github.com/yourorg/boilerplate/shared/auth"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userClient pb.UserServiceClient
	jwtSecret  string
	jwtExpiry  time.Duration
}

// NewUserHandler creates a new user handler
func NewUserHandler(userClient pb.UserServiceClient, jwtSecret string, jwtExpiry time.Duration) *UserHandler {
	return &UserHandler{
		userClient: userClient,
		jwtSecret:  jwtSecret,
		jwtExpiry:  jwtExpiry,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request body"})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	resp, err := h.userClient.CreateUser(ctx, &pb.CreateUserRequest{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		return handleGRPCError(c, err)
	}

	return c.JSON(http.StatusCreated, toUserResponse(resp.User))
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c echo.Context) error {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	resp, err := h.userClient.GetUser(ctx, &pb.GetUserRequest{Id: id})
	if err != nil {
		return handleGRPCError(c, err)
	}

	return c.JSON(http.StatusOK, toUserResponse(resp.User))
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body UpdateUserRequest true "User data"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c echo.Context) error {
	id := c.Param("id")

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request body"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	resp, err := h.userClient.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:        id,
		Name:      req.Name,
		AvatarUrl: req.AvatarURL,
	})
	if err != nil {
		return handleGRPCError(c, err)
	}

	return c.JSON(http.StatusOK, toUserResponse(resp.User))
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c echo.Context) error {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	_, err := h.userClient.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
	if err != nil {
		return handleGRPCError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// ListUsers godoc
// @Summary List users
// @Description List users with pagination
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param search query string false "Search query"
// @Success 200 {object} ListUsersResponse
// @Security BearerAuth
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	search := c.QueryParam("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	resp, err := h.userClient.ListUsers(ctx, &pb.ListUsersRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
		Search: search,
	})
	if err != nil {
		return handleGRPCError(c, err)
	}

	users := make([]UserResponse, len(resp.Users))
	for i, user := range resp.Users {
		users[i] = toUserResponse(user)
	}

	return c.JSON(http.StatusOK, ListUsersResponse{
		Users: users,
		Pagination: PaginationResponse{
			Page:       int(resp.Pagination.Page),
			PageSize:   int(resp.Pagination.PageSize),
			TotalCount: resp.Pagination.TotalCount,
			TotalPages: int(resp.Pagination.TotalPages),
		},
	})
}

// Login godoc
// @Summary User login
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request body"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	// Get user by email
	userResp, err := h.userClient.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{
		Email: req.Email,
	})
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "invalid credentials"})
	}

	// In production, verify password here
	// For now, we'll generate token

	// Generate JWT token
	token, err := auth.GenerateToken(userResp.User.Id, userResp.User.Email, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to generate token"})
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  toUserResponse(userResp.User),
	})
}

// Helper functions

func toUserResponse(user *pb.User) UserResponse {
	return UserResponse{
		ID:        user.Id,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarUrl,
		CreatedAt: user.CreatedAt.AsTime(),
		UpdatedAt: user.UpdatedAt.AsTime(),
	}
}
