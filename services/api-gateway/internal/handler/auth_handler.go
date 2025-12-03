package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	"github.com/yourorg/boilerplate/shared/auth"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userClient     pb.UserServiceClient
	jwtSecret      string
	jwtExpiry      time.Duration
	googleProvider *auth.OAuth2Provider
	githubProvider *auth.OAuth2Provider
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userClient pb.UserServiceClient,
	jwtSecret string,
	jwtExpiry time.Duration,
	googleProvider, githubProvider *auth.OAuth2Provider,
) *AuthHandler {
	return &AuthHandler{
		userClient:     userClient,
		jwtSecret:      jwtSecret,
		jwtExpiry:      jwtExpiry,
		googleProvider: googleProvider,
		githubProvider: githubProvider,
	}
}

// GoogleLogin godoc
// @Summary Google OAuth2 login
// @Description Redirect to Google OAuth2 login
// @Tags auth
// @Success 302
// @Router /auth/google [get]
func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	state := generateState()

	// In production, store state in session/cache for validation
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
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	// state := c.QueryParam("state")

	// In production, validate state parameter

	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	// Exchange code for token
	token, err := h.googleProvider.ExchangeCode(ctx, code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "failed to exchange code",
		})
	}

	// Get user info from Google
	userInfo, err := h.googleProvider.GetUserInfo(ctx, token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "failed to get user info",
		})
	}

	// Create or get user
	user, err := h.createOrGetUser(ctx, userInfo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "failed to create/get user",
		})
	}

	// Generate JWT token
	jwtToken, err := auth.GenerateToken(user.Id, user.Email, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "failed to generate token",
		})
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token: jwtToken,
		User:  toUserResponse(user),
	})
}

// GitHubLogin godoc
// @Summary GitHub OAuth2 login
// @Description Redirect to GitHub OAuth2 login
// @Tags auth
// @Success 302
// @Router /auth/github [get]
func (h *AuthHandler) GitHubLogin(c echo.Context) error {
	state := generateState()

	url := h.githubProvider.GetAuthURL(state)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback godoc
// @Summary GitHub OAuth2 callback
// @Description Handle GitHub OAuth2 callback
// @Tags auth
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} LoginResponse
// @Router /auth/github/callback [get]
func (h *AuthHandler) GitHubCallback(c echo.Context) error {
	code := c.QueryParam("code")

	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	// Exchange code for token
	token, err := h.githubProvider.ExchangeCode(ctx, code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "failed to exchange code",
		})
	}

	// Get user info from GitHub
	userInfo, err := h.githubProvider.GetUserInfo(ctx, token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "failed to get user info",
		})
	}

	// Create or get user
	user, err := h.createOrGetUser(ctx, userInfo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "failed to create/get user",
		})
	}

	// Generate JWT token
	jwtToken, err := auth.GenerateToken(user.Id, user.Email, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "failed to generate token",
		})
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token: jwtToken,
		User:  toUserResponse(user),
	})
}

// Helper functions

func (h *AuthHandler) createOrGetUser(ctx context.Context, userInfo *auth.UserInfo) (*pb.User, error) {
	// Try to get existing user
	userResp, err := h.userClient.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{
		Email: userInfo.Email,
	})
	if err == nil {
		return userResp.User, nil
	}

	// User doesn't exist, create new one
	createResp, err := h.userClient.CreateUser(ctx, &pb.CreateUserRequest{
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Password: generateRandomPassword(), // OAuth users don't use password
	})
	if err != nil {
		return nil, err
	}

	return createResp.User, nil
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
