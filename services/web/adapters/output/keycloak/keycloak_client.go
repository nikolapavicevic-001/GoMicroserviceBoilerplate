package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/microserviceboilerplate/web/domain/entity"
	"github.com/microserviceboilerplate/web/domain/port/output"
	"golang.org/x/oauth2"
)

// KeycloakClient implements the AuthRepository interface
type KeycloakClient struct {
	oauthConfig *oauth2.Config
	keycloakURL string
	realm       string
	clientID    string
	secret      string
	httpClient  *http.Client
}

// NewKeycloakClient creates a new Keycloak client
func NewKeycloakClient() output.AuthRepository {
	keycloakURL := getEnv("KEYCLOAK_URL", "http://keycloak:8080")
	realm := getEnv("KEYCLOAK_REALM", "microservice")
	clientID := getEnv("KEYCLOAK_CLIENT_ID", "web-client")
	clientSecret := getEnv("KEYCLOAK_CLIENT_SECRET", "web-client-secret")
	redirectURL := getEnv("REDIRECT_URL", "http://localhost:8080/auth/callback")

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth", keycloakURL, realm),
			TokenURL: fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", keycloakURL, realm),
		},
	}

	return &KeycloakClient{
		oauthConfig: oauthConfig,
		keycloakURL: keycloakURL,
		realm:       realm,
		clientID:    clientID,
		secret:      clientSecret,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type tokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Authenticate authenticates a user with email and password
func (c *KeycloakClient) Authenticate(ctx context.Context, email, password string) (*entity.User, error) {
	tr, err := c.passwordGrantToken(ctx, email, password)
	if err != nil {
		return nil, err
	}

	user, err := c.GetUserInfo(ctx, tr.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	user.RefreshToken = tr.RefreshToken
	return user, nil
}

func (c *KeycloakClient) passwordGrantToken(ctx context.Context, username, password string) (*tokenResponse, error) {
	// Prefer the configured token URL to avoid any drift between env vars and client state.
	tokenURL := c.oauthConfig.Endpoint.TokenURL
	if tokenURL == "" {
		tokenURL = fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.keycloakURL, c.realm)
	}

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.secret)
	data.Set("username", username)
	data.Set("password", password)
	data.Set("scope", "openid profile email")

	req, _ := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var er tokenErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&er)
		if er.Error != "" {
			if er.ErrorDescription != "" {
				return nil, fmt.Errorf("keycloak token error: %s: %s", er.Error, er.ErrorDescription)
			}
			return nil, fmt.Errorf("keycloak token error: %s", er.Error)
		}
		return nil, fmt.Errorf("keycloak token request failed with status: %d", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("token response missing access_token")
	}
	return &tr, nil
}

// CreateUser creates a new user in Keycloak
func (c *KeycloakClient) CreateUser(ctx context.Context, email, password, firstName, lastName string) error {
	adminToken, err := c.GetAdminToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admin token: %w", err)
	}

	userData := map[string]interface{}{
		"username":  email,
		"email":     email,
		"firstName": firstName,
		"lastName":  lastName,
		"enabled":   true,
		"credentials": []map[string]interface{}{
			{
				"type":      "password",
				"value":     password,
				"temporary": false,
			},
		},
	}

	userJSON, _ := json.Marshal(userData)
	req, _ := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/admin/realms/%s/users", c.keycloakURL, c.realm),
		bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("user creation failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetUserInfo retrieves user information using an access token
func (c *KeycloakClient) GetUserInfo(ctx context.Context, accessToken string) (*entity.User, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", c.keycloakURL, c.realm),
		nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	user := entity.NewUser(
		userInfo.Sub,
		userInfo.Email,
		userInfo.PreferredUsername,
		userInfo.Name,
	)
	return user, nil
}

// GetAdminToken retrieves an admin token for Keycloak operations
func (c *KeycloakClient) GetAdminToken(ctx context.Context) (string, error) {
	adminUser := getEnv("KEYCLOAK_ADMIN", "admin")
	adminPass := getEnv("KEYCLOAK_ADMIN_PASSWORD", "admin")
	clientID := "admin-cli"

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", clientID)
	data.Set("username", adminUser)
	data.Set("password", adminPass)

	req, _ := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", c.keycloakURL),
		strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get admin token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode admin token: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// UserInfo represents Keycloak user info response
type UserInfo struct {
	Sub               string `json:"sub"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
