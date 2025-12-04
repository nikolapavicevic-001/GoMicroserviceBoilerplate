package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	commonv1 "github.com/yourorg/boilerplate/shared/proto/gen/common/v1"
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	"github.com/yourorg/boilerplate/shared/auth"
	gatewayAuth "github.com/yourorg/boilerplate/services/api-gateway/internal/auth"
)

// Handler handles all HTTP requests for the API Gateway
type Handler struct {
	userClient     pb.UserServiceClient
	jwtSecret      string
	jwtExpiry      time.Duration
	googleProvider *auth.OAuth2Provider
	webAppURL      *url.URL
	webAppProxy    *httputil.ReverseProxy
}

// NewHandler creates a new handler
func NewHandler(
	userClient pb.UserServiceClient,
	jwtSecret string,
	jwtExpiry time.Duration,
	webAppURL *url.URL,
) *Handler {
	proxy := httputil.NewSingleHostReverseProxy(webAppURL)
	return &Handler{
		userClient:  userClient,
		jwtSecret:   jwtSecret,
		jwtExpiry:   jwtExpiry,
		webAppURL:   webAppURL,
		webAppProxy: proxy,
	}
}

// =============================================================================
// Auth Endpoints
// =============================================================================

// Register handles both JSON API and HTML form registration
func (h *Handler) Register(c echo.Context) error {
	// Check if it's a form submission or JSON API request
	isForm := c.Request().Header.Get("Content-Type") == "application/x-www-form-urlencoded"
	
	var email, name, password, passwordConfirm string
	if isForm {
		email = c.FormValue("email")
		name = c.FormValue("name")
		password = c.FormValue("password")
		passwordConfirm = c.FormValue("password_confirm")
	} else {
		var req RegisterRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request body"})
		}
		if err := req.Validate(); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		}
		email = req.Email
		name = req.Name
		password = req.Password
	}

	// Validate form input
	if isForm {
		if email == "" || name == "" || password == "" {
			return c.Redirect(http.StatusSeeOther, "/signup?error=all fields are required")
		}
		if len(password) < 8 {
			return c.Redirect(http.StatusSeeOther, "/signup?error=password must be at least 8 characters")
		}
		if password != passwordConfirm {
			return c.Redirect(http.StatusSeeOther, "/signup?error=passwords do not match")
		}
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	resp, err := h.userClient.CreateUser(ctx, &pb.CreateUserRequest{
		Email:    email,
		Name:     name,
		Password: password,
	})
	if err != nil {
		if isForm {
			return c.Redirect(http.StatusSeeOther, "/signup?error=registration failed")
		}
		return handleGRPCError(c, err)
	}

	// Set JWT cookie
	if err := gatewayAuth.SetJWTCookie(c, resp.User.Id, resp.User.Email, h.jwtSecret, h.jwtExpiry); err != nil {
		if isForm {
			return c.Redirect(http.StatusSeeOther, "/signup?error=authentication failed")
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to generate token"})
	}

	// Handle form submission - redirect to dashboard
	if isForm {
		// For HTMX requests, return redirect header
		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().Header().Set("HX-Redirect", "/dashboard")
			return c.NoContent(http.StatusOK)
		}
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}

	// Handle JSON API request - return token
	jwtToken, err := auth.GenerateToken(resp.User.Id, resp.User.Email, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to generate token"})
	}

	return c.JSON(http.StatusCreated, LoginResponse{
		Token: jwtToken,
		User:  toUserResponse(resp.User),
	})
}

// ShowLoginPage proxies to web-app to serve the login HTML template
func (h *Handler) ShowLoginPage(c echo.Context) error {
	h.webAppProxy.ServeHTTP(c.Response(), c.Request())
	return nil
}

// ShowSignupPage proxies to web-app to serve the signup HTML template
func (h *Handler) ShowSignupPage(c echo.Context) error {
	h.webAppProxy.ServeHTTP(c.Response(), c.Request())
	return nil
}

// Logout clears the JWT cookie and redirects to login
func (h *Handler) Logout(c echo.Context) error {
	gatewayAuth.ClearJWTCookie(c)
	return c.Redirect(http.StatusSeeOther, "/login")
}

