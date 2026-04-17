package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// GitHubConfig holds GitHub OAuth app credentials
type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string // typically http://localhost:PORT/callback
}

// GitHubUser represents the GitHub user info
type GitHubUser struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GitHubAuthManager handles GitHub OAuth flow
type GitHubAuthManager struct {
	config GitHubConfig
}

// NewGitHubAuthManager creates a new GitHub auth manager
func NewGitHubAuthManager(config GitHubConfig) *GitHubAuthManager {
	return &GitHubAuthManager{config: config}
}

// GetAuthorizationURL returns the GitHub OAuth authorization URL
// state: CSRF protection token (you should generate a random one)
func (gam *GitHubAuthManager) GetAuthorizationURL(state string) string {
	params := url.Values{}
	params.Set("client_id", gam.config.ClientID)
	params.Set("redirect_uri", gam.config.RedirectURI)
	params.Set("scope", "user:email")
	params.Set("state", state)
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

// ExchangeCodeForToken exchanges OAuth code for access token
func (gam *GitHubAuthManager) ExchangeCodeForToken(code string) (string, error) {
	params := url.Values{}
	params.Set("client_id", gam.config.ClientID)
	params.Set("client_secret", gam.config.ClientSecret)
	params.Set("code", code)
	params.Set("redirect_uri", gam.config.RedirectURI)

	req, err := http.NewRequest(
		http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("exchange code: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("github error: %s", result.Error)
	}

	return result.AccessToken, nil
}

// GetUserInfo fetches the authenticated GitHub user's info
func (gam *GitHubAuthManager) GetUserInfo(accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned status %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &user, nil
}

// VerifyToken checks if an access token is still valid
func (gam *GitHubAuthManager) VerifyToken(accessToken string) (bool, error) {
	_, err := gam.GetUserInfo(accessToken)
	if err != nil {
		return false, nil // Token is invalid/expired, not an error condition
	}
	return true, nil
}
