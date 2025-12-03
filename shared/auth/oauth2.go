package auth

import (
	"context"
	"encoding/json"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// OAuth2Provider represents an OAuth2 provider
type OAuth2Provider struct {
	Config *oauth2.Config
	Name   string
}

// NewGoogleProvider creates a new Google OAuth2 provider
func NewGoogleProvider(clientID, clientSecret, redirectURL string) *OAuth2Provider {
	return &OAuth2Provider{
		Name: "google",
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// NewGitHubProvider creates a new GitHub OAuth2 provider
func NewGitHubProvider(clientID, clientSecret, redirectURL string) *OAuth2Provider {
	return &OAuth2Provider{
		Name: "github",
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		},
	}
}

// GetAuthURL generates the OAuth2 authorization URL
func (p *OAuth2Provider) GetAuthURL(state string) string {
	return p.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges the authorization code for a token
func (p *OAuth2Provider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.Config.Exchange(ctx, code)
}

// GetUserInfo retrieves user information from the OAuth2 provider
func (p *OAuth2Provider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.Config.Client(ctx, token)

	var userInfoURL string
	switch p.Name {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "github":
		userInfoURL = "https://api.github.com/user"
	default:
		userInfoURL = ""
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	// GitHub uses 'login' instead of 'name'
	if p.Name == "github" && userInfo.Name == "" {
		var githubUser struct {
			Login string `json:"login"`
			Email string `json:"email"`
		}
		json.Unmarshal(body, &githubUser)
		userInfo.Name = githubUser.Login
		if userInfo.Email == "" {
			userInfo.Email = githubUser.Email
		}
	}

	userInfo.Provider = p.Name
	return &userInfo, nil
}

// UserInfo represents user information from OAuth2 provider
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Provider string `json:"-"`
}