// Login handles both JSON API and HTML form login
func (h *Handler) Login(c echo.Context) error {
	// Check if it's a form submission or JSON API request
	isForm := c.Request().Header.Get("Content-Type") == "application/x-www-form-urlencoded"
	
	var email, password string
	if isForm {
		email = c.FormValue("email")
		password = c.FormValue("password")
	} else {
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request body"})
		}
		email = req.Email
		password = req.Password
	}

	if email == "" || password == "" {
		if isForm {
			return c.Redirect(http.StatusSeeOther, "/login?error=email and password are required")
		}
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "email and password are required"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	resp, err := h.userClient.ValidatePassword(ctx, &pb.ValidatePasswordRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		if isForm {
			return c.Redirect(http.StatusSeeOther, "/login?error=invalid credentials")
		}
		return handleGRPCError(c, err)
	}

	// Set JWT cookie
	if err := gatewayAuth.SetJWTCookie(c, resp.User.Id, resp.User.Email, h.jwtSecret, h.jwtExpiry); err != nil {
		if isForm {
			return c.Redirect(http.StatusSeeOther, "/login?error=authentication failed")
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to generate token"})
	}

	// Handle form submission - redirect to dashboard
	if isForm {
		// For HTMX requests, return redirect header
		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().Header().Set("HX-Redirect", "/dashboard")
			return c.NoContent(http.StatusOK)
		}
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}

	// Handle JSON API request - return token
	jwtToken, err := auth.GenerateToken(resp.User.Id, resp.User.Email, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to generate token"})
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token: jwtToken,
		User:  toUserResponse(resp.User),
	})
}

// GoogleLogin godoc
// @Summary Google OAuth2 login
// @Description Redirect to Google OAuth2 login
// @Tags auth
// @Success 302
// @Router /api/v1/auth/google [get]
func (h *Handler) GoogleLogin(c echo.Context) error {
	if h.googleProvider == nil {
		return c.JSON(http.StatusNotImplemented, ErrorResponse{Message: "Google OAuth2 not configured"})
	}

	state := generateState()
	url := h.googleProvider.GetAuthURL(state)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback godoc
// @Summary Google OAuth2 callback
// @Description Handle Google OAuth2 callback
// @Tags auth
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} LoginResponse
// @Router /api/v1/auth/google/callback [get]
func (h *Handler) GoogleCallback(c echo.Context) error {
	if h.googleProvider == nil {
		return c.JSON(http.StatusNotImplemented, ErrorResponse{Message: "Google OAuth2 not configured"})
	}

	code := c.QueryParam("code")

	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	token, err := h.googleProvider.ExchangeCode(ctx, code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "failed to exchange code"})
	}

	userInfo, err := h.googleProvider.GetUserInfo(ctx, token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to get user info"})
	}

	user, err := h.createOrGetUser(ctx, userInfo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to create/get user"})
	}

	jwtToken, err := auth.GenerateToken(user.Id, user.Email, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to generate token"})
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token: jwtToken,
		User:  toUserResponse(user),
	})
}

// =============================================================================
// User CRUD Endpoints
// =============================================================================

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user account (admin endpoint, no JWT returned)
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users [post]
func (h *Handler) CreateUser(c echo.Context) error {
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
func (h *Handler) GetUser(c echo.Context) error {
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
func (h *Handler) UpdateUser(c echo.Context) error {
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
func (h *Handler) DeleteUser(c echo.Context) error {
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
func (h *Handler) ListUsers(c echo.Context) error {
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

// =============================================================================
// Helper Functions
// =============================================================================

func (h *Handler) createOrGetUser(ctx context.Context, userInfo *auth.UserInfo) (*pb.User, error) {
	userResp, err := h.userClient.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{
		Email: userInfo.Email,
	})
	if err == nil {
		return userResp.User, nil
	}

	createResp, err := h.userClient.CreateUser(ctx, &pb.CreateUserRequest{
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Password: generateRandomPassword(),
	})
	if err != nil {
		return nil, err
	}

	return createResp.User, nil
}

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

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func generateRandomPassword() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

