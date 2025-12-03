package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yourorg/boilerplate/services/web-app/internal/session"
	"github.com/yourorg/boilerplate/shared/auth"
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	sessionStore   *session.Store
	userClient     pb.UserServiceClient
	googleProvider *auth.OAuth2Provider
	githubProvider *auth.OAuth2Provider
}

func NewAuthHandler(
	sessionStore *session.Store,
	userClient pb.UserServiceClient,
	googleProvider *auth.OAuth2Provider,
	githubProvider *auth.OAuth2Provider,
) *AuthHandler {
	return &AuthHandler{
		sessionStore:   sessionStore,
		userClient:     userClient,
		googleProvider: googleProvider,
		githubProvider: githubProvider,
	}
}

// ShowLoginPage renders the login page
func (h *AuthHandler) ShowLoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"Title": "Login",
	})
}

// ShowSignupPage renders the signup page
func (h *AuthHandler) ShowSignupPage(c echo.Context) error {
	return c.Render(http.StatusOK, "signup.html", map[string]interface{}{
		"Title": "Sign Up",
	})
}

// Signup handles user registration
func (h *AuthHandler) Signup(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")

	// Validate input
	if name == "" || email == "" || password == "" {
		return c.Render(http.StatusBadRequest, "signup.html", map[string]interface{}{
			"Title": "Sign Up",
			"Error": "All fields are required",
		})
	}

	if len(password) < 8 {
		return c.Render(http.StatusBadRequest, "signup.html", map[string]interface{}{
			"Title": "Sign Up",
			"Error": "Password must be at least 8 characters",
		})
	}

	if password != confirmPassword {
		return c.Render(http.StatusBadRequest, "signup.html", map[string]interface{}{
			"Title": "Sign Up",
			"Error": "Passwords do not match",
		})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	// Create user
	_, err := h.userClient.CreateUser(ctx, &pb.CreateUserRequest{
		Email:    email,
		Name:     name,
		Password: password,
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return c.Render(http.StatusConflict, "signup.html", map[string]interface{}{
				"Title": "Sign Up",
				"Error": "An account with this email already exists",
			})
		}
		return c.Render(http.StatusInternalServerError, "signup.html", map[string]interface{}{
			"Title": "Sign Up",
			"Error": "Failed to create account. Please try again.",
		})
	}

	// For HTMX requests, return redirect header to login
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/login?registered=true")
		return c.NoContent(http.StatusOK)
	}

	return c.Redirect(http.StatusSeeOther, "/login?registered=true")
}

// Login handles standard email/password login
func (h *AuthHandler) Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Check if user just registered
	registered := c.QueryParam("registered") == "true"
	successMsg := ""
	if registered {
		successMsg = "Account created successfully! Please sign in."
	}

	if email == "" || password == "" {
		data := map[string]interface{}{
			"Title": "Login",
			"Error": "Email and password are required",
		}
		if successMsg != "" {
			data["Success"] = successMsg
		}
		return c.Render(http.StatusBadRequest, "login.html", data)
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	// Validate password using the new ValidatePassword RPC
	resp, err := h.userClient.ValidatePassword(ctx, &pb.ValidatePasswordRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return c.Render(http.StatusUnauthorized, "login.html", map[string]interface{}{
				"Title": "Login",
				"Error": "Invalid email or password",
			})
		}
		return c.Render(http.StatusInternalServerError, "login.html", map[string]interface{}{
			"Title": "Login",
			"Error": "An error occurred. Please try again.",
		})
	}

	user := resp.User

	// Set session
	if err := h.sessionStore.SetUser(c, user.Id, user.Email, user.Name); err != nil {
		return c.Render(http.StatusInternalServerError, "login.html", map[string]interface{}{
			"Title": "Login",
			"Error": "Failed to create session",
		})
	}

	// For HTMX requests, return redirect header
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/dashboard")
		return c.NoContent(http.StatusOK)
	}

	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	if err := h.sessionStore.Clear(c); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to logout")
	}

	return c.Redirect(http.StatusSeeOther, "/login")
}

// GoogleLogin redirects to Google OAuth2 login
func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	if h.googleProvider == nil {
		return c.String(http.StatusBadRequest, "Google OAuth2 not configured")
	}

	url := h.googleProvider.Config.AuthCodeURL("state", oauth2.AccessTypeOnline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles Google OAuth2 callback
func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	if h.googleProvider == nil {
		return c.String(http.StatusBadRequest, "Google OAuth2 not configured")
	}

	code := c.QueryParam("code")
	if code == "" {
		return c.String(http.StatusBadRequest, "No code provided")
	}

	// Exchange code for token
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	token, err := h.googleProvider.Config.Exchange(ctx, code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to exchange token")
	}

	// Get user info
	userInfo, err := h.googleProvider.GetUserInfo(ctx, token)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get user info")
	}

	// Get or create user
	user, err := h.getOrCreateUser(ctx, userInfo.Email, userInfo.Name, userInfo.Picture)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create user")
	}

	// Set session
	if err := h.sessionStore.SetUser(c, user.Id, user.Email, user.Name); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create session")
	}

	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// GitHubLogin redirects to GitHub OAuth2 login
func (h *AuthHandler) GitHubLogin(c echo.Context) error {
	if h.githubProvider == nil {
		return c.String(http.StatusBadRequest, "GitHub OAuth2 not configured")
	}

	url := h.githubProvider.Config.AuthCodeURL("state", oauth2.AccessTypeOnline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback handles GitHub OAuth2 callback
func (h *AuthHandler) GitHubCallback(c echo.Context) error {
	if h.githubProvider == nil {
		return c.String(http.StatusBadRequest, "GitHub OAuth2 not configured")
	}

	code := c.QueryParam("code")
	if code == "" {
		return c.String(http.StatusBadRequest, "No code provided")
	}

	// Exchange code for token
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	token, err := h.githubProvider.Config.Exchange(ctx, code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to exchange token")
	}

	// Get user info
	userInfo, err := h.githubProvider.GetUserInfo(ctx, token)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get user info")
	}

	// Get or create user
	user, err := h.getOrCreateUser(ctx, userInfo.Email, userInfo.Name, userInfo.Picture)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create user")
	}

	// Set session
	if err := h.sessionStore.SetUser(c, user.Id, user.Email, user.Name); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create session")
	}

	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// getOrCreateUser gets existing user or creates new one
func (h *AuthHandler) getOrCreateUser(ctx context.Context, email, name, avatarURL string) (*pb.User, error) {
	// Try to get existing user
	resp, err := h.userClient.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{
		Email: email,
	})
	if err == nil {
		return resp.User, nil
	}

	// If not found, create new user
	if status.Code(err) == codes.NotFound {
		createResp, err := h.userClient.CreateUser(ctx, &pb.CreateUserRequest{
			Email:    email,
			Name:     name,
			Password: "", // OAuth2 users don't need password
		})
		if err != nil {
			return nil, err
		}
		return createResp.User, nil
	}

	return nil, err
}
